package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	// List of targets based on the Services we created in Kubernetes
	// We use the port 80 for App 2 and 8080/8081 for the others
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
		
		fmt.Println("---------------------------------------")
		time.Sleep(10 * time.Second) // Wait between sweeps
	}
}
