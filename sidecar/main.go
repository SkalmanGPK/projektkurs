package main

import (
	"bytes"      // används för att läsa response body till en buffer
	"fmt"        // för utskrift i terminalen
	"net/http"   // för HTTP-anrop (GET requests)
	"time"       // för tidtagning och sleep
)

func main() {

	// Map som innehåller namn på tjänster + deras health endpoints
	targets := map[string]string{
		"App1 (Sensor)": "http://app1-service:8080/health",
		"App2 (Proxy) ": "http://ids-app2-service:80/health",
		"App3 (Filter)": "http://app3-service:8081/health",
	}

	// Skapar en HTTP-klient med timeout på 2 sekunder
	// Detta förhindrar att programmet fastnar om en tjänst inte svarar
	client := http.Client{Timeout: 2 * time.Second}

	fmt.Println("--- Network Latency Verifier Started ---")

	// Oändlig loop för kontinuerlig övervakning
	for {

		// Loopa igenom alla targets (tjänster)
		for name, url := range targets {

			// Starta tidtagning innan request skickas
			start := time.Now()

			// Skicka GET-request till tjänstens health endpoint
			resp, err := client.Get(url)

			// Om ett fel uppstår (t.ex. timeout eller tjänsten nere)
			if err != nil {
				fmt.Printf("[!] %s is unreachable: %v\n", name, err)
				continue // hoppa till nästa tjänst
			}

			// Beräkna latency (hur lång tid requesten tog)
			latency := time.Since(start)

			// Skriv ut resultat: namn, latency och HTTP-statuskod
			fmt.Printf("[+] %s | Latency: %v | Status: %s\n", name, latency, resp.Status)

			// Viktigt: stäng response body för att undvika minnesläckor
			resp.Body.Close()
		}

		// ============================
		// Hämta och logga statistik från App3
		// ============================

		statsResp, err := client.Get("http://app3-service:8081/stats")

		if err != nil {
			fmt.Printf("[!] Kunde inte hämta stats från App3: %v\n", err)
		} else {
			// Skapa en buffer för att läsa response body
			buf := new(bytes.Buffer)

			// Läs hela response body till buffern
			buf.ReadFrom(statsResp.Body)

			// Stäng body
			statsResp.Body.Close()

			// Skriv ut statistiken som en sträng
			fmt.Printf("[STATS] App3: %s\n", buf.String())
		}

		// Separator för tydligare logg
		fmt.Println("---------------------------------------")

		// Vänta 10 sekunder innan nästa mätning
		time.Sleep(10 * time.Second)
	}
}
