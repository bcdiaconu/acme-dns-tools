package cpanel

import (
	"acme-dns-tools/internal/cpanel/commands"
	"acme-dns-tools/internal/cpanel/queries"
)

// CPanelService provides CQRS interface for cPanel operations
type CPanelService struct {
	config         *CPanelConfig
	commandHandler commands.CommandHandler
	queryHandler   queries.QueryHandler
}

// NewCPanelService creates a new cPanel service with CQRS architecture
func NewCPanelService(config *CPanelConfig) *CPanelService {
	return &CPanelService{
		config:         config,
		commandHandler: commands.NewCPanelCommandHandler(),
		queryHandler:   queries.NewCPanelQueryHandler(),
	}
}

// Command methods (Write operations)

// CreateTxtRecord creates a new TXT record
func (s *CPanelService) CreateTxtRecord(domain, key, value string) error {
	cmd := &commands.CreateTxtRecordCommand{
		Request: commands.CreateTxtRecordRequest{
			Domain: domain,
			Key:    key,
			Value:  value,
		},
	}
	return s.commandHandler.HandleCreate(cmd)
}

// DeleteTxtRecord deletes a TXT record
func (s *CPanelService) DeleteTxtRecord(domain, key, value string) error {
	cmd := &commands.DeleteTxtRecordCommand{
		Request: commands.DeleteTxtRecordRequest{
			Domain: domain,
			Key:    key,
			Value:  value,
		},
	}
	return s.commandHandler.HandleDelete(cmd)
}

// EditTxtRecord edits a TXT record
func (s *CPanelService) EditTxtRecord(domain, key, oldValue, newValue string) error {
	cmd := &commands.EditTxtRecordCommand{
		Request: commands.EditTxtRecordRequest{
			Domain:   domain,
			Key:      key,
			OldValue: oldValue,
			NewValue: newValue,
		},
	}
	return s.commandHandler.HandleEdit(cmd)
}

// Query methods (Read operations)

// ListTxtRecords lists TXT records for a domain with optional key filter
func (s *CPanelService) ListTxtRecords(domain, keyFilter string) ([]TxtRecord, error) {
	query := &queries.ListTxtRecordsQuery{
		Request: queries.ListTxtRecordsRequest{
			Domain:    domain,
			KeyFilter: keyFilter,
		},
	}
	
	// Convert queries.TxtRecord to cpanel.TxtRecord
	queryRecords, err := s.queryHandler.HandleList(query)
	if err != nil {
		return nil, err
	}
	
	// Convert to cpanel package TxtRecord format
	var records []TxtRecord
	for _, qr := range queryRecords {
		records = append(records, TxtRecord{
			Line:  qr.Line,
			Key:   qr.Key,
			Value: qr.Value,
			Name:  qr.Name,
		})
	}
	
	return records, nil
}
