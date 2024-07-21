package parser

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// TimingResult represents a parsed timing result from the file
type TimingResult struct {
	BibNumber  int
	Timestamp  time.Time
	AntennaRow *int
	Antenna    *int
}

// parseIntField parses a string field to an integer pointer
func parseIntField(field string) (*int, error) {
	if field == "" {
		return nil, nil
	}
	value, err := strconv.Atoi(field)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// ParseTimingFile parses the timing data file and returns a slice of TimingResult
func ParseTimingFile(filePath string) ([]TimingResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var results []TimingResult

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			bibNumber, err := strconv.Atoi(parts[0])
			if err != nil {
				fmt.Println("error parsing bib number:", err)
				continue
			}
			timestamp, err := time.Parse("2006-01-02 15:04:05.000", parts[1])
			if err != nil {
				fmt.Println("error parsing timestamp:", err)
				continue
			}
			antennaRow, err := parseIntField(parts[2])
			if err != nil {
				fmt.Println("error parsing antenna_row:", err)
				continue
			}
			antenna, err := parseIntField(parts[3])
			if err != nil {
				fmt.Println("error parsing antenna:", err)
				continue
			}
			results = append(results, TimingResult{BibNumber: bibNumber, Timestamp: timestamp, AntennaRow: antennaRow, Antenna: antenna})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return results, nil
}
