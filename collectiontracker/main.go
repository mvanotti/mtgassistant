// program collectiontracker tells you how many boosters of a given expansion you need to open to complete your rare/mythic rare collection.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mvanotti/mtgassistant/carddb"
	"github.com/mvanotti/mtgassistant/collectionfinder"
)

var (
	mtgOutputLog = flag.String("log_file", `${USERPROFILE}\AppData\LocalLow\Wizards Of The Coast\MTGA\output_log.txt`, "Filepath of the MTG Arena Output Log, typically stored in an MTG folder inside C:\\Users")
	mtgDataPath  = flag.String("mtg_data", `C:\Program Files (x86)\Wizards of the Coast\MTGA\MTGA_Data\Downloads\Data`, "Path to the Downloads\\Data folder inside the MTG Arena Install Directory")
	mtgSet       = flag.String("set", "THB", "Expansion codename")
)

func main() {
	flag.Parse()
	log.Println("Parsing MTGA Log...")
	f, err := os.Open(os.ExpandEnv(*mtgOutputLog))
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer f.Close()
	cardLists, err := collectionfinder.FindCollection(f)
	if err != nil {
		log.Fatalf("failed to parse mtga logs: %v", err)
	}
	if len(cardLists) < 1 {
		log.Fatal("no decks found in the mtg logs. make sure to enable logs in the Arena app.")
	}
	cardList := cardLists[len(cardLists)-1]

	log.Println("Parsing MTG Data Files...")
	db, err := carddb.CreateLibrary(*mtgDataPath)
	if err != nil {
		log.Fatalf("createLibrary failed: %v", err)
	}

	predicate := func(c carddb.Card) bool {
		if c.Set != *mtgSet {
			return false
		}
		if c.Rarity != carddb.MythicRarity && c.Rarity != carddb.RareRarity {
			return false
		}
		return true
	}

	rares := uint32(0)
	mythics := uint32(0)
	for _, card := range db.Filter(predicate) {
		if card.Rarity == carddb.MythicRarity {
			mythics++
		}
		if card.Rarity == carddb.RareRarity {
			rares++
		}
	}

	missingRares := rares * 4
	missingMythics := mythics * 4

	for id, count := range cardList {
		card := db.GetCardByID(id)
		if card.Set != *mtgSet {
			continue
		}
		if card.Rarity == carddb.MythicRarity {
			missingMythics -= count
		} else if card.Rarity == carddb.RareRarity {
			missingRares -= count
		}
	}

	fmt.Printf("Missing Rares: %d\n", missingRares)
	fmt.Printf("Missing Mythics: %d\n", missingMythics)
}
