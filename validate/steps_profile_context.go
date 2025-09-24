package validate

import (
	"strings"

	"github.com/AgendoCerto/lib-bot/io"
)

// ProfileContextStep valida a configuração de contexto do profile
type ProfileContextStep struct{}

// NewProfileContextStep cria novo validador de contexto do profile
func NewProfileContextStep() *ProfileContextStep {
	return &ProfileContextStep{}
}

// ValidateDesign valida o contexto do profile no design
func (v *ProfileContextStep) ValidateDesign(design io.DesignDoc) []Issue {
	var issues []Issue

	// Verificar se há profile definido
	if len(design.Profile.Context) == 0 && len(design.Profile.Variables) == 0 {
		return issues // Profile vazio é válido
	}

	// Validar cada variável definida no contexto
	for varName, profileVar := range design.Profile.Context {
		path := "profile.context." + varName

		// Validar nome da variável
		if varName == "" {
			issues = append(issues, Issue{
				Code:     "profile_empty_variable_name",
				Severity: "error",
				Msg:      "Variable name cannot be empty",
				Path:     path,
			})
			continue
		}

		// Validar caracteres válidos no nome da variável
		if !isValidVariableName(varName) {
			issues = append(issues, Issue{
				Code:     "profile_invalid_variable_name",
				Severity: "error",
				Msg:      "Variable name must contain only letters, numbers, and underscore",
				Path:     path,
			})
		}

		// Validar tipo da variável
		if profileVar.Type == "" {
			issues = append(issues, Issue{
				Code:     "profile_missing_variable_type",
				Severity: "error",
				Msg:      "Variable type is required",
				Path:     path + ".type",
			})
		} else if !isValidVariableType(profileVar.Type) {
			issues = append(issues, Issue{
				Code:     "profile_invalid_variable_type",
				Severity: "error",
				Msg:      "Invalid variable type: " + profileVar.Type + ". Valid types: string, number, int, float, boolean, bool",
				Path:     path + ".type",
			})
		}

		// Validar valor default se especificado
		if profileVar.Default != "" {
			if !isValidDefaultForType(profileVar.Default, profileVar.Type) {
				issues = append(issues, Issue{
					Code:     "profile_invalid_default_value",
					Severity: "warning",
					Msg:      "Default value '" + profileVar.Default + "' may not be compatible with type '" + profileVar.Type + "'",
					Path:     path + ".default",
				})
			}
		}

		// Validar se variáveis obrigatórias têm valor default ou valor atual
		if profileVar.Required {
			hasDefault := profileVar.Default != ""
			hasCurrentValue := design.Profile.Variables != nil && design.Profile.Variables[varName] != nil

			if !hasDefault && !hasCurrentValue {
				issues = append(issues, Issue{
					Code:     "profile_required_without_value",
					Severity: "warning",
					Msg:      "Required variable should have a default value or current value in profile.variables",
					Path:     path,
				})
			}
		}
	}

	// Validar consistência entre profile.variables e profile.context
	if design.Profile.Variables != nil {
		for varName := range design.Profile.Variables {
			if _, hasDefinition := design.Profile.Context[varName]; !hasDefinition {
				issues = append(issues, Issue{
					Code:     "profile_variable_without_definition",
					Severity: "warning",
					Msg:      "Variable '" + varName + "' has value but no definition in profile.context",
					Path:     "profile.variables." + varName,
				})
			}
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

// isValidVariableType verifica se o tipo da variável é válido
func isValidVariableType(varType string) bool {
	validTypes := []string{
		"string", "str",
		"number", "num", "int", "integer", "float", "double",
		"boolean", "bool",
	}

	for _, valid := range validTypes {
		if strings.ToLower(varType) == valid {
			return true
		}
	}

	return false
}

// isValidDefaultForType verifica se o valor default é compatível com o tipo
func isValidDefaultForType(defaultValue, varType string) bool {
	switch strings.ToLower(varType) {
	case "boolean", "bool":
		return strings.ToLower(defaultValue) == "true" || strings.ToLower(defaultValue) == "false"
	case "number", "num", "int", "integer", "float", "double":
		// Para simplificar, aceita qualquer valor não-vazio para números
		// Em uma implementação mais robusta, poderia fazer parsing
		return defaultValue != ""
	case "string", "str":
		return true // Qualquer valor é válido para string
	default:
		return true // Para tipos desconhecidos, aceita
	}
}
