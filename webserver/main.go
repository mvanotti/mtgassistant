// program webserver is a web server that parses out MTG Arena Logs and outputs the contents of the cards boosters that you opened
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/mvanotti/mtgassistant/carddb"
	"github.com/mvanotti/mtgassistant/collectionfinder"
)

var (
	mtgDataPath = flag.String("mtg_data", `C:\Program Files (x86)\Wizards of the Coast\MTGA\MTGA_Data\Downloads\Data`, "Path to the Downloads\\Data folder inside the MTG Arena Install Directory")
	landingpage = flag.String("landing", "boostertracking.html", "Path to the landing page.")
)

const maxMtgaLogsSize int64 = 100 << 20 // 20 MiB

func uploadHandler(w http.ResponseWriter, r *http.Request) {
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

	fmt.Fprintf(w, "Iterating over %d boosters\n", len(boosterData))

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

func boosterTracker(w http.ResponseWriter, r *http.Request) {
	f, err := os.OpenFile(*landingpage, os.O_RDONLY, 0)
	if err != nil {
		log.Printf("failed to open landing page: %v", err)
		http.Error(w, "Could not open landing page", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(w, f); err != nil {
		log.Printf("failed to copy landing page: %v", err)
		http.Error(w, "Could not copy landing page", http.StatusInternalServerError)
	}
	f.Close()
}

var db carddb.CardDB

func main() {
	flag.Parse()

	log.Println("Parsing MTG Data Files...")

	var err error
	db, err = carddb.CreateLibrary(*mtgDataPath)
	if err != nil {
		log.Fatalf("createLibrary failed: %v", err)
	}
	log.Println("Starting Server")
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/boostertracking", boosterTracker)
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}
