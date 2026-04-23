package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	// 1. Define the destination: Where should the traffic go?
	// In your K8s cluster, App 3 is reachable via its service name and port.
	app3Url, _ := url.Parse("http://app3-service:8081")

	// 2. Initialize the Reverse Proxy
	// This built-in tool handles copying headers, body, and managing the connection.
	proxy := httputil.NewSingleHostReverseProxy(app3Url)

	// 3. Create the main handler for all incoming traffic
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		
		// --- IDS LOGIC START ---
		// Signature-based detection: Check for a specific "malicious" User-Agent
		if r.Header.Get("User-Agent") == "AttackTool/1.0" {
			fmt.Printf("[ALERT] Malicious traffic detected from IP: %s\n", r.RemoteAddr)
		}
		
		// Log that we are inspecting the sensor data packet
		if r.URL.Path == "/monitor" {
			fmt.Println("Inspecting incoming sensor data...")
		}
		// --- IDS LOGIC END ---

		// Forward the request automatically to App 3 (the target)
		// The proxy takes the response from App 3 and sends it back to App 1
		proxy.ServeHTTP(w, r)
	})

	// 4. Health endpoint for the sidecar-proxy
	// This allows your latency-testing tool to verify that App 2 is up.
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "IDS Proxy Active")
	})

	fmt.Println("Simplified IDS-Proxy listening on port 8080...")
	
	// Start the server on port 8080
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
