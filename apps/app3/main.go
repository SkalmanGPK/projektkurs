package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// Data structure from App 1
type SensorData struct {
	Timestamp   string `json:"timestamp"`
	MachineID   string `json:"machine_id"`
	Temperature int    `json:"temperature"`
	Pressure    int    `json:"pressure"`
}

func main() {
	// 1. Connect to the database (ExternalName route)
	// If this fails, the app will crash and K8s will restart it until it works.
	db, _ := sql.Open("mysql", "sensor_user:password123@tcp(mysql-external-service:3306)/industrial_db")

	// 2. Data processing endpoint
	http.HandleFunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
		var data SensorData
		json.NewDecoder(r.Body).Decode(&data)

		// Simple filter logic
		if data.Temperature > 95 {
			fmt.Printf("!!! CRITICAL TEMP on %s: %d !!!\n", data.MachineID, data.Temperature)
		}

		// Save to DB
		db.Exec("INSERT INTO sensor_data (timestamp, machine_id, temperature, pressure) VALUES (?, ?, ?, ?)",
			data.Timestamp, data.MachineID, data.Temperature, data.Pressure)

		fmt.Println("Stored data from:", data.MachineID)
	})

	// 3. Health endpoint for the sidecar-proxy
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	fmt.Println("App3 running on :8081...")
	http.ListenAndServe(":8081", nil)
}
