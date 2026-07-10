package atlas

import "fmt"

func BuildAtlasCommandTicketSchemaCompatibilityLedger(inputPath string) (AtlasCommandTicketSchemaCompatibilityLedger, error) {
	input, err := LoadJSON[AtlasCommandTicketSchemaCompatibilityLedgerInput](inputPath)
	if err != nil {
		return AtlasCommandTicketSchemaCompatibilityLedger{}, err
	}
	if err := ValidateAtlasCommandTicketSchemaCompatibilityLedgerInput(input); err != nil {
		return AtlasCommandTicketSchemaCompatibilityLedger{}, err
	}
	ledger := summarizeCommandTicketSchemaCompatibilityLedger(input.Entries)
	ledger.Schema = AtlasCommandTicketSchemaCompatibilityLedgerContract
	ledger.Status = "command_ticket_schema_compatible"
	ledger.SourceInputPath = publicArtifactRef(inputPath)
	ledger.SourceInputDigest = digestValue(input)
	ledger.SchedulesWork = false
	ledger.ExecutesWork = false
	ledger.ApprovesWork = false
	ledger.ClaimsAuthorityAdvance = false
	ledger.RSIRemainsDenied = true
	if err := ValidateAtlasCommandTicketSchemaCompatibilityLedger(ledger); err != nil {
		return AtlasCommandTicketSchemaCompatibilityLedger{}, err
	}
	return ledger, nil
}

func ValidateAtlasCommandTicketSchemaCompatibilityLedgerInput(input AtlasCommandTicketSchemaCompatibilityLedgerInput) error {
	var errs []string
	requireContract(&errs, "command_ticket_schema_compatibility_ledger_input", input.Schema, AtlasCommandTicketSchemaCompatibilityLedgerInputContract)
	if input.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	validateCommandTicketSchemaCompatibilityLedgerEntries(&errs, input.Entries)
	validateNoAuthorityEffects(&errs, input.SchedulesWork, input.ExecutesWork, input.ApprovesWork, input.ClaimsAuthorityAdvance, input.RSIRemainsDenied)
	return joinErrors(errs)
}

func ValidateAtlasCommandTicketSchemaCompatibilityLedger(ledger AtlasCommandTicketSchemaCompatibilityLedger) error {
	var errs []string
	requireContract(&errs, "command_ticket_schema_compatibility_ledger", ledger.Schema, AtlasCommandTicketSchemaCompatibilityLedgerContract)
	if ledger.Status != "command_ticket_schema_compatible" {
		errs = append(errs, "status must be command_ticket_schema_compatible")
	}
	requireField(&errs, "source_input_path", ledger.SourceInputPath)
	checkPublicPath(&errs, "source_input_path", ledger.SourceInputPath, true)
	validateRejectedTicketDigest(&errs, "source_input_digest", ledger.SourceInputDigest)
	validateCommandTicketSchemaCompatibilityLedgerEntries(&errs, ledger.Entries)
	expected := summarizeCommandTicketSchemaCompatibilityLedger(ledger.Entries)
	if ledger.EntryCount != expected.EntryCount {
		errs = append(errs, "entry_count must match entries")
	}
	if ledger.CompatibleEntryCount != expected.CompatibleEntryCount {
		errs = append(errs, "compatible_entry_count must match entries")
	}
	if ledger.AllEntriesCompatible != expected.AllEntriesCompatible {
		errs = append(errs, "all_entries_compatible must match entries")
	}
	validateNoAuthorityEffects(&errs, ledger.SchedulesWork, ledger.ExecutesWork, ledger.ApprovesWork, ledger.ClaimsAuthorityAdvance, ledger.RSIRemainsDenied)
	return joinErrors(errs)
}

func validateCommandTicketSchemaCompatibilityLedgerEntries(errs *[]string, entries []AtlasCommandTicketSchemaCompatibilityLedgerEntry) {
	if len(entries) == 0 {
		*errs = append(*errs, "entries must not be empty")
	}
	for i, item := range entries {
		prefix := fmt.Sprintf("entries[%d]", i)
		requireField(errs, prefix+".id", item.ID)
		requireField(errs, prefix+".producer_schema", item.ProducerSchema)
		requireField(errs, prefix+".consumer_schema", item.ConsumerSchema)
		if len(item.RequiredFields) == 0 {
			*errs = append(*errs, prefix+".required_fields must not be empty")
		}
		for j, field := range item.RequiredFields {
			requireField(errs, fmt.Sprintf("%s.required_fields[%d]", prefix, j), field)
		}
		if !item.Compatible {
			*errs = append(*errs, prefix+".compatible must be true for compatibility ledger")
		}
	}
}

func summarizeCommandTicketSchemaCompatibilityLedger(entries []AtlasCommandTicketSchemaCompatibilityLedgerEntry) AtlasCommandTicketSchemaCompatibilityLedger {
	ledger := AtlasCommandTicketSchemaCompatibilityLedger{
		EntryCount:           len(entries),
		AllEntriesCompatible: true,
		Entries:              append([]AtlasCommandTicketSchemaCompatibilityLedgerEntry(nil), entries...),
	}
	for _, item := range entries {
		if item.Compatible {
			ledger.CompatibleEntryCount++
		} else {
			ledger.AllEntriesCompatible = false
		}
	}
	return ledger
}
