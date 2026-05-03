package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type SensorData struct {
	Timestamp   string `json:"timestamp"`
	MachineID   string `json:"machine_id"`
	Temperature int    `json:"temperature"`
	Pressure    int    `json:"pressure"`
}

func main() {
	var received int
	var stored int

	db, _ := sql.Open("mysql", "sensor_user:password123@tcp(mysql-external-service:3306)/industrial_db")

	http.HandleFunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
		received++
		var data SensorData
		json.NewDecoder(r.Body).Decode(&data)

		if data.Temperature > 95 {
			fmt.Printf("!!! CRITICAL TEMP on %s: %d !!!\n", data.MachineID, data.Temperature)
		}

		db.Exec("INSERT INTO sensor_data (timestamp, machine_id, temperature, pressure) VALUES (?, ?, ?, ?)",
			data.Timestamp, data.MachineID, data.Temperature, data.Pressure)
		
		stored++

		fmt.Println("Stored data from:", data.MachineID)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"received": %d, "stored": %d}`, received, stored)
	})
	
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/ingest"
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	fmt.Println("App3 running on :8081...")
	http.ListenAndServe(":8081", nil)
}
