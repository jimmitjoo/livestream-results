package main

import (
	"database/sql"
	"fmt"
	"github.com/jimmitjoo/livestream-results/pkg/db"
	"log"
)

func main() {
	// Set up the database
	db.SetupDatabase()

	// Connect to the database
	database, err := sql.Open("sqlite3", "./race_timing.db")
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer func(database *sql.DB) {
		err := database.Close()
		if err != nil {
			log.Fatalf("Error closing the database: %v", err)
		}
	}(database)

	// Parse the timing data file
	results, err := db.ParseTimingFile("path/to/timing_data.txt")
	if err != nil {
		log.Fatalf("Error parsing timing data: %v", err)
	}

	// Insert parsed results into the database
	eventID := 1 // Assume an event ID for demonstration purposes
	for _, result := range results {
		err := db.InsertTimingResult(database, result, eventID)
		if err != nil {
			log.Printf("Error inserting timing result for bib number %d: %v", result.BibNumber, err)
		}
	}

	fmt.Println("Timing data parsed and inserted successfully!")
}
