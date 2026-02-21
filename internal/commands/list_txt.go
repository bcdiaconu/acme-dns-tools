package commands

import (
	"acme-dns-tools/internal/cpanel"
	"fmt"
)

type ListTxtCommand struct{}

func (c *ListTxtCommand) ValidateArgs(args map[string]string) error {
	if args["domain"] == "" {
		return fmt.Errorf("domain is required")
	}
	return nil
}

func (c *ListTxtCommand) Execute(cpCfg *cpanel.CPanelConfig, args map[string]string) error {
	domain := args["domain"]
	key := args["key"] // Optional - if provided, filter by key

	records, err := cpCfg.ListTxtRecords(domain, key)
	if err != nil {
		return fmt.Errorf("failed to list TXT records: %w", err)
	}

	if len(records) == 0 {
		if key != "" {
			fmt.Printf("No TXT records found for key '%s' in domain '%s'\n", key, domain)
		} else {
			fmt.Printf("No TXT records found for domain '%s'\n", domain)
		}
		return nil
	}

	fmt.Printf("TXT records for domain '%s':\n", domain)
	for _, record := range records {
		if key == "" || record.Key == key {
			fmt.Printf("  Line: %-3d | Key: %-30s | Value: %s\n", record.Line, record.Key, record.Value)
		}
	}

	return nil
}

func (c *ListTxtCommand) Usage() string {
	return "list-txt --domain <domain> [--key <key>]"
}
