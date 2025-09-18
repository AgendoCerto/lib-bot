// Package persistence define estruturas para configuração de persistência de dados
package persistence

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Scope define onde a informação será persistida
type Scope string

const (
	ScopeContext Scope = "context" // Dados temporários do fluxo atual
	ScopeProfile Scope = "profile" // Dados persistentes do usuário
)

// SanitizationType define tipos de sanitização predefinidas
type SanitizationType string

const (
	// Extratores de números
	SanitizeNumbersOnly  SanitizationType = "numbers_only" // Extrai apenas números
	SanitizeLettersOnly  SanitizationType = "letters_only" // Extrai apenas letras
	SanitizeAlphanumeric SanitizationType = "alphanumeric" // Extrai letras e números

	// Formatadores específicos
	SanitizeCPF   SanitizationType = "cpf"   // Extrai e formata CPF
	SanitizeCEP   SanitizationType = "cep"   // Extrai e formata CEP
	SanitizePhone SanitizationType = "phone" // Extrai e formata telefone brasileiro

	// Monetário
	SanitizeBRL SanitizationType = "monetary_brl" // Extrai e formata valor monetário em reais

	// Normalizadores de texto
	SanitizeNameCase   SanitizationType = "name_case"   // Converte para formato de nome próprio
	SanitizeUpperCase  SanitizationType = "uppercase"   // Converte para maiúsculo
	SanitizeLowerCase  SanitizationType = "lowercase"   // Converte para minúsculo
	SanitizeTrimSpaces SanitizationType = "trim_spaces" // Remove espaços extras

	// Email (validação simples)
	SanitizeEmail SanitizationType = "email" // Valida formato de email

	// Data com timezone
	SanitizeDateTimezone SanitizationType = "get_date_timezone" // Extrai data/hora com timezone configurável

	// Custom regex
	SanitizeCustom SanitizationType = "custom" // Regex personalizada
)

// SanitizationConfig configura sanitização de dados de entrada
type SanitizationConfig struct {
	Type        SanitizationType `json:"type"`                   // Tipo de sanitização
	CustomRegex string           `json:"custom_regex,omitempty"` // Regex personalizada (quando type=custom)
	Replacement string           `json:"replacement,omitempty"`  // String de substituição
	Description string           `json:"description,omitempty"`  // Descrição da sanitização
	StrictMode  bool             `json:"strict_mode,omitempty"`  // Se true, falha se não conseguir sanitizar
}

// PersistenceConfig configura persistência de dados para um match
type PersistenceConfig struct {
	Enabled      bool                `json:"enabled"`                 // Se persistência está habilitada
	Scope        Scope               `json:"scope"`                   // Onde persistir: context ou profile
	Key          string              `json:"key"`                     // Chave para armazenamento (ex: "phone_number")
	Sanitization *SanitizationConfig `json:"sanitization,omitempty"`  // Configuração de sanitização
	Required     bool                `json:"required,omitempty"`      // Se o campo é obrigatório
	DefaultValue string              `json:"default_value,omitempty"` // Valor padrão se vazio
}

// MatchConfig estende a configuração de match com persistência
type MatchConfig struct {
	Pattern     string             `json:"pattern"`               // Pattern do match (regex, exact, etc.)
	Type        string             `json:"type"`                  // Tipo do match: "exact", "regex", "contains"
	Persistence *PersistenceConfig `json:"persistence,omitempty"` // Configuração de persistência
}

// PersistenceInfo contém informações sobre as chaves de persistência disponíveis no fluxo
type PersistenceInfo struct {
	ContextKeys []string `json:"context_keys"` // Chaves disponíveis no contexto
	ProfileKeys []string `json:"profile_keys"` // Chaves disponíveis no perfil
}

// Validator interface para validação de configurações de persistência
type Validator interface {
	ValidateConfig(config PersistenceConfig) []ValidationIssue
	ValidateKeyReference(key string, scope Scope) ValidationIssue
}

// ValidationIssue representa um problema de validação
type ValidationIssue struct {
	Code     string `json:"code"`
	Severity string `json:"severity"` // "error", "warn", "info"
	Message  string `json:"message"`
	Path     string `json:"path"`
}

// DefaultValidator implementação padrão do validador
type DefaultValidator struct{}

// ValidateConfig valida uma configuração de persistência
func (v DefaultValidator) ValidateConfig(config PersistenceConfig) []ValidationIssue {
	var issues []ValidationIssue

	if !config.Enabled {
		return issues // Não há o que validar se não está habilitado
	}

	// Valida chave
	if config.Key == "" {
		issues = append(issues, ValidationIssue{
			Code:     "persistence_key_required",
			Severity: "error",
			Message:  "Chave de persistência é obrigatória quando habilitada",
			Path:     "persistence.key",
		})
	}

	// Valida scope
	if config.Scope != ScopeContext && config.Scope != ScopeProfile {
		issues = append(issues, ValidationIssue{
			Code:     "persistence_invalid_scope",
			Severity: "error",
			Message:  "Scope deve ser 'context' ou 'profile'",
			Path:     "persistence.scope",
		})
	}

	// Valida sanitização
	if config.Sanitization != nil {
		issues = append(issues, v.validateSanitization(*config.Sanitization)...)
	}

	return issues
}

// validateSanitization valida configuração de sanitização
func (v DefaultValidator) validateSanitization(config SanitizationConfig) []ValidationIssue {
	var issues []ValidationIssue

	// Valida tipo
	validTypes := []SanitizationType{
		SanitizeNumbersOnly, SanitizeLettersOnly, SanitizeAlphanumeric,
		SanitizeCPF, SanitizeCEP, SanitizePhone,
		SanitizeNameCase, SanitizeUpperCase, SanitizeLowerCase, SanitizeTrimSpaces,
		SanitizeEmail, SanitizeCustom,
	}

	isValidType := false
	for _, validType := range validTypes {
		if config.Type == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		issues = append(issues, ValidationIssue{
			Code:     "sanitization_invalid_type",
			Severity: "error",
			Message:  "Tipo de sanitização inválido",
			Path:     "persistence.sanitization.type",
		})
	}

	// Valida regex personalizada
	if config.Type == SanitizeCustom {
		if config.CustomRegex == "" {
			issues = append(issues, ValidationIssue{
				Code:     "sanitization_custom_regex_required",
				Severity: "error",
				Message:  "Regex personalizada é obrigatória para tipo 'custom'",
				Path:     "persistence.sanitization.custom_regex",
			})
		} else {
			// Testa se a regex é válida
			_, err := regexp.Compile(config.CustomRegex)
			if err != nil {
				issues = append(issues, ValidationIssue{
					Code:     "sanitization_invalid_regex",
					Severity: "error",
					Message:  "Regex personalizada inválida: " + err.Error(),
					Path:     "persistence.sanitization.custom_regex",
				})
			}
		}
	}

	return issues
}

// ValidateKeyReference valida se uma referência de chave é válida
func (v DefaultValidator) ValidateKeyReference(key string, scope Scope) ValidationIssue {
	if key == "" {
		return ValidationIssue{
			Code:     "empty_key_reference",
			Severity: "error",
			Message:  "Referência de chave vazia",
		}
	}

	if scope != ScopeContext && scope != ScopeProfile {
		return ValidationIssue{
			Code:     "invalid_scope_reference",
			Severity: "error",
			Message:  "Scope inválido na referência",
		}
	}

	return ValidationIssue{} // Sem problemas
}

// DefaultSanitizer implementação padrão do sanitizador
type DefaultSanitizer struct{}

// Sanitize aplica sanitização em um texto
func (s DefaultSanitizer) Sanitize(input string, config SanitizationConfig) (string, error) {
	if input == "" {
		return input, nil
	}

	switch config.Type {
	case SanitizeNumbersOnly:
		return s.extractNumbers(input), nil

	case SanitizeLettersOnly:
		return s.extractLetters(input), nil

	case SanitizeAlphanumeric:
		return s.extractAlphanumeric(input), nil

	case SanitizeCPF:
		return s.sanitizeCPF(input)

	case SanitizeCEP:
		return s.sanitizeCEP(input)

	case SanitizePhone:
		return s.sanitizePhone(input)

	case SanitizeBRL:
		return s.sanitizeBRL(input)

	case SanitizeNameCase:
		return s.nameCase(input), nil

	case SanitizeUpperCase:
		return strings.ToUpper(strings.TrimSpace(input)), nil

	case SanitizeLowerCase:
		return strings.ToLower(strings.TrimSpace(input)), nil

	case SanitizeTrimSpaces:
		return s.trimExtraSpaces(input), nil

	case SanitizeEmail:
		return s.sanitizeEmail(input)

	case SanitizeDateTimezone:
		return s.getDateWithTimezone(input, config)

	case SanitizeCustom:
		return s.applyCustomRegex(input, config)

	default:
		return input, errors.New("tipo de sanitização não suportado")
	}
}

// extractNumbers extrai apenas números de um texto
func (s DefaultSanitizer) extractNumbers(input string) string {
	re := regexp.MustCompile(`\d`)
	matches := re.FindAllString(input, -1)
	return strings.Join(matches, "")
}

// extractLetters extrai apenas letras de um texto
func (s DefaultSanitizer) extractLetters(input string) string {
	re := regexp.MustCompile(`[a-zA-ZÀ-ÿ]`)
	matches := re.FindAllString(input, -1)
	return strings.Join(matches, "")
}

// extractAlphanumeric extrai letras e números
func (s DefaultSanitizer) extractAlphanumeric(input string) string {
	re := regexp.MustCompile(`[a-zA-Z0-9À-ÿ]`)
	matches := re.FindAllString(input, -1)
	return strings.Join(matches, "")
}

// sanitizeCPF extrai e formata CPF
func (s DefaultSanitizer) sanitizeCPF(input string) (string, error) {
	numbers := s.extractNumbers(input)
	if len(numbers) != 11 {
		return "", errors.New("CPF deve ter 11 dígitos")
	}
	// Formato: XXX.XXX.XXX-XX
	return numbers[:3] + "." + numbers[3:6] + "." + numbers[6:9] + "-" + numbers[9:], nil
}

// sanitizeCEP extrai e formata CEP
func (s DefaultSanitizer) sanitizeCEP(input string) (string, error) {
	numbers := s.extractNumbers(input)
	if len(numbers) != 8 {
		return "", errors.New("CEP deve ter 8 dígitos")
	}
	// Formato: XXXXX-XXX
	return numbers[:5] + "-" + numbers[5:], nil
}

// sanitizePhone extrai e formata telefone brasileiro
func (s DefaultSanitizer) sanitizePhone(input string) (string, error) {
	numbers := s.extractNumbers(input)

	// Remove código do país se presente
	if len(numbers) == 13 && strings.HasPrefix(numbers, "55") {
		numbers = numbers[2:]
	}

	if len(numbers) == 11 {
		// Celular: (XX) 9XXXX-XXXX
		return "(" + numbers[:2] + ") " + numbers[2:7] + "-" + numbers[7:], nil
	} else if len(numbers) == 10 {
		// Fixo: (XX) XXXX-XXXX
		return "(" + numbers[:2] + ") " + numbers[2:6] + "-" + numbers[6:], nil
	}

	return "", errors.New("telefone deve ter 10 ou 11 dígitos")
}

// nameCase converte texto para formato de nome próprio
func (s DefaultSanitizer) nameCase(input string) string {
	input = strings.TrimSpace(input)
	words := strings.Fields(input)

	for i, word := range words {
		if len(word) > 0 {
			// Preposições comuns ficam em minúsculo
			if len(word) <= 3 && (word == "de" || word == "da" || word == "do" || word == "e" || word == "dos" || word == "das") {
				words[i] = strings.ToLower(word)
			} else {
				words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
			}
		}
	}

	// Primeira palavra sempre maiúscula
	if len(words) > 0 && len(words[0]) > 0 {
		words[0] = strings.ToUpper(words[0][:1]) + strings.ToLower(words[0][1:])
	}

	return strings.Join(words, " ")
}

// trimExtraSpaces remove espaços extras
func (s DefaultSanitizer) trimExtraSpaces(input string) string {
	// Remove espaços no início e fim
	input = strings.TrimSpace(input)
	// Substitui múltiplos espaços por um único
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(input, " ")
}

// sanitizeEmail valida formato básico de email
func (s DefaultSanitizer) sanitizeEmail(input string) (string, error) {
	input = strings.TrimSpace(strings.ToLower(input))
	re := regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)
	if !re.MatchString(input) {
		return "", errors.New("formato de email inválido")
	}
	return input, nil
}

// sanitizeBRL extrai e formata valor monetário em reais
func (s DefaultSanitizer) sanitizeBRL(input string) (string, error) {
	re := regexp.MustCompile(`[\d,.]+`)
	match := re.FindString(input)
	if match == "" {
		return "", errors.New("valor monetário não encontrado")
	}
	match = strings.ReplaceAll(match, ".", "")
	match = strings.ReplaceAll(match, ",", ".")
	var valor float64
	_, err := fmt.Sscanf(match, "%f", &valor)
	if err != nil {
		return "", errors.New("valor monetário inválido")
	}
	return fmt.Sprintf("R$ %.2f", valor), nil
}

// getDateWithTimezone retorna data/hora atual no timezone especificado
func (s DefaultSanitizer) getDateWithTimezone(input string, config SanitizationConfig) (string, error) {
	tz := config.Replacement
	if tz == "" {
		tz = "America/Sao_Paulo"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return "", errors.New("timezone inválido")
	}
	now := time.Now().In(loc)
	return now.Format("2006-01-02 15:04:05 MST"), nil
}

// applyCustomRegex aplica regex personalizada
func (s DefaultSanitizer) applyCustomRegex(input string, config SanitizationConfig) (string, error) {
	re, err := regexp.Compile(config.CustomRegex)
	if err != nil {
		return "", err
	}

	if config.Replacement != "" {
		return re.ReplaceAllString(input, config.Replacement), nil
	}

	// Se não há replacement, extrai matches
	matches := re.FindAllString(input, -1)
	return strings.Join(matches, ""), nil
}
