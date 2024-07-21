package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jimmitjoo/livestream-results/pkg/parser"
	"github.com/mattn/go-sqlite3"
)

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
