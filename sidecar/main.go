package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	// Hämtar konfiguration från Kubernetes env-variabler
	nodeName := os.Getenv("NODE_NAME")
	podIP := os.Getenv("POD_IP")
	
	if nodeName == "" {
		nodeName = "unknown-node"
	}

	fmt.Printf("Starting sidecar proxy on node: %s (IP: %s)\n", nodeName, podIP)

	// 1. Starta en hälso-endpoint
	// Denna gör att andra sidecars kan pinga denna pod.
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			// Vi svarar med namnet på noden så vi ser var svaret kommer ifrån
			fmt.Fprintf(w, "OK (from node: %s, pod IP: %s)", nodeName, podIP)
		})
		
		port := ":8080"
		fmt.Printf("Sidecar health server listening on %s...\n", port)
		if err := http.ListenAndServe(port, nil); err != nil {
			fmt.Printf("HTTP Server failed: %v\n", err)
		}
	}()

	// 2. Konfigurera HTTP-klient för mätningar
	httpClient := &http.Client{
		Timeout: 3 * time.Second,
	}

	// Lista över targets (Services i Kubernetes)
	// Vi inkluderar app1-service för att kunna mäta latens även mot oss själva/våra replikor
	targets := []string{
		"app1-service",
		"app2-service",
		"app3-service",
	}

	// 3. Huvudloop för latensmätning
	for {
		fmt.Println("--- Starting latency sweep ---")
		
		for _, target := range targets {
			start := time.Now()
			url := fmt.Sprintf("http://%s:8080/health", target)

			resp, err := httpClient.Get(url)
			if err != nil {
				fmt.Printf("[!] Failed to reach %s: %v\n", target, err)
				continue
			}

			latency := time.Since(start)
			resp.Body.Close() // Viktigt: stäng alltid body för att undvika minnesläckor

			fmt.Printf("[+] Target: %-15s | Latency: %-10v | Status: %s\n", 
				target, latency, resp.Status)
		}

		// Vänta 5 sekunder innan nästa mätrunda
		fmt.Println("--- Sweep complete, sleeping 5s ---")
		time.Sleep(5 * time.Second)
	}
}
