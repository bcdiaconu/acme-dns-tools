package commands

import "acme-dns-tools/internal/cpanel"

// Command represents a DNS operation command
type Command interface {
	Execute(cpCfg *cpanel.CPanelConfig, args map[string]string) error
	ValidateArgs(args map[string]string) error
	Usage() string
}

// CommandFactory creates command instances
type CommandFactory interface {
	CreateCommand(name string) (Command, error)
}

// DefaultCommandFactory implements CommandFactory
type DefaultCommandFactory struct{}

func NewCommandFactory() CommandFactory {
	return &DefaultCommandFactory{}
}

func (f *DefaultCommandFactory) CreateCommand(name string) (Command, error) {
	switch name {
	case "set-txt":
		return &SetTxtCommand{}, nil
	case "delete-txt":
		return &DeleteTxtCommand{}, nil
	case "edit-txt":
		return &EditTxtCommand{}, nil
	case "list-txt":
		return &ListTxtCommand{}, nil
	default:
		return nil, &UnknownCommandError{Command: name}
	}
}

// UnknownCommandError represents an error for unknown commands
type UnknownCommandError struct {
	Command string
}

func (e *UnknownCommandError) Error() string {
	return "unknown command: " + e.Command
}
