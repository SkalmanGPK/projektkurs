package main // defining that this is a runnable application

import (
	"database/sql" // package to speak with SQL databases
	"encoding/json" // used for JSON
	"fmt" //to write text in the terminal
	"log" // Logging issues
	"net/http" // To create webservers
	- "github.com/go-sql-driver/mysql" // To let go understand how to speak with mysql
)

type SensorData struct { // mapping JSON-fields to go-variables
	Timestamp string `json:"timestamp"`
	MachineID string `json:"machine_id"`
	Temperature int `json:"temperature"`
	Pressure int `json:"pressure"`
}

var db *sql.DB // A global variable that holds the database connection open to the entire app

func ingestHandler(w http.ResponseWriter, r *http.Request) { // handler function, runs everytime a call is made to the server.
	defer r.Body.Close() // Closes the stream when func is done.
	var data SensorData

	err := json.NewDecoder(r.Body).Decode(&data) // Reading JSON data from the call and saves it in the "data-variable"
	if err != nil {
		http.Error(w, "bad request", 400) // if JSON is broken, answer with code 400.
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
