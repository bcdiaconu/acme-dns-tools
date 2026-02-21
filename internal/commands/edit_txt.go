package commands

import (
	"acme-dns-tools/internal/cpanel"
	"errors"
)

// EditTxtCommand implements the edit-txt command
type EditTxtCommand struct{}

func (c *EditTxtCommand) Execute(cpCfg *cpanel.CPanelConfig, args map[string]string) error {
	domain := args["domain"]
	key := args["key"]
	oldValue := args["old-value"]
	newValue := args["new-value"]

	return cpCfg.EditTxtRecord(domain, key, oldValue, newValue)
}

func (c *EditTxtCommand) ValidateArgs(args map[string]string) error {
	if args["domain"] == "" {
		return errors.New("domain is required")
	}
	if args["key"] == "" {
		return errors.New("key is required")
	}
	if args["old-value"] == "" {
		return errors.New("old-value is required")
	}
	if args["new-value"] == "" {
		return errors.New("new-value is required")
	}
	return nil
}

func (c *EditTxtCommand) Usage() string {
	return "edit-txt --domain <domain> --key <key> --old-value <old-value> --new-value <new-value>"
}
