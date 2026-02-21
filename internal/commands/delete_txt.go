package commands

import (
	"errors"
	"fmt"

	"acme-dns-tools/internal/cpanel"
)

// DeleteTxtCommand implements the delete-txt command
type DeleteTxtCommand struct{}

func (c *DeleteTxtCommand) Execute(cpCfg *cpanel.CPanelConfig, args map[string]string) error {
	domain := args["domain"]
	key := args["key"]
	value := args["value"]

	err := cpCfg.DeleteTxtRecord(domain, key, value)
	if err != nil {
		return fmt.Errorf("failed to delete TXT record: %w", err)
	}

	fmt.Println("TXT record deleted successfully.")
	return nil
}

func (c *DeleteTxtCommand) ValidateArgs(args map[string]string) error {
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

func (c *DeleteTxtCommand) Usage() string {
	return "delete-txt --domain <domain> --key <key> --value <value>"
}
