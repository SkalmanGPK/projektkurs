package main

import (
	"fmt"
	"net/http"
	"time"
)

// simple IDS state
var (
	requestCount int
)

func idsHandler(w http.ResponseWriter, r *http.Request) {
	requestCount++

	// check headers
	ua := r.Header.Get("User-Agent")
	if ua == "AttackTool/1.0" {
		fmt.Printf("Malicious User-Agent detected: %s from IP %s\n", ua, r.RemoteAddr)
	}

	// simulate DPI
	if r.URL.RawQuery != "" {
		fmt.Printf("Located suspicious query string: %s\n", r.URL.RawQuery)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Traffic Monitored")
}

func main() {

	http.HandleFunc("/monitor", idsHandler)

	// background DoS detection
	go func() {
		for {
			time.Sleep(10 * time.Second)

			if requestCount > 50 {
				fmt.Printf("Traffic spike detected! %d requests in 10s. Possible DoS attack.\n", requestCount)
			}

			requestCount = 0
		}
	}()

	fmt.Println("IDS sensor is active on port 8080...")
	http.ListenAndServe(":8080", nil)
}
