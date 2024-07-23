package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/jimmitjoo/livestream-results/pkg/db"
	"github.com/jimmitjoo/livestream-results/pkg/parser"
	"github.com/jimmitjoo/livestream-results/pkg/sheets"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"strconv"
)

var database *sql.DB
var watcher *fsnotify.Watcher
var sheetsService *sheets.SheetsService
var sheetName string

func main() {
	var err error

	// Set up the database
	database, err = db.SetupDatabase()
	if err != nil {
		log.Fatalf("Error setting up the database: %v", err)
	}
	defer database.Close()

	// Initialize file watcher
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Error creating file watcher: %v", err)
	}
	defer watcher.Close()

	// Set up Google Sheets service
	sheetsService, err = sheets.NewSheetsService("credentials.json", "1bRygOoC50s3AZT8lfUpZl2EvWGHfEpEyh-_X-9r6xAc")
	if err != nil {
		log.Fatalf("Error setting up Google Sheets service: %v", err)
	}

	// Set up HTTP handlers
	http.HandleFunc("/start-watch", startWatchHandler)
	http.HandleFunc("/google-sheets", googleSheetsHandler)
	http.HandleFunc("/read-startlista", readParticipantsHandler)
	http.HandleFunc("/list-participants", listParticipantsHandler)

	// Serve static files from the frontend directory
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func startWatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		FilePath string `json:"filePath"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if requestData.FilePath == "" {
		http.Error(w, "File path is required", http.StatusBadRequest)
		return
	}

	// Start watching the specified file
	go watchFile(requestData.FilePath)

	fmt.Fprintf(w, "Started watching file: %s", requestData.FilePath)
}

func googleSheetsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		SheetID   string `json:"sheetID"`
		SheetName string `json:"sheetName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sheetsService.SheetID = requestData.SheetID
	sheetName = requestData.SheetName

	fmt.Fprintf(w, "Google Sheets ID: %s, Sheet Name: %s", requestData.SheetID, requestData.SheetName)
}

func readParticipantsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		PrimaryEventName      string `json:"primaryEventName"`
		ParticipantsSheetName string `json:"participantsSheetName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Event Name:", requestData.PrimaryEventName)
	fmt.Println("Sheet Name:", requestData.ParticipantsSheetName)

	data, err := sheetsService.ReadSheet(requestData.ParticipantsSheetName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading participants: %v", err), http.StatusInternalServerError)
		return
	}

	primaryEventID, err := db.GetEventByName(database, requestData.PrimaryEventName)
	if err != nil {
		primaryEventID, err = db.CreateEvent(database, requestData.PrimaryEventName, 0, "")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating primary event: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Loop through the data and log
	for _, row := range data {
		if len(row) < 1 {
			fmt.Println("Row is empty")
			continue
		}
		var bibNumber int
		// check if row[0] interface value is a number
		// if not, skip the row
		if _, ok := row[0].(float64); !ok {

			// Try to convert the value to an int
			if _, err := strconv.Atoi(row[0].(string)); err != nil {
				fmt.Println("Row does not have a valid bib number: ", row[0])
				continue
			}

			bibNumber, err = strconv.Atoi(row[0].(string)) // Convert the value to an int
			if err != nil {
				fmt.Println("Error converting bib number to int: ", err)
				continue
			}
			fmt.Println("Bib Number:", bibNumber)
		}

		// if row[9] does not exist, skip the row
		if len(row) < 8 {
			fmt.Println("Row does not have enough columns")
			continue
		}

		// Get the first name from the row
		firstName := row[1].(string)
		fmt.Println("First Name:", firstName)

		// Get the last name from the row
		lastName := row[2].(string)
		fmt.Println("Last Name:", lastName)

		// Get the birthdate from the row
		birthdate := row[3].(string)
		fmt.Println("Birthdate:", birthdate)

		// Get the club from the row
		club := row[4].(string)
		fmt.Println("Club:", club)

		// Get the classification from the row
		classification := row[5].(string)
		fmt.Println("Classification:", classification)

		// Get the gender from the row
		var gender string
		genderString := row[6].(string)
		if genderString == "K" || genderString == "F" || genderString == "W" || genderString == "Kvinna" || genderString == "Woman" {
			gender = "F"
		}
		if genderString == "M" || genderString == "Man" || genderString == "Male" {
			gender = "M"
		}

		fmt.Println("Gender", gender)

		eventName := requestData.PrimaryEventName + " " + classification
		fmt.Println("Event: ", eventName)

		// Check if the event exists in the database
		eventID, err := db.GetEventByName(database, eventName)
		if err != nil {
			eventID, err = db.CreateEvent(database, eventName, primaryEventID, classification)
		}
		if eventID > 0 {
			participant := db.Participant{
				BibNumber:      bibNumber,
				FirstName:      firstName,
				LastName:       lastName,
				Gender:         gender,
				Birthdate:      birthdate,
				Club:           club,
				Classification: classification,
			}

			err := db.InsertParticipant(database, participant, eventID)
			if err != nil {
				fmt.Println("Error inserting participant: ", err)
				continue
			}

			fmt.Println("Event does not exist in the database")
			continue
		}

		fmt.Println("-----")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func listParticipantsHandler(w http.ResponseWriter, r *http.Request) {

	// Group participants by event
	participants, err := db.GetParticipants(database)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting participants: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(participants)
}

func watchFile(filePath string) {
	err := watcher.Add(filePath)
	if err != nil {
		log.Fatalf("Error adding file to watcher: %v", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Chmod) != 0 {
				log.Printf("File modified: %s", event.Name)
				results, err := parser.ParseTimingFile(filePath)
				if err != nil {
					log.Printf("Error parsing timing data: %v", err)
					continue
				}

				for _, result := range results {
					// Find participant by bib number
					participant, err := db.GetParticipantByBibNumber(database, result.BibNumber)

					err = db.InsertTimingResult(database, result, participant)
					if err != nil {
						log.Printf("Error inserting timing result for bib number %d: %v", result.BibNumber, err)
					}
				}

				log.Println("Timing data parsed and inserted successfully!")

				// Get new data and update Google Sheets
				data, err := getNewData()
				if err != nil {
					log.Printf("Error getting new data: %v", err)
					continue
				}

				err = sheetsService.UpdateSheet(sheetName, data)
				if err != nil {
					log.Printf("Error updating Google Sheets: %v", err)
				} else {
					log.Println("Google Sheets updated successfully")
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Error watching file: %v", err)
		}
	}
}

func getNewData() ([][]interface{}, error) {
	// Retrieve new data from the database
	rows, err := database.Query("SELECT bib_number, timestamp, placement FROM timing_results ORDER BY timestamp ASC LIMIT 10000")
	if err != nil {
		return nil, fmt.Errorf("error querying new data: %v", err)
	}
	defer rows.Close()

	var data [][]interface{}
	for rows.Next() {
		var bibNumber int
		var timestamp string
		// placement is a nullable column, so we need to use a pointer
		var placement *int

		if err := rows.Scan(&bibNumber, &timestamp, &placement); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		participant, err := db.GetParticipantByBibNumber(database, bibNumber)
		if err != nil {
			return nil, fmt.Errorf("error getting participant by bib number: %v", err)
		}

		data = append(data, []interface{}{bibNumber, participant.FirstName, participant.LastName, participant.Club, participant.Birthdate, timestamp, placement})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error with rows: %v", err)
	}

	return data, nil
}
