package atlas

import "fmt"

func BuildAtlasCovenantTicketSchemaAuthorityLedger(inputPath string) (AtlasCovenantTicketSchemaAuthorityLedger, error) {
	input, err := LoadJSON[AtlasCovenantTicketSchemaAuthorityLedgerInput](inputPath)
	if err != nil {
		return AtlasCovenantTicketSchemaAuthorityLedger{}, err
	}
	if err := ValidateAtlasCovenantTicketSchemaAuthorityLedgerInput(input); err != nil {
		return AtlasCovenantTicketSchemaAuthorityLedger{}, err
	}
	ledger := summarizeCovenantTicketSchemaAuthorityLedger(input.Entries)
	ledger.Schema = AtlasCovenantTicketSchemaAuthorityLedgerContract
	ledger.Status = "covenant_ticket_schema_authority_compatible"
	ledger.SourceInputPath = publicArtifactRef(inputPath)
	ledger.SourceInputDigest = digestValue(input)
	ledger.SchedulesWork = false
	ledger.ExecutesWork = false
	ledger.ApprovesWork = false
	ledger.ClaimsAuthorityAdvance = false
	ledger.RSIRemainsDenied = true
	if err := ValidateAtlasCovenantTicketSchemaAuthorityLedger(ledger); err != nil {
		return AtlasCovenantTicketSchemaAuthorityLedger{}, err
	}
	return ledger, nil
}

func ValidateAtlasCovenantTicketSchemaAuthorityLedgerInput(input AtlasCovenantTicketSchemaAuthorityLedgerInput) error {
	var errs []string
	requireContract(&errs, "covenant_ticket_schema_authority_ledger_input", input.Schema, AtlasCovenantTicketSchemaAuthorityLedgerInputContract)
	if input.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	validateCovenantTicketSchemaAuthorityLedgerEntries(&errs, input.Entries)
	validateNoAuthorityEffects(&errs, input.SchedulesWork, input.ExecutesWork, input.ApprovesWork, input.ClaimsAuthorityAdvance, input.RSIRemainsDenied)
	return joinErrors(errs)
}

func ValidateAtlasCovenantTicketSchemaAuthorityLedger(ledger AtlasCovenantTicketSchemaAuthorityLedger) error {
	var errs []string
	requireContract(&errs, "covenant_ticket_schema_authority_ledger", ledger.Schema, AtlasCovenantTicketSchemaAuthorityLedgerContract)
	if ledger.Status != "covenant_ticket_schema_authority_compatible" {
		errs = append(errs, "status must be covenant_ticket_schema_authority_compatible")
	}
	requireField(&errs, "source_input_path", ledger.SourceInputPath)
	checkPublicPath(&errs, "source_input_path", ledger.SourceInputPath, true)
	validateRejectedTicketDigest(&errs, "source_input_digest", ledger.SourceInputDigest)
	validateCovenantTicketSchemaAuthorityLedgerEntries(&errs, ledger.Entries)
	expected := summarizeCovenantTicketSchemaAuthorityLedger(ledger.Entries)
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

func validateCovenantTicketSchemaAuthorityLedgerEntries(errs *[]string, entries []AtlasCovenantTicketSchemaAuthorityLedgerEntry) {
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

func summarizeCovenantTicketSchemaAuthorityLedger(entries []AtlasCovenantTicketSchemaAuthorityLedgerEntry) AtlasCovenantTicketSchemaAuthorityLedger {
	ledger := AtlasCovenantTicketSchemaAuthorityLedger{
		EntryCount:           len(entries),
		AllEntriesCompatible: true,
		Entries:              append([]AtlasCovenantTicketSchemaAuthorityLedgerEntry(nil), entries...),
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
