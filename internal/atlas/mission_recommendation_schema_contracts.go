package atlas

type atlasRecommendationEvidenceSchemaContracts struct {
	RefactoringRecommendations     string
	NextTrackDecision              string
	ConsumedRecommendationLedger   string
	TrackRegistry                  string
	CommandRunLedger               string
	CommandRunLedgerRollup         string
	RunLedgerCoverageCheck         string
	FinalResponseGates             string
	EvidenceValidationReport       string
	EvidenceSchemaRegistry         string
	EvidenceSchemaRegistryCoverage string
	SchemaHealthRepairPrompt       string
	ControlPlane                   []string
}

func defaultAtlasRecommendationEvidenceSchemaContracts() atlasRecommendationEvidenceSchemaContracts {
	contracts := atlasRecommendationEvidenceSchemaContracts{
		RefactoringRecommendations:     AOMissionRefactoringRecommendationsContract,
		NextTrackDecision:              AtlasRecommendationNextTrackDecisionContract,
		ConsumedRecommendationLedger:   AtlasConsumedRecommendationLedgerContract,
		TrackRegistry:                  AtlasRecommendationTrackRegistryContract,
		CommandRunLedger:               AtlasRecommendationCommandRunLedgerContract,
		CommandRunLedgerRollup:         AtlasRecommendationCommandRunLedgerRollupContract,
		RunLedgerCoverageCheck:         AtlasRecommendationRunLedgerCoverageCheckContract,
		FinalResponseGates:             AtlasRecommendationFinalResponseGatesContract,
		EvidenceValidationReport:       AtlasRecommendationEvidenceValidationReportContract,
		EvidenceSchemaRegistry:         AtlasRecommendationEvidenceSchemaRegistryContract,
		EvidenceSchemaRegistryCoverage: AtlasRecommendationEvidenceSchemaRegistryCoverageContract,
		SchemaHealthRepairPrompt:       AtlasSchemaHealthRepairPromptContract,
	}
	contracts.ControlPlane = []string{
		contracts.RefactoringRecommendations,
		contracts.NextTrackDecision,
		contracts.ConsumedRecommendationLedger,
		contracts.TrackRegistry,
		contracts.CommandRunLedger,
		contracts.CommandRunLedgerRollup,
		contracts.RunLedgerCoverageCheck,
		contracts.FinalResponseGates,
		contracts.EvidenceValidationReport,
		contracts.EvidenceSchemaRegistry,
		contracts.EvidenceSchemaRegistryCoverage,
		contracts.SchemaHealthRepairPrompt,
	}
	return contracts
}
