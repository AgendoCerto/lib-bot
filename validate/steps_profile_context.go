package validate

import (
	"github.com/AgendoCerto/lib-bot/io"
)

// ProfileContextStep valida a configuração de variáveis do bot
type ProfileContextStep struct{}

// NewProfileContextStep cria novo validador de variáveis do bot
func NewProfileContextStep() *ProfileContextStep {
	return &ProfileContextStep{}
}

// ValidateDesign valida as variáveis do bot no design
func (v *ProfileContextStep) ValidateDesign(design io.DesignDoc) []Issue {
	var issues []Issue

	// Verificar se há variables definido
	if len(design.Variables.Context) == 0 && len(design.Variables.State) == 0 && len(design.Variables.Global) == 0 {
		return issues // Variables vazio é válido
	}

	// Validar context (array de strings)
	for _, varName := range design.Variables.Context {
		path := "variables.context[" + varName + "]"

		// Validar nome da variável
		if varName == "" {
			issues = append(issues, Issue{
				Code:     "variables_empty_context_name",
				Severity: "error",
				Msg:      "Context variable name cannot be empty",
				Path:     path,
			})
			continue
		}

		// Validar caracteres válidos no nome da variável
		if !isValidVariableName(varName) {
			issues = append(issues, Issue{
				Code:     "variables_invalid_context_name",
				Severity: "error",
				Msg:      "Variable name must contain only letters, numbers, and underscore",
				Path:     path,
			})
		}
	}

	// Validar state (array de strings)
	for _, varName := range design.Variables.State {
		path := "variables.state[" + varName + "]"

		// Validar nome da variável
		if varName == "" {
			issues = append(issues, Issue{
				Code:     "variables_empty_state_name",
				Severity: "error",
				Msg:      "State variable name cannot be empty",
				Path:     path,
			})
			continue
		}

		// Validar caracteres válidos no nome da variável
		if !isValidVariableName(varName) {
			issues = append(issues, Issue{
				Code:     "variables_invalid_state_name",
				Severity: "error",
				Msg:      "Variable name must contain only letters, numbers, and underscore",
				Path:     path,
			})
		}
	}

	// Validar global (objeto chave-valor)
	for varName := range design.Variables.Global {
		path := "variables.global." + varName

		// Validar nome da variável
		if varName == "" {
			issues = append(issues, Issue{
				Code:     "variables_empty_global_name",
				Severity: "error",
				Msg:      "Global variable name cannot be empty",
				Path:     path,
			})
			continue
		}

		// Validar caracteres válidos no nome da variável
		if !isValidVariableName(varName) {
			issues = append(issues, Issue{
				Code:     "variables_invalid_global_name",
				Severity: "error",
				Msg:      "Variable name must contain only letters, numbers, and underscore",
				Path:     path,
			})
		}
	}

	return issues
}

// isValidVariableName verifica se o nome da variável é válido
func isValidVariableName(name string) bool {
	if name == "" {
		return false
	}

	// Não pode começar com número
	if name[0] >= '0' && name[0] <= '9' {
		return false
	}

	// Só pode conter letras, números e underscore
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}

	return true
}
