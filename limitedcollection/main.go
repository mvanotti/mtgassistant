// program limitedcollection helps users arrange limited tournaments by allowing them to easily export their booster results.
// This program works by parsing the "Magic The Gathering - Arena" to capture the user's Collection and Inventory.
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
	diffStart    = flag.Int("diff_start", 0, "Starting diff point")
	diffEnd      = flag.Int("diff_end", 1, "Last diff message")
)

func main() {
	flag.Parse()
	log.Println("Parsing MTGA Log...")
	f, err := os.Open(os.ExpandEnv(*mtgOutputLog))
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer f.Close()
	boosterData, err := collectionfinder.FindBoosters(f)
	if err != nil {
		log.Fatalf("failed to parse mtga logs: %v", err)
	}

	log.Println("Parsing MTG Data Files...")
	db, err := carddb.CreateLibrary(*mtgDataPath)
	if err != nil {
		log.Fatalf("createLibrary failed: %v", err)
	}

	for i, booster := range boosterData {
		fmt.Printf("Booster #%d\n", i)

		for _, id := range booster.CardIds {
			card := db.GetCardByID(id)
			fmt.Printf("%d %s (%s) %s\n", 1, card.Name, card.Set, card.CollectorNumber)
		}

		fmt.Printf("\nCommon Wildcards: %d\nUncommon Wildcards: %d\nRare Wildcards: %d\nMythic Wildcards: %d\n",
			booster.CommonWildcards, booster.UncommonWildcards, booster.RareWildcards, booster.MythicWildcards)
	}
}
