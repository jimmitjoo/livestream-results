package main

import (
	"database/sql"
	"github.com/jimmitjoo/livestream-results/pkg/db"
	"log"

	"github.com/fsnotify/fsnotify"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Set up the database
	db.SetupDatabase()

	// Connect to the database
	database, err := sql.Open("sqlite3", "./race_timing.db")
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer database.Close()

	// Start watching the file
	filePath := "path/to/timing_data.txt" // Replace with your file path
	go watchFile(filePath, database)

	// Keep the program running
	select {}
}

// watchFile watches the specified file and triggers the parsing function on modifications
func watchFile(filePath string, database *sql.DB) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Error creating file watcher: %v", err)
	}
	defer func(watcher *fsnotify.Watcher) {
		err := watcher.Close()
		if err != nil {
			log.Fatalf("Error closing file watcher: %v", err)
		}
	}(watcher)

	err = watcher.Add(filePath)
	if err != nil {
		log.Fatalf("Error adding file to watcher: %v", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("Watcher event channel closed")
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Chmod) != 0 {
				results, err := db.ParseTimingFile(filePath)
				if err != nil {
					log.Printf("Error parsing timing data: %v", err)
					continue
				}

				// Insert parsed results into the database
				eventID := 1 // Assume an event ID for demonstration purposes
				for _, result := range results {
					err := db.InsertTimingResult(database, result, eventID)
					if err != nil {
						log.Printf("Error inserting timing result for bib number %d: %v", result.BibNumber, err)
					}
				}
			} else {
				log.Printf("File event not recognized: %v", event)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				log.Println("Watcher error channel closed")
				return
			}
			log.Printf("Error watching file: %v", err)
		}
	}
}
