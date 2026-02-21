package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"dns-proxy/internal/api"
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

	// Adapter that implements api.TxtRecordSetter by calling the CLI
	type cliSetter struct{}

	func (c *cliSetter) CreateTxtRecord(domain, key, value string) error {
		cmd := exec.Command("/usr/local/bin/dns-proxy-cli", "set-txt", "--domain", domain, "--key", key, "--value", value)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("dns-proxy-cli error: %v, output: %s", err, string(output))
		}
		return nil
	}

	setter := &cliSetter{}
	http.HandleFunc("/set_txt", api.SetTxtHandler(apiToken, setter))

	log.Println("dns-proxy API listening on :5000...")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
