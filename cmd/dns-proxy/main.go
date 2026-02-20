package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func loadAPIKey(path string) string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer file.Close()
	var apiKey string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "API_KEY" {
			apiKey = strings.TrimSpace(parts[1])
		}
	}
	if apiKey == "" {
		log.Fatal("API_KEY not found in config file")
	}
	return apiKey
}

func main() {
	apiKey := loadAPIKey("/etc/acme-dns-tools/dns-proxy.conf")

	http.HandleFunc("/set_txt", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		expected := "Bearer " + apiKey
		if authHeader != expected {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			Domain string `json:"domain"`
			Key    string `json:"key"`
			Value  string `json:"value"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil || req.Domain == "" || req.Key == "" || req.Value == "" {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		cmd := exec.Command("/usr/local/bin/dns-proxy-cli", "set-txt", "--domain", req.Domain, "--key", req.Key, "--value", req.Value)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("dns-proxy-cli error: %v, output: %s", err, string(output))
			http.Error(w, string(output), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("TXT record set"))
	})

	log.Println("dns-proxy API listening on :5000...")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
