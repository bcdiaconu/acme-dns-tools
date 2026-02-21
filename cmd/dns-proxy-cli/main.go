package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"acme-dns-tools/internal/commands"
	"acme-dns-tools/internal/cpanel"
)

func loadCPanelConfig(path string) map[string]string {
	cfg := make(map[string]string)
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			cfg[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return cfg
}

func main() {
	ignoreErrors := false
	filteredArgs := []string{}
	for _, arg := range os.Args[1:] {
		if arg == "-i" || arg == "--ignore-errors" {
			ignoreErrors = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	if len(filteredArgs) < 1 {
		fmt.Println("Usage: dns-proxy-cli [-i|--ignore-errors] <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  set-txt --domain <domain> --key <key> --value <value>")
		fmt.Println("  delete-txt --domain <domain> --key <key> --value <value>")
		fmt.Println("  edit-txt --domain <domain> --key <key> --old-value <old-value> --new-value <new-value>")
		fmt.Println("  list-txt --domain <domain> [--key <key>]")
		os.Exit(1)
	}

	subcmd := filteredArgs[0]

	// Create command factory and get command
	factory := commands.NewCommandFactory()
	cmd, err := factory.CreateCommand(subcmd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		if ignoreErrors {
			os.Exit(0)
		}
		os.Exit(1)
	}

	// Parse arguments based on command
	args := parseCommandArgs(subcmd, filteredArgs[1:])

	// Validate arguments
	if err := cmd.ValidateArgs(args); err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Usage: %s\n", cmd.Usage())
		if ignoreErrors {
			os.Exit(0)
		}
		os.Exit(1)
	}

	// Load cPanel config
	cfg := loadCPanelConfig("/etc/acme-dns-tools/dns-proxy-cli.conf")
	cpCfg, err := cpanel.NewCPanelConfig(cfg)
	if err != nil {
		log.Printf("%v", err)
		if ignoreErrors {
			os.Exit(0)
		}
		os.Exit(1)
	}

	// Execute command
	if err := cmd.Execute(cpCfg, args); err != nil {
		log.Printf("%v", err)
		if ignoreErrors {
			os.Exit(0)
		}
		os.Exit(1)
	}
}

func parseCommandArgs(subcmd string, args []string) map[string]string {
	var cmdFlags *flag.FlagSet

	switch subcmd {
	case "set-txt", "delete-txt":
		cmdFlags = flag.NewFlagSet(subcmd, flag.ExitOnError)
		domain := cmdFlags.String("domain", "", "Domain name")
		key := cmdFlags.String("key", "", "TXT record key")
		value := cmdFlags.String("value", "", "TXT record value")

		cmdFlags.Parse(args)

		return map[string]string{
			"domain": *domain,
			"key":    *key,
			"value":  *value,
		}
	case "edit-txt":
		cmdFlags = flag.NewFlagSet(subcmd, flag.ExitOnError)
		domain := cmdFlags.String("domain", "", "Domain name")
		key := cmdFlags.String("key", "", "TXT record key")
		oldValue := cmdFlags.String("old-value", "", "Current TXT record value")
		newValue := cmdFlags.String("new-value", "", "New TXT record value")

		cmdFlags.Parse(args)

		return map[string]string{
			"domain":    *domain,
			"key":       *key,
			"old-value": *oldValue,
			"new-value": *newValue,
		}
	case "list-txt":
		cmdFlags = flag.NewFlagSet(subcmd, flag.ExitOnError)
		domain := cmdFlags.String("domain", "", "Domain name")
		key := cmdFlags.String("key", "", "TXT record key filter (optional)")

		cmdFlags.Parse(args)

		return map[string]string{
			"domain": *domain,
			"key":    *key,
		}
	default:
		return nil
	}
}
