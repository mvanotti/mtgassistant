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

const playerCollectionMessage string = "<== PlayerInventory.GetPlayerCardsV3"

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

		if strings.HasPrefix(line, playerCollectionMessage) {
			log.Println("Found Player Cards Collection Message... Parsing")
			decoder := json.NewDecoder(reader)
			var cardsJSON cardListMsg
			if err := decoder.Decode(&cardsJSON); err != nil {
				return nil, fmt.Errorf("failed to decode CardListMsg: %v", err)
			}
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
	}

	return cardLists, nil
}
