package main

import (
	"database/sql"
	"enconding/json"
	"fmt"
	"log"
	"net/http"
)

type SensorData struct {
	Timestamp string 'json:"timestamp"'
	MachineID string 'json:"machine_id"'
	Temperature int 'json:"temperature"'
	Pressure int 'json:"pressure"'
}

var db *sql.DB

func ingestHandler(w http.ResponseWriter, r *http.Request) {
	
	var data SensorData

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "bad request", 400)
		return
	}

	// filtering logic
	if data.Temperature > 95 {
		fmt.Printf("alert: overheating machine %s\n", data.MachineID)
	}

	query := `
	INSERT INTO sensor_data (timestamp, machine_id, temperature, pressure)
	VALUES (?, ?, ?, ?)
	`
	_, err = db.Exec(query,
		data.Timestamp,
		data.MachineID,
		data.Temperature,
		data.Pressure)

	if err != nil {
		log.Println("DB error:", err)
	}

	fmt.Println("Stored sensor data:", data)

	w.WriteHeader(http.StatusOK)
}

func main() {

	var err error

	db, err = sql.Open("mysql",
		"sensor_user:password123@tcp(mysql-service:3306)/industrial_db")

	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/ingest", ingestHandler)

	fmt.Println("App3 listening on :8081")
	http.ListenAndServe(":8081", nil)
}
