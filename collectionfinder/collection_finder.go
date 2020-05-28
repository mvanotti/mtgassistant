// Package collectionfinder parses the Magic The Gathering: Logs, returning user decks.
package collectionfinder

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const playerCollectionMessage string = "[UnityCrossThreadLogger]<== PlayerInventory.GetPlayerCardsV3"
const playerInventoryMessage string = "[UnityCrossThreadLogger]<== PlayerInventory.GetPlayerInventory"
const playerInventoryUpdatedMessage string = "[UnityCrossThreadLogger]<== Inventory.Updated"

// PlayerInventory represents the inventory of a player.
type PlayerInventory struct {
	PlayerID   string `json:"playerId"`
	WcCommon   int    `json:"wcCommon"`
	WcUncommon int    `json:"wcUncommon"`
	WcRare     int    `json:"wcRare"`
	WcMythic   int    `json:"wcMythic"`
}

type arenaMessage struct {
	ID      int             `json:"id"`
	Payload json.RawMessage `json:"payload"`
}

type cardListMsg map[string]uint32

type inventoryUpdateJSON struct {
	Context string `json:"context"`
	Updates []updatesMsg
}

type updatesMsg struct {
	Delta           deltaMsg             `json:"delta"`
	AetherizedCards []aetherizedCardsMsg `json:"aetherizedCards"`
}

type deltaMsg struct {
	WcCommonDelta   int `json:"wcCommonDelta"`
	WcUncommonDelta int `json:"wcUncommonDelta"`
	WcRareDelta     int `json:"wcRareDelta"`
	WcMythicDelta   int `json:"wcMythicDelta"`
}

type aetherizedCardsMsg struct {
	GrpID uint64 `json:"grpId"`
}

// BoosterContents represent the contents from a booster pack.
type BoosterContents struct {
	CommonWildcards   int
	UncommonWildcards int
	RareWildcards     int
	MythicWildcards   int
	CardIds           []uint64
}

// FindBoosters returns the list of all opened boosters in the MTG Arena Logs.
func FindBoosters(mtgalogs io.Reader) ([]BoosterContents, error) {
	inventoryUpdates, err := findMessages(mtgalogs, playerInventoryUpdatedMessage)
	if err != nil {
		return nil, err
	}

	res := make([]BoosterContents, 0)
	for _, updateRaw := range inventoryUpdates {
		var update inventoryUpdateJSON
		json.Unmarshal([]byte(updateRaw), &update)
		if update.Context != "Booster.Open" {
			continue
		}

		var contents BoosterContents
		for _, u := range update.Updates {
			contents.CommonWildcards += u.Delta.WcCommonDelta
			contents.UncommonWildcards += u.Delta.WcUncommonDelta
			contents.RareWildcards += u.Delta.WcRareDelta
			contents.MythicWildcards += u.Delta.WcMythicDelta
			for _, c := range u.AetherizedCards {
				contents.CardIds = append(contents.CardIds, c.GrpID)
			}
		}
		res = append(res, contents)
	}
	return res, nil
}

// FindInventory returns a list of all the inventories that appear in the MTG Logs
func FindInventory(mtgalogs io.Reader) ([]PlayerInventory, error) {
	inventories, err := findMessages(mtgalogs, playerInventoryMessage)
	if err != nil {
		return nil, err
	}

	res := make([]PlayerInventory, len(inventories), len(inventories))

	for i, inventory := range inventories {
		json.Unmarshal([]byte(inventory), &res[i])
	}
	return res, nil
}

func findMessages(mtgalogs io.Reader, prefix string) ([]json.RawMessage, error) {
	reader := bufio.NewReader(mtgalogs)
	res := make([]json.RawMessage, 0)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read line %v", err)
		}
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		// This line might contain the entire json payload, or just one part.
		// Look for the first appearence of `{`
		// Example: `[UnityCrossThreadLogger]<== PlayerInventory.GetPlayerCardsV3 {"id":232,...`
		var jsonReader io.Reader
		ind := strings.Index(line, "{")
		if ind != -1 {
			// This line contains a `{` at position ind.
			// Create a substring with that and put into a reader.
			sr := strings.NewReader(line[ind:])
			jsonReader = io.MultiReader(sr, reader)
		} else {
			jsonReader = reader
		}

		decoder := json.NewDecoder(jsonReader)
		var msg arenaMessage
		if err := decoder.Decode(&msg); err != nil {
			return nil, fmt.Errorf("failed to decode arena message: %v", err)
		}
		res = append(res, msg.Payload)
	}

	return res, nil
}

// FindCollection returns the list of user collections from the MTGA Logs.
func FindCollection(mtgalogs io.Reader) ([]map[uint64]uint32, error) {
	collections, err := findMessages(mtgalogs, playerCollectionMessage)
	if err != nil {
		return nil, err
	}

	cardLists := make([]map[uint64]uint32, 0)

	for _, c := range collections {
		var playerCards cardListMsg
		json.Unmarshal([]byte(c), &playerCards)

		cards := make(map[uint64]uint32)
		for txtID, count := range playerCards {
			id, err := strconv.Atoi(txtID)
			if err != nil {
				return nil, fmt.Errorf("found non-numeric ID %q: %v", txtID, err)
			}
			cards[uint64(id)] = count
		}
		cardLists = append(cardLists, cards)
	}

	return cardLists, nil
}
