package db

import (
	"database/sql"
	"fmt"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

func SetupDatabase() {
	// Connect to the database (creates the database file if it doesn't exist)
	db, err := sql.Open("sqlite3", "./race_timing.db")
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			fmt.Println("Error closing the database:", err)
		}
	}(db)

	// Create the tables
	createTables(db)
}

func createTables(db *sql.DB) {
	// Create events table
	eventsTable := `CREATE TABLE IF NOT EXISTS events (
        event_id INTEGER PRIMARY KEY AUTOINCREMENT,
        event_name TEXT NOT NULL,
        parent_event_id INTEGER,
        classification TEXT,
        FOREIGN KEY (parent_event_id) REFERENCES events(event_id)
    );`
	_, err := db.Exec(eventsTable)
	if err != nil {
		fmt.Println("Error creating events table:", err)
		return
	}

	// Create participants table
	participantsTable := `CREATE TABLE IF NOT EXISTS participants (
        participant_id INTEGER PRIMARY KEY AUTOINCREMENT,
        bib_number INTEGER NOT NULL,
        first_name TEXT NOT NULL,
        last_name TEXT NOT NULL,
        gender TEXT NOT NULL,
        birthdate TEXT NOT NULL,
        club TEXT,
        classification TEXT
    );`
	_, err = db.Exec(participantsTable)
	if err != nil {
		fmt.Println("Error creating participants table:", err)
		return
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
        FOREIGN KEY (event_id) REFERENCES events(event_id)
    	UNIQUE(bib_number, event_id, timestamp)
    );`
	_, err = db.Exec(timingResultsTable)
	if err != nil {
		fmt.Println("Error creating timing_results table:", err)
		return
	}

	// Create participant_events table
	participantEventsTable := `CREATE TABLE IF NOT EXISTS participant_events (
        participant_event_id INTEGER PRIMARY KEY AUTOINCREMENT,
        participant_id INTEGER NOT NULL,
        event_id INTEGER NOT NULL,
        FOREIGN KEY (participant_id) REFERENCES participants(participant_id),
        FOREIGN KEY (event_id) REFERENCES events(event_id)
    );`
	_, err = db.Exec(participantEventsTable)
	if err != nil {
		fmt.Println("Error creating participant_events table:", err)
		return
	}
}

// InsertTimingResult inserts a TimingResult into the timing_results table
func InsertTimingResult(db *sql.DB, result TimingResult, eventID int) error {
	query := `INSERT INTO timing_results (bib_number, event_id, timestamp, antenna_row, antenna, placement)
              VALUES (?, ?, ?, ?, ?, NULL)`
	_, err := db.Exec(query, result.BibNumber, eventID, result.Timestamp.Format("2006-01-02 15:04:05.000"), result.AntennaRow, result.Antenna)
	if err != nil {
		// Check if the error is a UNIQUE constraint violation
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
			return nil // Suppress the error as it is expected
		}
		return fmt.Errorf("error inserting timing result: %w", err)
	}
	return nil
}
