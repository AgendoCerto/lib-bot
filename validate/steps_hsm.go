package validate

import (
	"github.com/AgendoCerto/lib-bot/adapter"
	"github.com/AgendoCerto/lib-bot/component"
)

// HSMValidationStep valida configurações básicas de HSM
type HSMValidationStep struct{}

// NewHSMValidationStep cria um novo step de validação HSM simplificado
func NewHSMValidationStep() *HSMValidationStep {
	return &HSMValidationStep{}
}

func (s *HSMValidationStep) Check(spec component.ComponentSpec, caps adapter.Capabilities, path string) []Issue {
	var issues []Issue

	// Se não há HSM, não há o que validar
	if spec.HSM == nil {
		return issues
	}

	// Verifica se adapter suporta HSM
	if !caps.SupportsHSM {
		issues = append(issues, Issue{
			Code: "hsm.adapter.unsupported", Severity: Err,
			Path: path + ".hsm", Msg: "HSM not supported by adapter",
		})
		return issues
	}

	hsmTemplate := spec.HSM

	// Validação básica: Nome do template é obrigatório
	if err := hsmTemplate.Validate(); err != nil {
		issues = append(issues, Issue{
			Code: "hsm.template.invalid", Severity: Err,
			Path: path + ".hsm.name",
			Msg:  "HSM template validation failed: " + err.Error(),
		})
	}

	return issues
}
