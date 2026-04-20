package main

import (
	"fmt"
	"net/http"
	"time"
)

// Global variable to keep track of amount of calls (permission for IDS)
var (
	requestCount int
)
// idsHandler thats ran at every incoming HTTP-call
func idsHandler(w http.ResponseWriter, r *http.Request) {
	requestCount++ // raises the count for every visitor

	// Signaturebased detection: Looking for specific "User-Agents"
	// User-agent often shows if it is a regular web browser or a script thats making the call.
	ua := r.Header.Get("User-Agent")
	if ua == "AttackTool/1.0" {
		fmt.Printf("Malicious User-Agent detected: %s from IP %s\n", ua, r.RemoteAddr)
	}

	// DPI (Deep Packet Inspection) Simulation:
	// Looking after suspicious strings in the URL (for example SQL-injections or script)
	if r.URL.RawQuery != "" {
		fmt.Printf("Located suspicious query string: %s\n", r.URL.RawQuery)
	}
	// Answering the client that the trafic is being supervised
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Traffic Monitored")
}

func main() {
	// registrating the endpoint /monitor
	http.HandleFunc("/monitor", idsHandler)

	// Anomaly-detection (DoS-preventation) is ran in it's own "Goroutine" (background)
	go func() {
		for {
			// Waiting in 10 sec between every control
			time.Sleep(10 * time.Second)
			
			// If the amout of calls is above 50 during the last 10 seconds it's flagged as DoS
			if requestCount > 50 {
				fmt.Printf("Traffic spike detected! %d requests in 10s. Possible DoS attack.\n", requestCount)
			}
			// The counter is set to 0 for the next 10 second duration
			requestCount = 0
		}
	}()

	fmt.Println("IDS sensor is active on port 8080...")
	// Starting server and listening on port 8080
	http.ListenAndServe(":8080", nil)
}
