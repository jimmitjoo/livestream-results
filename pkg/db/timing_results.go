package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jimmitjoo/livestream-results/pkg/parser"
	"github.com/mattn/go-sqlite3"
)

type Participant struct {
	BibNumber      int
	FirstName      string
	LastName       string
	Gender         string
	Birthdate      string
	Club           string
	Classification string
}

// GetEventByName retrieves an event by its name
func GetEventByName(db *sql.DB, eventName string) (int, error) {
	var eventID int
	err := db.QueryRow("SELECT event_id FROM events WHERE event_name = ?", eventName).Scan(&eventID)
	if err != nil {
		return 0, fmt.Errorf("error retrieving event by name: %w", err)
	}
	return eventID, nil
}

func CreateEvent(db *sql.DB, eventName string, parentEventID int, classification string) (int, error) {
	query := `INSERT INTO events (event_name, parent_event_id, classification) VALUES (?, ?, ?)`
	result, err := db.Exec(query, eventName, parentEventID, classification)
	if err != nil {
		return 0, fmt.Errorf("error creating event: %w", err)
	}

	eventID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting last insert ID: %w", err)
	}

	return int(eventID), nil
}

func InsertParticipant(db *sql.DB, participant Participant, eventID int) error {
	query := `INSERT INTO participants (event_id, bib_number, first_name, last_name, gender, birthdate, club, classification)
    		  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.Exec(query, eventID, participant.BibNumber, participant.FirstName, participant.LastName, participant.Gender, participant.Birthdate, participant.Club, participant.Classification)
	if err != nil {
		// Check if the error is a UNIQUE constraint violation
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrConstraint {
			return nil // Suppress the error as it is expected
		}
		return fmt.Errorf("error inserting participant: %w", err)
	}

	return nil
}

// InsertTimingResult inserts a TimingResult into the timing_results table
func InsertTimingResult(db *sql.DB, result parser.TimingResult, eventID int) error {
	query := `INSERT INTO timing_results (bib_number, event_id, timestamp, antenna_row, antenna, placement)
              VALUES (?, ?, ?, ?, ?, NULL)`

	fmt.Println(query)
	_, err := db.Exec(query, result.BibNumber, eventID, result.Timestamp.Format("2006-01-02 15:04:05.000"), result.AntennaRow, result.Antenna)
	if err != nil {
		// Check if the error is a UNIQUE constraint violation
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrConstraint {
			return nil // Suppress the error as it is expected
		}
		return fmt.Errorf("error inserting timing result: %w", err)
	}
	return nil
}
