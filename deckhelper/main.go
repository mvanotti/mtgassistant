// program deckhelper takes a list of cards (a card deck in mtga format,
// and will tell you how many cards do you have to craft.
// It will try to get your collection for the MTG Arena logs, and create a database
// of cards using the MTG Arena resource files.
// The MTGA Format for cards is:
// [Number of Copies] [Card Name] ([Expansion]) [CollectorNumber]
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/mvanotti/mtgassistant/carddb"
	"github.com/mvanotti/mtgassistant/collectionfinder"
)

var stdExpansions = map[string]bool{
	"GRN": true,
	"RNA": true,
	"WAR": true,
	"M20": true,
	"ELD": true,
}

var basicLandNames = map[string]bool{
	"Swamp":    true,
	"Plains":   true,
	"Island":   true,
	"Forest":   true,
	"Mountain": true,
}

var cardRarity = map[uint64]string{
	0: "Token",
	1: "Basic Land",
	2: "Common",
	3: "Uncommon",
	4: "Rare",
	5: "Mythic Rare",
}

var (
	mtgOutputLog = flag.String("log_file", `${USERPROFILE}\AppData\LocalLow\Wizards Of The Coast\MTGA\output_log.txt`, "Filepath of the MTG Arena Output Log, typically stored in an MTG folder inside C:\\Users")
	deckPath     = flag.String("deck", "", "Path to the file containing your mtga deck.")
	mtgDataPath  = flag.String("mtg_data", `C:\Program Files (x86)\Wizards of the Coast\MTGA\MTGA_Data\Downloads\Data`, "Path to the Downloads\\Data folder inside the MTG Arena Install Directory")
)

type card struct {
	count int
	name  string
	expn  string
	cc    string
}

var ignoredLines = map[string]bool{
	"Deck":      true,
	"Sideboard": true,
	"Commander": true,
	"":          true,
}

var mtgaRegexp *regexp.Regexp = regexp.MustCompile(`^([1-9][0-9]*) (.*) \(([A-Z0-9]{3})\) ?(.*)?$`)

func isBasicLand(str string) bool {
	return basicLandNames[str]
}
func isExpansionInStandard(str string) bool {
	return stdExpansions[str]
}

func parseFile(path string) ([]card, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()
	cls := make([]card, 0)

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		ln := strings.TrimSpace(scanner.Text())
		if ignoredLines[ln] {
			continue
		}
		ls := mtgaRegexp.FindStringSubmatch(ln)
		if len(ls) != 5 { // [matchedline count name expn cc]
			return nil, fmt.Errorf("[%d] could not parse line %q", lineNum, ln)
		}

		ls = ls[1:]
		amount, err := strconv.Atoi(ls[0])
		if err != nil {
			return nil, fmt.Errorf("[%d] could not parse card amount: %v", lineNum, err)
		}
		if isBasicLand(ls[1]) {
			continue
		}
		if !isExpansionInStandard(ls[2]) {
			continue
		}
		c := card{
			count: amount,
			name:  ls[1],
			expn:  ls[2],
			cc:    ls[3],
		}
		cls = append(cls, c)
		lineNum++
	}

	return cls, nil
}

func deckDistance(collection map[uint64]uint32, db carddb.CardDB, deck []card) (map[uint64]uint32, error) {
	res := make(map[uint64]uint32)
	for _, c := range deck {
		candidates := make([]*carddb.Card, 0)
		cs := db.GetCard(c.name)
		for _, card := range cs {
			if !isExpansionInStandard(card.Set) {
				continue
			}
			candidates = append(candidates, card)
		}
		if len(candidates) < 1 {
			return nil, fmt.Errorf("Card %q not found in standard rotation", c.name)
		}

		count := uint32(c.count)
		for _, candidate := range candidates {
			if collection[candidate.ID] > count {
				count = 0
				break
			}
			count -= collection[candidate.ID]
		}
		if count > 0 {
			res[candidates[0].ID] = count
		}
	}
	return res, nil
}

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
	collection := cardLists[0]

	log.Println("Parsing MTG Data Files...")
	db, err := carddb.CreateLibrary(*mtgDataPath)
	if err != nil {
		log.Fatalf("createLibrary failed: %v", err)
	}

	deck, err := parseFile(*deckPath)
	if err != nil {
		log.Fatalf("failed to parse deck file: %v", err)
	}

	dist, err := deckDistance(collection, db, deck)
	if err != nil {
		log.Fatalf("failed to get deck distance: %v", err)
	}

	totalCount := uint32(0)
	byRarity := make(map[uint64]uint32)
	for id, count := range dist {
		card := db.GetCardByID(id)
		if card == nil {
			log.Fatalf("invalid card id %d", id)
		}
		byRarity[card.Rarity] += count
		fmt.Printf("%d %s (%s)\n", count, card.Name, cardRarity[card.Rarity])
		totalCount += count
	}
	fmt.Printf("Need to craft %d cards\n", totalCount)
	for rarity, count := range byRarity {
		fmt.Printf("%s: %d\n", cardRarity[rarity], count)
	}
}
