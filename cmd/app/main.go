package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jimmitjoo/livestream-results/pkg/db"
	"github.com/jimmitjoo/livestream-results/pkg/parser"
	"log"
	"net/http"

	"github.com/fsnotify/fsnotify"
	_ "github.com/mattn/go-sqlite3"
)

var database *sql.DB
var watcher *fsnotify.Watcher

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

	// Set up HTTP handlers
	http.HandleFunc("/", uploadHandler)
	http.HandleFunc("/start-watch", startWatchHandler)
	http.HandleFunc("/google-sheets", googleSheetsHandler)

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `
    <html>
    <body>
        <h1>Start Watching Timing Data File</h1>
        <form id="watch-form">
            <label for="filePath">Timing Data File Path:</label><br>
            <input type="text" id="filePath" name="filePath" placeholder="e.g., C:\path\to\file.txt"><br>
            <input type="submit" value="Start Watching">
        </form>
        <p id="watch-feedback"></p>
        <h1>Google Sheets Export</h1>
        <form id="sheets-form">
            <label for="sheetID">Google Sheets ID:</label><br>
            <input type="text" id="sheetID" name="sheetID"><br>
            <label for="sheetName">Sheet Name:</label><br>
            <input type="text" id="sheetName" name="sheetName"><br>
            <input type="submit" value="Submit">
        </form>
        <p id="sheets-feedback"></p>

        <script>
            document.getElementById('watch-form').addEventListener('submit', function(event) {
                event.preventDefault();
                const filePath = document.getElementById('filePath').value;
                fetch('/start-watch', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ filePath })
                })
                .then(response => response.text())
                .then(data => {
                    document.getElementById('watch-feedback').innerText = data;
                })
                .catch(error => {
                    document.getElementById('watch-feedback').innerText = 'Error starting watch: ' + error;
                });
            });

            document.getElementById('sheets-form').addEventListener('submit', function(event) {
                event.preventDefault();
                const sheetID = document.getElementById('sheetID').value;
                const sheetName = document.getElementById('sheetName').value;
                fetch('/google-sheets', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ sheetID, sheetName })
                })
                .then(response => response.text())
                .then(data => {
                    document.getElementById('sheets-feedback').innerText = data;
                })
                .catch(error => {
                    document.getElementById('sheets-feedback').innerText = 'Error submitting Google Sheets info: ' + error;
                });
            });
        </script>
    </body>
    </html>
    `
	fmt.Fprint(w, tmpl)
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

	// Process Google Sheets details
	fmt.Fprintf(w, "Google Sheets ID: %s, Sheet Name: %s", requestData.SheetID, requestData.SheetName)

	// Implement Google Sheets export functionality here
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

				// Insert parsed results into the database
				eventID := 1 // Assume an event ID for demonstration purposes
				for _, result := range results {
					err := db.InsertTimingResult(database, result, eventID)
					if err != nil {
						log.Printf("Error inserting timing result for bib number %d: %v", result.BibNumber, err)
					}
				}

				log.Println("Timing data parsed and inserted successfully!")
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Error watching file: %v", err)
		}
	}
}
