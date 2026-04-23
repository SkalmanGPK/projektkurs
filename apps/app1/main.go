package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// SensorData defining the structure on the data we're sending (same as app 3 is expecting)
type SensorData struct {
	Timestamp   string `json:"timestamp"`
	MachineID   string `json:"machine_id"`
	Temperature int    `json:"temperature"`
	Pressure    int    `json:"pressure"`
}

func main() {
	// 1. Starting a simple health-server (for the sidecar-proxy) in it's own thread
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "OK")
		})
		fmt.Println("Health-server started on port 8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Printf("Health-server failed: %v\n", err)
		}
	}()

	// URL to app 2 (IDS proxy)
	// In kubernetes it is reached via it's servicename
	app2URL := "http://ids-app2-service.default.svc.cluster.local:80/monitor"

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	fmt.Println("App1 (Sensor Simulator) started, sending data to app2...")

	//2. Main loop to generate and send data
	for {
		// create random data
		data := SensorData{
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
		MachineID:   fmt.Sprintf("MC-%d", rand.Intn(5)+1),
		Temperature: rand.Intn(81) + 20, // 20-100
		Pressure:    rand.Intn(201) + 900, // 900-1100
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("Error with JSON: %v\n", err)
			continue
		}

		// send POST to app2
		resp, err := client.Post(app2URL, "application/json", bytes.NewBuffer(jsonData))

		if err != nil {
			fmt.Printf("[!] Failed to send to app 2: %v\n", err)
		} else {
			fmt.Printf("[+] sent data: %s | status from app 2: %s\n", string(jsonData), resp.Status)
			resp.Body.Close() // close body
		}

		//wait 5 seconds before next simulation
		time.Sleep(5 * time.Second)
	}
}
