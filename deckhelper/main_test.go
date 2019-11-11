package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseDecks(t *testing.T) {
	decksPattern := filepath.Join("testdecks", "*.txt")
	decks, err := filepath.Glob(decksPattern)
	if err != nil {
		t.Fatalf("could not get test decks: %v", err)
	}
	if len(decks) == 0 {
		t.Fatal("no test decks found")
	}
	for _, path := range decks {
		deckFile, err := os.Open(path)
		if err != nil {
			t.Errorf("failed to open deck %q: %v", path, err)
			continue
		}
		if _, err := parseDeck(deckFile); err != nil {
			t.Errorf("failed to parse deck: %v", err)
		}
		deckFile.Close()
	}

}

func TestParseIgnoredWords(t *testing.T) {
	deck := `
	Deck
	Commander
	Sideboard

	1 Llanowar Elves (WAR) 223
	Deck
	Commander
	2 Elf Scout (WAR) 224
	`

	r := strings.NewReader(deck)
	cards, err := parseDeck(r)
	if err != nil {
		t.Fatalf("failed to parse deck: %v", err)
	}
	if len(cards) > 2 {
		t.Fatalf("Wrong amount of cards. want 2, got %d", len(cards))
	}
}
