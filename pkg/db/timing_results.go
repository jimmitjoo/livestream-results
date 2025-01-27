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
	EventID        int
	EventName      string
}

func GetEvents(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT event_name FROM events")
	if err != nil {
		return nil, fmt.Errorf("error retrieving events: %w", err)
	}
	defer rows.Close()

	var events []string
	for rows.Next() {
		var eventName string
		if err := rows.Scan(&eventName); err != nil {
			return nil, fmt.Errorf("error scanning event name: %w", err)
		}
		events = append(events, eventName)
	}

	return events, nil
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

func GetParticipants(db *sql.DB) (map[string][]Participant, error) {
	// participants should be in groups of event name
	participants := make(map[string][]Participant)

	query := `
    SELECT 
        events.event_name AS event_name, 
        participants.bib_number, 
        participants.event_id, 
        participants.first_name, 
        participants.last_name, 
        participants.birthdate, 
        participants.club 
    FROM participants
    JOIN events ON events.event_id = participants.event_id
    `

	rows, err := db.Query(query)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, fmt.Errorf("error retrieving participants: %w", err)
	}
	defer rows.Close()

	// Iterate over the rows and group them based on event name
	for rows.Next() {
		var participant Participant
		if err := rows.Scan(&participant.EventName, &participant.BibNumber, &participant.EventID, &participant.FirstName, &participant.LastName, &participant.Birthdate, &participant.Club); err != nil {
			return nil, fmt.Errorf("error scanning participant: %w", err)
		}
		// check if the event name already exists in the participants slice
		if participants[participant.EventName] == nil {
			participants[participant.EventName] = []Participant{}
		}

		participants[participant.EventName] = append(participants[participant.EventName], participant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error with rows: %w", err)
	}

	return participants, nil
}

func GetParticipantByBibNumber(db *sql.DB, bibNumber int) (Participant, error) {
	var participant Participant
	err := db.QueryRow("SELECT event_id, first_name, last_name, birthdate, club FROM participants WHERE bib_number = ?", bibNumber).Scan(&participant.EventID, &participant.FirstName, &participant.LastName, &participant.Birthdate, &participant.Club)
	if err != nil {
		return Participant{}, fmt.Errorf("error retrieving participant by bib number: %w", err)
	}

	return participant, nil
}

// InsertTimingResult inserts a TimingResult into the timing_results table
func InsertTimingResult(db *sql.DB, result parser.TimingResult, participant Participant) error {
	query := `INSERT INTO timing_results (bib_number, event_id, timestamp, antenna_row, antenna, placement)
              VALUES (?, ?, ?, ?, ?, NULL)`

	_, err := db.Exec(query, result.BibNumber, participant.EventID, result.Timestamp.Format("2006-01-02 15:04:05.000"), result.AntennaRow, result.Antenna)
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
