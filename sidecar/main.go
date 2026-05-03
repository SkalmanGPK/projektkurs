package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

func main() {
	targets := map[string]string{
		"App1 (Sensor)": "http://app1-service:8080/health",
		"App2 (Proxy) ": "http://ids-app2-service:80/health",
		"App3 (Filter)": "http://app3-service:8081/health",
	}

	client := http.Client{Timeout: 2 * time.Second}
	fmt.Println("--- Network Latency Verifier Started ---")

	for {
		for name, url := range targets {
			start := time.Now()
			resp, err := client.Get(url)

			if err != nil {
				fmt.Printf("[!] %s is unreachable: %v\n", name, err)
				continue
			}

			latency := time.Since(start)
			fmt.Printf("[+] %s | Latency: %v | Status: %s\n", name, latency, resp.Status)
			resp.Body.Close()
		}

		// Hämta stats (received vs stored enligt app3)
		statsResp, err := client.Get("http://app3-service:8081/stats")
		if err != nil {
			fmt.Printf("[!] Kunde inte hamta stats fran App3: %v\n", err)
		} else {
			buf := new(bytes.Buffer)
			buf.ReadFrom(statsResp.Body)
			statsResp.Body.Close()
			fmt.Printf("[STATS] App3: %s\n", buf.String())
		}

		// Hämta faktiskt antal rader i databasen
		dbResp, err := client.Get("http://app3-service:8081/dbcount")
		if err != nil {
			fmt.Printf("[DBCOUNT] App3: {\"db_actual\": -1}\n")
		} else {
			buf := new(bytes.Buffer)
			buf.ReadFrom(dbResp.Body)
			dbResp.Body.Close()
			fmt.Printf("[DBCOUNT] App3: %s\n", buf.String())
		}

		fmt.Println("---------------------------------------")
		time.Sleep(10 * time.Second)
	}
}
