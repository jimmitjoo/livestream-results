package main

import (
	"fmt"
	"github.com/jimmitjoo/livestream-results/pkg/db"
)

func main() {
	fmt.Println("Setting up the database...")
	db.SetupDatabase()
	fmt.Println("Database setup complete.")
}
