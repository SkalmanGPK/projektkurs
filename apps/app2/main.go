package main

import (
	"fmt"
	"net/http"
	"time"
)

// Enkel struktur för att hålla koll på misstänkt aktivitet
var (
	requestCount int
	lastReset time.Time
)

func idsHandler(w http.ResponseWriter, r *http.Request) {
	requestCount++

	// analysera Headers
	// om någon skickar ett konstigt User-Agent, flagga det
	ua := r.Header.Get("User-Agent")
	if ua == "AttackTool/1.0" {
		fmt.Printf("Malicious User-Agent detected: $s from IP $s\n", ua, r.RemoteAddr)
	}

	// simulera deep packet inspection (DPI)
	// Om "SQL injection"-liknande strängar finns i URL:en
	if r.URL.RawQuery != "" {
		fmt.Printf("Located suspicious query string: %s\n", r.URL.RawQuery)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Traffic Monitored")

	func main() {
		// monitoring endpoint
		http.HandleFunc("/monitor", idsHandler)

		// Bakgrundsprocess som kollar efter Dos attacker
		go func() {
			for {
				time.sleep(10 * time.Second)
				if requestCount > 50 {
					fmt.Printf("Traffic spike detected! %d requests in 10s. Possible DoS attack.\n", RequestCount)
				}
				requestCount = 0
			}
		}()

		fmt.Println("IDS sensor is active on port 8080...)
		http.ListendAndServe(":8080", nil)
	}
	
