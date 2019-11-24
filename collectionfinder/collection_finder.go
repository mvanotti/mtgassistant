// Package collectionfinder parses the Magic The Gathering: Logs, returning user decks.
package collectionfinder

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

const playerCollectionMessage string = "[UnityCrossThreadLogger]<== PlayerInventory.GetPlayerCardsV3"

type getPlayerCardsV3JSON struct {
	ID      int         `json:"id"`
	Payload cardListMsg `json:"payload"`
}
type cardListMsg map[string]uint32

// FindCollection returns the list of user collections from the MTGA Logs.
func FindCollection(mtgalogs io.Reader) ([]map[uint64]uint32, error) {
	reader := bufio.NewReader(mtgalogs)
	cardLists := make([]map[uint64]uint32, 0)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read line %v", err)
		}
		if !strings.HasPrefix(line, playerCollectionMessage) {
			continue
		}
		log.Println("Found Player Cards Collection Message... Parsing")
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
		var getPlayerCardsMsg getPlayerCardsV3JSON
		if err := decoder.Decode(&getPlayerCardsMsg); err != nil {
			return nil, fmt.Errorf("failed to decode CardListMsg: %v", err)
		}
		cardsJSON := getPlayerCardsMsg.Payload
		log.Printf("Found %d cards!", len(cardsJSON))
		cards := make(map[uint64]uint32)
		for txtID, count := range cardsJSON {
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
