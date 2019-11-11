// Package carddb provides functions to interact with a Magic The Gathering Arena cards.
// To use this library, you need a mtg arena installation, or at least the card assets that
// are typically stored in the MTGA_Data\Downloads\Data folder inside the MTGA installation directory.
package carddb

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type cardDB struct {
	byName   map[string][]*Card
	texts    map[uint64]string
	cardList []Card
	byID     map[uint64]*Card
}

// CardJSON is the JSON representation of a card, as it appears in the MTGA Resource files.
type CardJSON struct {
	ID              uint64   `json:"grpid"`
	TitleID         uint64   `json:"titleId"`
	CollectorNumber string   `json:"CollectorNumber"`
	Set             string   `json:"set"`
	Rarity          uint64   `json:"rarity"`
	ColorIdentity   []uint64 `json:"colorIdentity"`
	CastingCost     string   `json:"castingcost"`
	Types           []uint64 `json:"types"`
	Subtypes        []uint64 `json:"subtypes"`
	Supertypes      []uint64 `json:"supertypes"`
	CardTypeTextID  uint64   `json:"cardTypeTextId"`
	SubtypeTextID   uint64   `json:"subtypeTextId"`
}

type textJSON struct {
	ID   uint64 `json:"id"`
	Text string `json:"text"`
}

type langJSON struct {
	LangKey string     `json:"langkey"`
	Keys    []textJSON `json:"keys"`
}

// Card represents an Magic The Gathering: Arena card.
// It comes with the name and with the JSON representation of the card.
type Card struct {
	Name string
	CardJSON
}

// CardDB lets you interact with a Magic The Gathering: Arena card database.
type CardDB interface {
	// Returns all the cards with the given name.
	GetCard(name string) []*Card

	// Returns the card with the given ID, nil if it doesn't exist.
	GetCardByID(id uint64) *Card

	// ForEach runs f over each card in the database
	ForEach(f func(Card))
}

func (db *cardDB) GetCard(name string) []*Card {
	return db.byName[name]
}

func (db *cardDB) GetCardByID(id uint64) *Card {
	return db.byID[id]
}

func (db *cardDB) ForEach(f func(Card)) {
	for _, c := range db.cardList {
		f(c)
	}
}

// NewLibrary creates a new database of magic cards. It needs to use the files that are used by MTG Arena
// that describes the cards in JSON format. They are commonly named data_cards_<hash> and data_loc_<hash>
// The language is the language of the card names that you would like (probably "EN").
func NewLibrary(cardsFile io.Reader, textsFile io.Reader, textsLang string) (CardDB, error) {
	cards, err := parseCardsFile(cardsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cards file: %v", err)
	}
	texts, err := parseTextsFile(textsFile, textsLang)
	if err != nil {
		return nil, fmt.Errorf("failed to parse texts file: %v", err)
	}

	// cards is a list with all the cards. Really we would like to index it by the card name.
	// However, we might have multiple cards with the same name for different expansions, so
	// we are going to have to get all the versions.
	cardList := []Card{}
	byName := make(map[string][]*Card)
	byID := make(map[uint64]*Card)
	for i, cardjson := range cards {
		name, ok := texts[cardjson.TitleID]
		if !ok {
			return nil, fmt.Errorf("Missing card text for card %d", cardjson.ID)
		}
		card := Card{name, cardjson}
		cardList = append(cardList, card)
		if _, ok := byName[name]; !ok {
			byName[name] = []*Card{}
		}
		byID[card.ID] = &cardList[i]
		byName[name] = append(byName[name], &cardList[i])
	}

	return &cardDB{
		byName:   byName,
		byID:     byID,
		cardList: cardList,
		texts:    texts,
	}, nil
}

func parseTextsFile(textsFile io.Reader, lang string) (map[uint64]string, error) {
	decoder := json.NewDecoder(textsFile)
	var texts []langJSON
	if err := decoder.Decode(&texts); err != nil {
		return nil, fmt.Errorf("Failed to decode JSON: %v", err)
	}

	m := make(map[uint64]string)
	for _, v := range texts {
		if v.LangKey != lang {
			log.Printf("Skipping language %q", v.LangKey)
			continue
		}
		for _, kv := range v.Keys {
			m[kv.ID] = kv.Text
		}
		break
	}
	return m, nil
}

func parseCardsFile(cardsFile io.Reader) ([]CardJSON, error) {
	decoder := json.NewDecoder(cardsFile)
	var cards []CardJSON
	if err := decoder.Decode(&cards); err != nil {
		return nil, fmt.Errorf("Failed to decode JSON: %v", err)
	}
	return cards, nil
}

// FindMTGAResourceFiles returns the paths for the resource files needed by carddb, given the Data
// path for MTG Arena. Typically the files are stored in the mtgDataPath folder, but their names have a hash
// at the end. This function just try to look in that folder for the correct files.
func findMTGAResourceFiles(mtgDataPath string) (cardsFilePath string, textsFilePath string, err error) {
	mtgDataTextsFileGlob := filepath.Join(mtgDataPath, "data_loc_"+"*"+".mtga")
	mtgDataCardsFileGlob := filepath.Join(mtgDataPath, "data_cards_"+"*"+".mtga")

	textsFiles, err := filepath.Glob(mtgDataTextsFileGlob)
	if err != nil {
		err = fmt.Errorf("Failed to look for texts file: %v", err)
		return
	}
	if len(textsFiles) != 1 {
		err = fmt.Errorf("More than one texts file found: %v", textsFiles)
		return
	}

	cardsFiles, err := filepath.Glob(mtgDataCardsFileGlob)
	if err != nil {
		err = fmt.Errorf("Failed to look for cards file: %v", err)
		return
	}
	if len(cardsFiles) != 1 {
		err = fmt.Errorf("More than one cards file found: %v", cardsFiles)
		return
	}

	cardsFilePath, textsFilePath = cardsFiles[0], textsFiles[0]
	return
}

// CreateLibrary is a helper function for NewLibrary, it takes the Data path inside the MTG installation
// directory, and tries to find the required resource files for creating the database.
func CreateLibrary(mtgDataPath string) (CardDB, error) {
	cardsPath, textsPath, err := findMTGAResourceFiles(mtgDataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find mtga resource files: %v", err)
	}
	cardsFile, err := os.Open(cardsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cards resource file: %v", err)
	}
	defer cardsFile.Close()
	textsFile, err := os.Open(textsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open texts resource file: %v", err)
	}
	defer textsFile.Close()

	db, err := NewLibrary(cardsFile, textsFile, "EN")
	return db, err
}
