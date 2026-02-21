package commands

import (
	"errors"
	"fmt"

	"acme-dns-tools/internal/cpanel"
)

// SetTxtCommand implements the set-txt command
type SetTxtCommand struct{}

func (c *SetTxtCommand) Execute(cpCfg *cpanel.CPanelConfig, args map[string]string) error {
	domain := args["domain"]
	key := args["key"]
	value := args["value"]

	err := cpCfg.CreateTxtRecord(domain, key, value)
	if err != nil {
		return fmt.Errorf("failed to set TXT record: %w", err)
	}

	fmt.Println("TXT record set successfully.")
	return nil
}

func (c *SetTxtCommand) ValidateArgs(args map[string]string) error {
	if args["domain"] == "" {
		return errors.New("--domain is required")
	}
	if args["key"] == "" {
		return errors.New("--key is required")
	}
	if args["value"] == "" {
		return errors.New("--value is required")
	}
	return nil
}

func (c *SetTxtCommand) Usage() string {
	return "set-txt --domain <domain> --key <key> --value <value>"
}
