package main

import (
	"acme-dns-tools/internal/api"
	"acme-dns-tools/internal/config"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

const configPath = "/etc/acme-dns-tools/dns-proxy-api.conf"
const defaultCertsBaseDir = "/etc/letsencrypt/live"

func main() {
	cfg := config.LoadConfig(configPath)

	// --- DNS management API key (existing) ---
	apiKey := cfg["API_KEY"]
	if apiKey == "" {
		log.Fatal("API_KEY not found in config file")
	}

	// --- Cert serving: Bearer token ---
	certBearerToken := cfg["CERT_BEARER_TOKEN"]
	if certBearerToken == "" {
		log.Fatal("CERT_BEARER_TOKEN not found in config file")
	}

	// --- Cert serving: DNS allowlist (comma-separated hostnames, FCrDNS) ---
	certDNSAllowlistRaw := cfg["CERT_DNS_ALLOWLIST"]
	if certDNSAllowlistRaw == "" {
		log.Fatal("CERT_DNS_ALLOWLIST not found in config file")
	}
	var certDNSAllowlist []string
	for _, h := range strings.Split(certDNSAllowlistRaw, ",") {
		h = strings.TrimSpace(h)
		if h != "" {
			certDNSAllowlist = append(certDNSAllowlist, h)
		}
	}

	// --- Cert serving: base directory (optional, defaults to letsencrypt live) ---
	certsBaseDir := cfg["CERT_BASE_DIR"]
	if certsBaseDir == "" {
		certsBaseDir = defaultCertsBaseDir
	}

	// --- TLS (optional) ---
	tlsCert := cfg["TLS_CERT"]
	tlsKey := cfg["TLS_KEY"]

	// --- /set_txt handler (existing) ---
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

	// --- /certs/ handler (new: pull-based cert serving) ---
	http.Handle("/certs/", api.CertsHandler(certBearerToken, certDNSAllowlist, certsBaseDir))

	if tlsCert != "" && tlsKey != "" {
		log.Println("dns-proxy API listening on :5000 (TLS)...")
		log.Fatal(http.ListenAndServeTLS(":5000", tlsCert, tlsKey, nil))
	} else {
		log.Println("dns-proxy API listening on :5000 (plain HTTP)...")
		log.Fatal(http.ListenAndServe(":5000", nil))
	}
}
