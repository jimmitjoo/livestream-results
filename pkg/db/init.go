package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

// SetupDatabase initializes the database and creates necessary tables
func SetupDatabase() (*sql.DB, error) {
	// Connect to the database (creates the database file if it doesn't exist)
	db, err := sql.Open("sqlite3", "./race_timing.db")
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	// Create the tables
	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// createTables creates the necessary tables in the database
func createTables(db *sql.DB) error {
	// Create events table
	eventsTable := `CREATE TABLE IF NOT EXISTS events (
        event_id INTEGER PRIMARY KEY AUTOINCREMENT,
        event_name TEXT NOT NULL,
        parent_event_id INTEGER,
        classification TEXT,
        FOREIGN KEY (parent_event_id) REFERENCES events(event_id)
    );`
	if _, err := db.Exec(eventsTable); err != nil {
		return fmt.Errorf("error creating events table: %w", err)
	}

	// Create participants table
	participantsTable := `CREATE TABLE IF NOT EXISTS participants (
        participant_id INTEGER PRIMARY KEY AUTOINCREMENT,
        event_id INTEGER NOT NULL,
        bib_number INTEGER NOT NULL,
        first_name TEXT NOT NULL,
        last_name TEXT NOT NULL,
        gender TEXT NOT NULL,
        birthdate TEXT NOT NULL,
        club TEXT,
        classification TEXT,
        FOREIGN KEY (event_id) REFERENCES events(event_id),
    	UNIQUE (bib_number, event_id)
    );`
	if _, err := db.Exec(participantsTable); err != nil {
		return fmt.Errorf("error creating participants table: %w", err)
	}

	// Create timing_results table
	timingResultsTable := `CREATE TABLE IF NOT EXISTS timing_results (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        bib_number INTEGER NOT NULL,
        event_id INTEGER NOT NULL,
        timestamp TEXT NOT NULL,
        antenna_row INTEGER,
        antenna INTEGER,
        placement INTEGER,
        FOREIGN KEY (event_id) REFERENCES events(event_id),
        UNIQUE (bib_number, event_id, timestamp)
    );`
	if _, err := db.Exec(timingResultsTable); err != nil {
		return fmt.Errorf("error creating timing_results table: %w", err)
	}

	return nil
}
