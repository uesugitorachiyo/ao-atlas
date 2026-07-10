package atlas

import "fmt"

func BuildAtlasBlueprintTicketSchemaCompatibilityLedger(inputPath string) (AtlasBlueprintTicketSchemaCompatibilityLedger, error) {
	input, err := LoadJSON[AtlasBlueprintTicketSchemaCompatibilityLedgerInput](inputPath)
	if err != nil {
		return AtlasBlueprintTicketSchemaCompatibilityLedger{}, err
	}
	if err := ValidateAtlasBlueprintTicketSchemaCompatibilityLedgerInput(input); err != nil {
		return AtlasBlueprintTicketSchemaCompatibilityLedger{}, err
	}
	ledger := summarizeBlueprintTicketSchemaCompatibilityLedger(input.Entries)
	ledger.Schema = AtlasBlueprintTicketSchemaCompatibilityLedgerContract
	ledger.Status = "blueprint_ticket_schema_compatible"
	ledger.SourceInputPath = publicArtifactRef(inputPath)
	ledger.SourceInputDigest = digestValue(input)
	ledger.SchedulesWork = false
	ledger.ExecutesWork = false
	ledger.ApprovesWork = false
	ledger.ClaimsAuthorityAdvance = false
	ledger.RSIRemainsDenied = true
	if err := ValidateAtlasBlueprintTicketSchemaCompatibilityLedger(ledger); err != nil {
		return AtlasBlueprintTicketSchemaCompatibilityLedger{}, err
	}
	return ledger, nil
}

func ValidateAtlasBlueprintTicketSchemaCompatibilityLedgerInput(input AtlasBlueprintTicketSchemaCompatibilityLedgerInput) error {
	var errs []string
	requireContract(&errs, "blueprint_ticket_schema_compatibility_ledger_input", input.Schema, AtlasBlueprintTicketSchemaCompatibilityLedgerInputContract)
	if input.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	validateBlueprintTicketSchemaCompatibilityLedgerEntries(&errs, input.Entries)
	validateNoAuthorityEffects(&errs, input.SchedulesWork, input.ExecutesWork, input.ApprovesWork, input.ClaimsAuthorityAdvance, input.RSIRemainsDenied)
	return joinErrors(errs)
}

func ValidateAtlasBlueprintTicketSchemaCompatibilityLedger(ledger AtlasBlueprintTicketSchemaCompatibilityLedger) error {
	var errs []string
	requireContract(&errs, "blueprint_ticket_schema_compatibility_ledger", ledger.Schema, AtlasBlueprintTicketSchemaCompatibilityLedgerContract)
	if ledger.Status != "blueprint_ticket_schema_compatible" {
		errs = append(errs, "status must be blueprint_ticket_schema_compatible")
	}
	requireField(&errs, "source_input_path", ledger.SourceInputPath)
	checkPublicPath(&errs, "source_input_path", ledger.SourceInputPath, true)
	validateRejectedTicketDigest(&errs, "source_input_digest", ledger.SourceInputDigest)
	validateBlueprintTicketSchemaCompatibilityLedgerEntries(&errs, ledger.Entries)
	expected := summarizeBlueprintTicketSchemaCompatibilityLedger(ledger.Entries)
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

func validateBlueprintTicketSchemaCompatibilityLedgerEntries(errs *[]string, entries []AtlasBlueprintTicketSchemaCompatibilityLedgerEntry) {
	if len(entries) == 0 {
		*errs = append(*errs, "entries must not be empty")
	}
	for i, item := range entries {
		prefix := fmt.Sprintf("entries[%d]", i)
		requireField(errs, prefix+".id", item.ID)
		requireField(errs, prefix+".blueprint_ticket_schema", item.BlueprintTicketSchema)
		requireField(errs, prefix+".covenant_ticket_schema", item.CovenantTicketSchema)
		requireField(errs, prefix+".command_ticket_schema", item.CommandTicketSchema)
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

func summarizeBlueprintTicketSchemaCompatibilityLedger(entries []AtlasBlueprintTicketSchemaCompatibilityLedgerEntry) AtlasBlueprintTicketSchemaCompatibilityLedger {
	ledger := AtlasBlueprintTicketSchemaCompatibilityLedger{
		EntryCount:           len(entries),
		AllEntriesCompatible: true,
		Entries:              append([]AtlasBlueprintTicketSchemaCompatibilityLedgerEntry(nil), entries...),
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
