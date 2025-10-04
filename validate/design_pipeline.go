package validate

import (
	"github.com/AgendoCerto/lib-bot/io"
)

// DesignValidator interface para validação de design completo
type DesignValidator interface {
	ValidateDesign(design io.DesignDoc) []Issue
}

// DesignValidationPipeline pipeline para validação de design completo
type DesignValidationPipeline struct {
	validators []DesignValidator
}

// NewDesignValidationPipeline cria pipeline de validação de design
func NewDesignValidationPipeline() *DesignValidationPipeline {
	return &DesignValidationPipeline{
		validators: []DesignValidator{
			NewAdapterComplianceStep(),
			NewDocumentationComplianceStep(),
			NewComponentBehaviorStep(), // CRÍTICO: Validação de behaviors permitidos por componente
			NewOutputMappingStep(),     // CRÍTICO: Validação de mapeamento output-to-ID
			NewLiquidLengthStep(),      // CRÍTICO: Validação de limites considerando templates Liquid
			NewProfileContextStep(),    // NOVO: Validação de profile context
			NewWhatsAppLimitsStep(),    // NOVO: Validação de limites WhatsApp Business API
		},
	}
}

// ValidateDesign executa validação completa do design
func (p *DesignValidationPipeline) ValidateDesign(design io.DesignDoc) []Issue {
	var allIssues []Issue

	for _, validator := range p.validators {
		issues := validator.ValidateDesign(design)
		allIssues = append(allIssues, issues...)
	}

	return allIssues
}

// AddValidator adiciona novo validador ao pipeline
func (p *DesignValidationPipeline) AddValidator(validator DesignValidator) {
	p.validators = append(p.validators, validator)
}
