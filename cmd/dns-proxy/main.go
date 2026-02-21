package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"dns-proxy/internal/api"
	"dns-proxy/internal/config"
	"dns-proxy/internal/cpanel"
)

func loadToken(path string) string {
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
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "DNS_RESOLVER_API_TOKEN" {
			apiKey = strings.TrimSpace(parts[1])
		}
	}
	if apiKey == "" {
		log.Fatal("DNS_RESOLVER_API_TOKEN not found in config file")
	}
	return apiKey
}

func main() {
	apiToken := loadToken("/etc/acme-dns-tools/dns-proxy-api.conf")

	// Adapter that implements api.TxtRecordSetter by using internal cPanel client
	cfgMap := config.LoadConfig("/etc/acme-dns-tools/dns-proxy-cli.conf")
	cpCfg, err := cpanel.NewCPanelConfig(cfgMap)
	if err != nil {
		log.Fatalf("failed to load cPanel config: %v", err)
	}

	type internalSetter struct{
		cp *cpanel.CPanelConfig
	}

	func (s *internalSetter) CreateTxtRecord(domain, key, value string) error {
		return s.cp.CreateTxtRecord(domain, key, value)
	}

	setter := &internalSetter{cp: cpCfg}
	http.HandleFunc("/set_txt", api.SetTxtHandler(apiToken, setter))

	log.Println("dns-proxy API listening on :5000...")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
