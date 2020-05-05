// program webserver is a web server that parses out MTG Arena Logs and outputs the contents of the cards boosters that you opened
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/mvanotti/mtgassistant/carddb"
	"github.com/mvanotti/mtgassistant/collectionfinder"
)

var (
	mtgDataPath = flag.String("mtg_data", `C:\Program Files (x86)\Wizards of the Coast\MTGA\MTGA_Data\Downloads\Data`, "Path to the Downloads\\Data folder inside the MTG Arena Install Directory")
	landingpage = flag.String("landing", "boostertracking.html", "Path to the landing page.")
	jsonFormat  = flag.Bool("json", true, "Whether or not to output booster info in JSON format.")
)

const maxMtgaLogsSize int64 = 100 << 20 // 20 MiB

// BoosterContents represent the contents of a MTG:Arena booster pack.
type BoosterContents struct {
	WcCommon   int      `json:"wcc"`
	WcUncommon int      `json:"wcu"`
	WcRare     int      `json:"wcr"`
	WcMythic   int      `json:"wcm"`
	Cards      []string `json:"cards"`
}

func outputJSON(w http.ResponseWriter, boosterData []collectionfinder.BoosterContents) {
	var boosters []BoosterContents

	w.Header().Add("Content-Type", "application/json")

	for _, booster := range boosterData {
		var contents BoosterContents
		contents.WcCommon = booster.CommonWildcards
		contents.WcUncommon = booster.UncommonWildcards
		contents.WcRare = booster.RareWildcards
		contents.WcMythic = booster.MythicWildcards
		for _, id := range booster.CardIds {
			card := db.GetCardByID(id)
			c := fmt.Sprintf("%d %s (%s) %s", 1, card.Name, card.Set, card.CollectorNumber)
			contents.Cards = append(contents.Cards, c)
		}
		boosters = append(boosters, contents)
	}

	enc := json.NewEncoder(w)
	enc.Encode(boosters)
}

func outputPlain(w http.ResponseWriter, boosterData []collectionfinder.BoosterContents) {
	w.Header().Add("Content-Type", "text/plain")

	for i, booster := range boosterData {
		fmt.Fprintf(w, "Booster #%d\n", i)

		for _, id := range booster.CardIds {
			card := db.GetCardByID(id)
			fmt.Fprintf(w, "%d %s (%s) %s\n", 1, card.Name, card.Set, card.CollectorNumber)
		}

		fmt.Fprintf(w, "\nCommon Wildcards: %d\nUncommon Wildcards: %d\nRare Wildcards: %d\nMythic Wildcards: %d\n",
			booster.CommonWildcards, booster.UncommonWildcards, booster.RareWildcards, booster.MythicWildcards)
	}
}

func uploadHandler(dc carddb.CardDB, jsonFormat bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(maxMtgaLogsSize); err != nil {
			log.Printf("could not parse multipart form: %v", err)
			http.Error(w, "Invalid Request", http.StatusPreconditionFailed)
		}

		file, _, err := r.FormFile("mtgalogs")
		if err != nil {
			log.Printf("couldnt get uploaded file: %v", err)
			http.Error(w, "Could not retrieve mtg logs file", http.StatusPreconditionFailed)
			return
		}
		defer file.Close()
		boosterData, err := collectionfinder.FindBoosters(file)
		if err != nil {
			log.Fatalf("failed to parse mtga logs: %v", err)
		}
		if jsonFormat {
			outputJSON(w, boosterData)
		} else {
			outputPlain(w, boosterData)
		}
	}
}

func boosterTracker(landingpagedata []byte) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write(landingpagedata)
	}
}

var db carddb.CardDB

func main() {
	flag.Parse()

	log.Println("Parsing MTG Data Files...")

	landingpagedata, err := ioutil.ReadFile(*landingpage)
	if err != nil {
		log.Fatalf("failed to parse landing page file: %v", err)
	}

	db, err = carddb.CreateLibrary(*mtgDataPath)
	if err != nil {
		log.Fatalf("createLibrary failed: %v", err)
	}

	log.Println("Starting Server")
	http.HandleFunc("/upload", uploadHandler(db, *jsonFormat))
	http.HandleFunc("/boostertracking", boosterTracker(landingpagedata))
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}
