// program deckhelper takes a list of cards (a card deck in mtga format,
// and will tell you how many cards do you have to craft.
// It will try to get your collection for the MTG Arena logs, and create a database
// of cards using the MTG Arena resource files.
// The MTGA Format for cards is:
// [Number of Copies] [Card Name] ([Expansion]) [CollectorNumber]
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/mvanotti/mtgassistant/carddb"
	"github.com/mvanotti/mtgassistant/collectionfinder"
)

var basicLandNames = map[string]bool{
	"Swamp":    true,
	"Plains":   true,
	"Island":   true,
	"Forest":   true,
	"Mountain": true,
}

type deckHelper struct {
	enabledExpansions map[string]bool   // expansions that are available.
	db                carddb.CardDB     // database of all magic cards in the arena.
	collection        map[uint64]uint32 // The user's card collection.
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
	mtgOutputLog  = flag.String("log_file", `${USERPROFILE}\AppData\LocalLow\Wizards Of The Coast\MTGA\output_log.txt`, "Filepath of the MTG Arena Output Log, typically stored in an MTG folder inside C:\\Users")
	deckPath      = flag.String("deck", "", "Path to the file containing your mtga deck.")
	mtgDataPath   = flag.String("mtg_data", `C:\Program Files (x86)\Wizards of the Coast\MTGA\MTGA_Data\Downloads\Data`, "Path to the Downloads\\Data folder inside the MTG Arena Install Directory")
	enabledSets   = flag.String("sets", "STD", "Comma separated list of enabled sets. The string `STD` refers to all standard sets, and `ALL` to all sets (historic).")
	fromClipboard = flag.Bool("clipboard", false, "If set to true, will read the deck from the clipboard instead of a file.")
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

func parseDeck(r io.Reader) ([]card, error) {
	cls := make([]card, 0)

	scanner := bufio.NewScanner(r)
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

func (helper deckHelper) isExpansionEnabled(set string) bool {
	return helper.enabledExpansions[set]
}

func (helper deckHelper) deckDistance(deck []card) (map[uint64]uint32, error) {
	res := make(map[uint64]uint32)
	for _, c := range deck {
		candidates := make([]*carddb.Card, 0)
		cs := helper.db.GetCard(c.name)
		for _, card := range cs {
			if !helper.isExpansionEnabled(card.Set) {
				continue
			}
			candidates = append(candidates, card)
		}

		if len(candidates) < 1 {
			return nil, fmt.Errorf("Card %q not found in enabled sets", c.name)
		}

		count := uint32(c.count)
		for _, candidate := range candidates {
			if helper.collection[candidate.ID] > count {
				count = 0
				break
			}
			count -= helper.collection[candidate.ID]
		}
		if count > 0 {
			res[candidates[0].ID] = count
		}
	}
	return res, nil
}

var allSets = []string{"RNA", "PLS", "9ED", "NPH", "C13", "MOR", "WWK", "M11", "AVR", "CHK", "WTH", "LRW", "M10", "XLN", "SCG", "8ED", "SOK", "DIS", "RTR", "GTC", "ORI", "BFZ", "EMN", "M19", "MH1", "10E", "ME4", "RIX", "WAR", "MIR", "RAV", "ROE", "DAR", "G18", "GRN", "M20", "ELD", "DST", "5DN", "ME2", "AKH", "ANA", "INV", "CMD", "ZEN"}
var stdSets = []string{"ELD", "M20", "WAR", "GRN", "RNA"}

func parseExpansions(enabledSets string) (map[string]bool, error) {
	var enabledExpansions = make(map[string]bool)
	for _, set := range allSets {
		enabledExpansions[set] = false
	}

	sets := strings.Split(enabledSets, ",")
	for _, set := range sets {
		if set == "ALL" {
			for _, set := range allSets {
				enabledExpansions[set] = true
			}
			break
		}
		if set == "STD" {
			for _, set := range stdSets {
				enabledExpansions[set] = true
			}
			continue
		}
		if _, ok := enabledExpansions[set]; !ok {
			return nil, fmt.Errorf("invalid set: %v", set)
		}
		enabledExpansions[set] = true
	}
	return enabledExpansions, nil
}

func newDeckHelper(mtgOutputLogPath string, mtgDataPath string, enabledSets string) (*deckHelper, error) {
	log.Println("Parsing MTGA Log...")
	f, err := os.Open(os.ExpandEnv(mtgOutputLogPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}
	defer f.Close()
	cardLists, err := collectionfinder.FindCollection(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mtga logs: %v", err)
	}
	if len(cardLists) < 1 {
		return nil, errors.New("no decks found in the mtg logs. make sure to enable logs in the Arena app")
	}
	collection := cardLists[len(cardLists)-1]
	log.Printf("Collection has %d cards", len(collection))

	log.Println("Parsing MTG Data Files...")
	db, err := carddb.CreateLibrary(mtgDataPath)
	if err != nil {
		return nil, fmt.Errorf("createLibrary failed: %v", err)
	}

	enabledExpansions, err := parseExpansions(enabledSets)
	if err != nil {
		return nil, fmt.Errorf("failed to parse enabled expansions list: %v", err)
	}

	return &deckHelper{enabledExpansions, db, collection}, nil
}

func main() {
	flag.Parse()
	helper, err := newDeckHelper(*mtgOutputLog, *mtgDataPath, *enabledSets)
	if err != nil {
		log.Fatalf("failed to create deck helper: %v", err)
	}

	var deckReader io.Reader
	if !*fromClipboard {
		deckFile, err := os.Open(*deckPath)
		if err != nil {
			log.Fatalf("failed to open deck file: %v", err)
		}
		defer deckFile.Close()
		deckReader = deckFile
	} else {
		clipboard, err := clipboard.ReadAll()
		if err != nil {
			log.Fatalf("failed to read clipboard: %v", err)
		}
		deckReader = strings.NewReader(clipboard)
	}

	deck, err := parseDeck(deckReader)
	if err != nil {
		log.Fatalf("failed to parse deck file: %v", err)
	}

	dist, err := helper.deckDistance(deck)
	if err != nil {
		log.Fatalf("failed to get deck distance: %v", err)
	}

	totalCount := uint32(0)
	byRarity := make(map[uint64]uint32)
	for id, count := range dist {
		card := helper.db.GetCardByID(id)
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
