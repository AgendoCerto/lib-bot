// Package persistence provides data persistence and sanitization functionality.
package persistence

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Static errors for better error handling.
var (
	ErrUnsupportedSanitizationType = errors.New("unsupported sanitization type")
	ErrInvalidCPFLength            = errors.New("CPF must have 11 digits")
	ErrInvalidCEPLength            = errors.New("CEP must have 8 digits")
	ErrInvalidPhoneLength          = errors.New("phone must have 10 or 11 digits")
	ErrInvalidEmailFormat          = errors.New("invalid email format")
	ErrMonetaryValueNotFound       = errors.New("monetary value not found")
	ErrInvalidMonetaryValue        = errors.New("invalid monetary value")
	ErrInvalidTimezone             = errors.New("invalid timezone")
)

const (
	cpfDigits   = 11
	cepDigits   = 8
	phoneDigits = 10
	cellDigits  = 11
	countryCode = 13
)

// Compile-time interface implementation checks.
var (
	_ ConfigValidator = (*DefaultValidator)(nil)
	_ KeyValidator    = (*DefaultValidator)(nil)
	_ Sanitizer       = (*DefaultSanitizer)(nil)
)

// DefaultValidator provides default validation implementation.
type DefaultValidator struct{}

// ValidateConfig validates a persistence configuration.
func (v DefaultValidator) ValidateConfig(config Config) []ValidationIssue {
	var issues []ValidationIssue

	if !config.Enabled {
		return issues // Nothing to validate if not enabled
	}

	// Validate key
	if config.Key == "" {
		issues = append(issues, ValidationIssue{
			Code:     "persistence_key_required",
			Severity: "error",
			Message:  "Persistence key is required when enabled",
			Path:     "persistence.key",
		})
	}

	// Validate scope
	if config.Scope != ScopeContext && config.Scope != ScopeProfile {
		issues = append(issues, ValidationIssue{
			Code:     "persistence_invalid_scope",
			Severity: "error",
			Message:  "Scope must be 'context' or 'profile'",
			Path:     "persistence.scope",
		})
	}

	// Validate sanitization
	if config.Sanitization != nil {
		issues = append(issues, v.validateSanitization(*config.Sanitization)...)
	}

	return issues
}

// validateSanitization validates sanitization configuration.
func (v DefaultValidator) validateSanitization(config SanitizationConfig) []ValidationIssue {
	var issues []ValidationIssue

	// Validate type
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
			Message:  "Invalid sanitization type",
			Path:     "persistence.sanitization.type",
		})
	}

	// Validate custom regex
	if config.Type == SanitizeCustom {
		if config.CustomRegex == "" {
			issues = append(issues, ValidationIssue{
				Code:     "sanitization_custom_regex_required",
				Severity: "error",
				Message:  "Custom regex is required for type 'custom'",
				Path:     "persistence.sanitization.customRegex",
			})
		} else {
			// Test if regex is valid
			_, err := regexp.Compile(config.CustomRegex)
			if err != nil {
				issues = append(issues, ValidationIssue{
					Code:     "sanitization_invalid_regex",
					Severity: "error",
					Message:  fmt.Sprintf("Invalid custom regex: %v", err),
					Path:     "persistence.sanitization.customRegex",
				})
			}
		}
	}

	return issues
}

// ValidateKeyReference validates if a key reference is valid.
func (v DefaultValidator) ValidateKeyReference(key string, scope Scope) ValidationIssue {
	if key == "" {
		return ValidationIssue{
			Code:     "empty_key_reference",
			Severity: "error",
			Message:  "Empty key reference",
		}
	}

	if scope != ScopeContext && scope != ScopeProfile {
		return ValidationIssue{
			Code:     "invalid_scope_reference",
			Severity: "error",
			Message:  "Invalid scope in reference",
		}
	}

	return ValidationIssue{} // No problems
}

// DefaultSanitizer provides default sanitization implementation.
type DefaultSanitizer struct{}

// Sanitize applies sanitization to text.
func (s DefaultSanitizer) Sanitize(input string, config SanitizationConfig) (string, error) {
	if input == "" {
		return input, nil
	}

	return s.applySanitizationType(input, config)
}

// applySanitizationType applies the specific sanitization type.
func (s DefaultSanitizer) applySanitizationType(input string, config SanitizationConfig) (string, error) {
	switch config.Type {
	case SanitizeNumbersOnly:
		return s.extractNumbers(input), nil
	case SanitizeLettersOnly:
		return s.extractLetters(input), nil
	case SanitizeAlphanumeric:
		return s.extractAlphanumeric(input), nil
	default:
		return s.applyComplexSanitization(input, config)
	}
}

// applyComplexSanitization handles more complex sanitization types.
func (s DefaultSanitizer) applyComplexSanitization(input string, config SanitizationConfig) (string, error) {
	switch config.Type {
	case SanitizeCPF:
		return s.sanitizeCPF(input)
	case SanitizeCEP:
		return s.sanitizeCEP(input)
	case SanitizePhone:
		return s.sanitizePhone(input)
	case SanitizeBRL:
		return s.sanitizeBRL(input)
	default:
		return s.applyTextSanitization(input, config)
	}
}

// applyTextSanitization handles text-based sanitization types.
func (s DefaultSanitizer) applyTextSanitization(input string, config SanitizationConfig) (string, error) {
	switch config.Type {
	case SanitizeNameCase:
		return s.nameCase(input), nil
	case SanitizeUpperCase:
		return strings.ToUpper(strings.TrimSpace(input)), nil
	case SanitizeLowerCase:
		return strings.ToLower(strings.TrimSpace(input)), nil
	case SanitizeTrimSpaces:
		return s.trimExtraSpaces(input), nil
	default:
		return s.applySpecialSanitization(input, config)
	}
}

// applySpecialSanitization handles special sanitization types.
func (s DefaultSanitizer) applySpecialSanitization(input string, config SanitizationConfig) (string, error) {
	switch config.Type {
	case SanitizeEmail:
		return s.sanitizeEmail(input)
	case SanitizeDateTimezone:
		return s.getDateWithTimezone(config)
	case SanitizeCustom:
		return s.applyCustomRegex(input, config)
	default:
		return input, fmt.Errorf("%w: %s", ErrUnsupportedSanitizationType, config.Type)
	}
}

// extractNumbers extracts only numbers from text.
func (s DefaultSanitizer) extractNumbers(input string) string {
	regexPattern := regexp.MustCompile(`\d`)
	matches := regexPattern.FindAllString(input, -1)

	return strings.Join(matches, "")
}

// extractLetters extracts only letters from text.
func (s DefaultSanitizer) extractLetters(input string) string {
	regexPattern := regexp.MustCompile(`[a-zA-ZÀ-ÿ]`)
	matches := regexPattern.FindAllString(input, -1)

	return strings.Join(matches, "")
}

// extractAlphanumeric extracts letters and numbers.
func (s DefaultSanitizer) extractAlphanumeric(input string) string {
	regexPattern := regexp.MustCompile(`[a-zA-Z0-9À-ÿ]`)
	matches := regexPattern.FindAllString(input, -1)

	return strings.Join(matches, "")
}

// sanitizeCPF extracts and formats CPF.
func (s DefaultSanitizer) sanitizeCPF(input string) (string, error) {
	numbers := s.extractNumbers(input)
	if len(numbers) != cpfDigits {
		return "", ErrInvalidCPFLength
	}
	// Format: XXX.XXX.XXX-XX
	return numbers[:3] + "." + numbers[3:6] + "." + numbers[6:9] + "-" + numbers[9:], nil
}

// sanitizeCEP extracts and formats CEP.
func (s DefaultSanitizer) sanitizeCEP(input string) (string, error) {
	numbers := s.extractNumbers(input)
	if len(numbers) != cepDigits {
		return "", ErrInvalidCEPLength
	}
	// Format: XXXXX-XXX
	return numbers[:5] + "-" + numbers[5:], nil
}

// sanitizePhone extracts and formats Brazilian phone.
func (s DefaultSanitizer) sanitizePhone(input string) (string, error) {
	numbers := s.extractNumbers(input)

	// Remove country code if present
	if len(numbers) == countryCode && strings.HasPrefix(numbers, "55") {
		numbers = numbers[2:]
	}

	switch len(numbers) {
	case cellDigits:
		// Cell: (XX) 9XXXX-XXXX
		return "(" + numbers[:2] + ") " + numbers[2:7] + "-" + numbers[7:], nil
	case phoneDigits:
		// Landline: (XX) XXXX-XXXX
		return "(" + numbers[:2] + ") " + numbers[2:6] + "-" + numbers[6:], nil
	default:
		return "", ErrInvalidPhoneLength
	}
}

// nameCase converts text to proper name format.
func (s DefaultSanitizer) nameCase(input string) string {
	input = strings.TrimSpace(input)
	words := strings.Fields(input)

	s.processWords(words)
	s.ensureFirstWordCapitalized(words)

	return strings.Join(words, " ")
}

// processWords applies proper case to each word.
func (s DefaultSanitizer) processWords(words []string) {
	for index, word := range words {
		if len(word) > 0 {
			words[index] = s.formatWord(word)
		}
	}
}

// formatWord formats a single word based on Portuguese rules.
func (s DefaultSanitizer) formatWord(word string) string {
	if s.isPreposition(word) {
		return strings.ToLower(word)
	}

	return strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
}

// ensureFirstWordCapitalized ensures the first word is always capitalized.
func (s DefaultSanitizer) ensureFirstWordCapitalized(words []string) {
	if len(words) > 0 && len(words[0]) > 0 {
		words[0] = strings.ToUpper(words[0][:1]) + strings.ToLower(words[0][1:])
	}
}

// isPreposition checks if a word is a common Portuguese preposition.
func (s DefaultSanitizer) isPreposition(word string) bool {
	prepositions := []string{"de", "da", "do", "e", "dos", "das"}
	lowerWord := strings.ToLower(word)

	for _, prep := range prepositions {
		if len(word) <= 3 && lowerWord == prep {
			return true
		}
	}

	return false
}

// trimExtraSpaces removes extra spaces.
func (s DefaultSanitizer) trimExtraSpaces(input string) string {
	// Remove spaces at beginning and end
	input = strings.TrimSpace(input)
	// Replace multiple spaces with single space
	regexPattern := regexp.MustCompile(`\s+`)

	return regexPattern.ReplaceAllString(input, " ")
}

// sanitizeEmail validates basic email format.
func (s DefaultSanitizer) sanitizeEmail(input string) (string, error) {
	input = strings.TrimSpace(strings.ToLower(input))
	regexPattern := regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)

	if !regexPattern.MatchString(input) {
		return "", ErrInvalidEmailFormat
	}

	return input, nil
}

// sanitizeBRL extracts and formats monetary value in reais.
func (s DefaultSanitizer) sanitizeBRL(input string) (string, error) {
	regexPattern := regexp.MustCompile(`[\d,.]+`)
	match := regexPattern.FindString(input)

	if match == "" {
		return "", ErrMonetaryValueNotFound
	}

	match = strings.ReplaceAll(match, ".", "")
	match = strings.ReplaceAll(match, ",", ".")

	var value float64
	_, err := fmt.Sscanf(match, "%f", &value)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidMonetaryValue, err)
	}

	return fmt.Sprintf("R$ %.2f", value), nil
}

// getDateWithTimezone returns current date/time in specified timezone.
func (s DefaultSanitizer) getDateWithTimezone(config SanitizationConfig) (string, error) {
	tz := config.Replacement
	if tz == "" {
		tz = "America/Sao_Paulo"
	}

	location, err := time.LoadLocation(tz)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidTimezone, err)
	}

	now := time.Now().In(location)

	return now.Format("2006-01-02 15:04:05 MST"), nil
}

// applyCustomRegex applies custom regex.
func (s DefaultSanitizer) applyCustomRegex(input string, config SanitizationConfig) (string, error) {
	regexPattern, err := regexp.Compile(config.CustomRegex)
	if err != nil {
		return "", fmt.Errorf("invalid regex: %w", err)
	}

	if config.Replacement != "" {
		return regexPattern.ReplaceAllString(input, config.Replacement), nil
	}

	// If no replacement, extract matches
	matches := regexPattern.FindAllString(input, -1)

	return strings.Join(matches, ""), nil
}
