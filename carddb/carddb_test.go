package carddb

import (
	"strings"
	"testing"
)

var testCardsJSON = `
[
  {
	"grpid": 0, "titleId": 0, "artId": 0, "isToken": false,
	"isCollectible": false, "isCraftable": false, "artSize": 0, "power": 8,
	"toughness": 4, "flavorId": 1, "CollectorNumber": "0", 
	"altDeckLimit": null, "cmc": 8, "rarity": 2, "artistCredit": "John Doe",
	"set": "MI", "linkedFaceType": 0, "types": [ 2 ], "subtypes": [ 55 ],
	"supertypes": [], "cardTypeTextId": 10, "subtypeTextId": 11,
    "colors": [5], "frameColors": [5], "frameDetails": [],
    "colorIdentity": [ 5 ],
    "abilities": [ { "abilityId": 7, "textId": 7 } ],
    "hiddenAbilities": [], "linkedFaces": [], "castingcost": "o6oGoG",
    "linkedTokens": [], "knownSupportedStyles": []
  },
  {
    "grpid": 1, "titleId": 1, "artId": 1, "isToken": false,
    "isCollectible": true, "isCraftable": false, "artSize": 0,
    "power": 0, "toughness": 0, "flavorId": 1, "CollectorNumber": "0",
    "altDeckLimit": 250, "cmc": 0, "rarity": 1, "artistCredit": "John Doe",
    "set": "MI", "linkedFaceType": 0, "types": [ 5 ], "subtypes": [ 29 ],
    "supertypes": [ 1 ], "cardTypeTextId": 8, "subtypeTextId": 12,
	"colors": [], "frameColors": [ 5 ], "frameDetails": [],
	"colorIdentity": [ 5 ], "abilities": [], "hiddenAbilities": [],
	"linkedFaces": [], "castingcost": "o0", "linkedTokens": [],
    "knownSupportedStyles": []
  },
  {
    "grpid": 2, "titleId": 2, "artId": 2, "isToken": false,
    "isCollectible": true, "isCraftable": true, "artSize": 0, "power": 0,
	"toughness": 0, "flavorId": 1, "CollectorNumber": "79",
	"altDeckLimit": null, "cmc": 6, "rarity": 4,
    "artistCredit": "John Doe", "set": "WAR", "linkedFaceType": 0,
    "types": [ 1 ], "subtypes": [], "supertypes": [2],
    "cardTypeTextId": 8, "subtypeTextId": 0,
    "colors": [ 3 ], "frameColors": [ 3 ], "frameDetails": [],
    "colorIdentity": [ 3 ],
    "abilities": [{ "abilityId": 4, "textId": 4 },
      { "abilityId": 5, "textId": 5 },
      { "abilityId": 6, "textId": 6 }],
    "hiddenAbilities": [],"linkedFaces": [], "castingcost": "o3oBoBoB",
    "linkedTokens": [], "knownSupportedStyles": [ "DA" ]
  }
]
`

var testTextsJSON = `
[
	{ "langkey": "EN", "keys" : [
		{"id": 1, "text": "TEST1"},
		{"id": 2, "text": "TEST2"},
		{"id": 3, "text": "TEST3"}
	]},
	{ "langkey": "ES", "keys" : [
		{"id": 1, "text": "PRUEBA1"},
		{"id": 2, "text": "PRUEBA2"},
		{"id": 3, "text": "PRUEBA3"}
	]}
]
`

func TestParseCardsFile(t *testing.T) {
	r := strings.NewReader(testCardsJSON)

	cards, err := parseCardsFile(r)
	if err != nil {
		t.Fatalf("failed to parse cards file: %v", err)
	}
	if len(cards) != 3 {
		t.Fatalf("bad amount of cards. Want %d, got %d", 3, len(cards))
	}
	for i := uint64(0); i < uint64(len(cards)); i++ {
		if cards[i].ID != i {
			t.Errorf("Wrong card ID for card %d. Want %d, got %d", i, i, cards[i].ID)
		}
		if cards[i].TitleID != i {
			t.Errorf("Wrong text ID for card %d. Want %d, got %d", i, i, cards[i].TitleID)
		}
	}
}

func TestParseTextsFile(t *testing.T) {
	r := strings.NewReader(testTextsJSON)

	wantEN := map[uint64]string{1: "TEST1", 2: "TEST2", 3: "TEST3"}
	texts, err := parseTextsFile(r, "EN")
	if err != nil {
		t.Fatalf("failed to parse texts file: %v", err)
	}
	if len(texts) != len(wantEN) {
		t.Fatalf("wrong number of texts. Want %d, got %d", len(wantEN), len(texts))
	}

	for k, v := range wantEN {
		if texts[k] != v {
			t.Errorf("wrong text for key %d. want %q, got %q", k, v, texts[k])
		}
	}
	r = strings.NewReader(testTextsJSON)
	wantES := map[uint64]string{1: "PRUEBA1", 2: "PRUEBA2", 3: "PRUEBA3"}
	texts, err = parseTextsFile(r, "ES")
	if err != nil {
		t.Fatalf("failed to parse texts file: %v", err)
	}
	if len(texts) != len(wantES) {
		t.Fatalf("wrong number of texts. Want %d, got %d", len(wantES), len(texts))
	}

	for k, v := range wantES {
		if texts[k] != v {
			t.Errorf("wrong text for key %d. want %q, got %q", k, v, texts[k])
		}
	}
}

func TestLibrary(t *testing.T) {
	var cardsJSON = `[
		{"grpid": 0, "titleId": 100, "CollectorNumber": "300", "set": "WAR", "rarity": 0, "castingcost": "o0", "types":[0], "cardTypeTextId":200},
		{"grpid": 1, "titleId": 101, "CollectorNumber": "301", "set": "ELD", "rarity": 0, "castingcost": "o1", "types":[1], "cardTypeTextId":201},
		{"grpid": 2, "titleId": 102, "CollectorNumber": "302", "set": "WAR", "rarity": 0, "castingcost": "o0g2", "types":[1], "cardTypeTextId":201}
	]
	`

	var textsJSON = `[
		{ "langkey": "EN", "keys" : [
			{"id": 100, "text": "Land of the Elves"},
			{"id": 101, "text": "Hydrosis Krosis"},
			{"id": 102, "text": "Chandra, Who burns the world"},
			{"id": 200, "text": "Land"},
			{"id": 201, "text": "PlainWalker"}
			]
		}
	]
	`

	cardsFile := strings.NewReader(cardsJSON)
	textsFile := strings.NewReader(textsJSON)

	db, err := NewLibrary(cardsFile, textsFile, "EN")
	if err != nil {
		t.Fatalf("failed to create library: %v", err)
	}

	ls := db.GetCard("Land of the Elves")
	if len(ls) != 1 {
		t.Fatalf("GetCard failed. wanted 1 card, got %d", len(ls))
	}
	if ls[0].Name != "Land of the Elves" {
		t.Errorf("Name mismatch. Want %q, got %q", "Land of the Elves", ls[0].Name)
	}
	if ls[0].ID != 0 {
		t.Errorf("ID mismatch. want 0, got %d", ls[0].ID)
	}

	c := db.GetCardByID(0)
	if c != ls[0] {
		t.Errorf("Card By Name and by ID mismatch")
	}
}
