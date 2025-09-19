// Package sanitize provides default implementations for sanitization.
package sanitize

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/AgendoCerto/lib-bot/persistence"
)

// Static errors for better error handling.
var (
	ErrInvalidCPFLength   = errors.New("CPF must have 11 digits")
	ErrInvalidCEPLength   = errors.New("CEP must have 8 digits")
	ErrInvalidPhoneLength = errors.New("phone must have 10 or 11 digits")
	ErrInvalidEmailFormat = errors.New("invalid email format")
	ErrMonetaryNotFound   = errors.New("monetary value not found")
	ErrInvalidMonetary    = errors.New("invalid monetary value")
)

const (
	cpfDigits   = 11
	cepDigits   = 8
	phoneDigits = 10
	cellDigits  = 11
	countryCode = 13
)

// DefaultExtractor provides default text extraction implementation.
type DefaultExtractor struct{}

// Extract extracts specific characters from input text.
func (e *DefaultExtractor) Extract(input string, sanitizationType persistence.SanitizationType) string {
	switch sanitizationType {
	case persistence.SanitizeNumbersOnly:
		return e.extractNumbers(input)
	case persistence.SanitizeLettersOnly:
		return e.extractLetters(input)
	case persistence.SanitizeAlphanumeric:
		return e.extractAlphanumeric(input)
	default:
		return input
	}
}

// extractNumbers extracts only numbers from text.
func (e *DefaultExtractor) extractNumbers(input string) string {
	regexPattern := regexp.MustCompile(`\d`)
	matches := regexPattern.FindAllString(input, -1)
	return strings.Join(matches, "")
}

// extractLetters extracts only letters from text.
func (e *DefaultExtractor) extractLetters(input string) string {
	regexPattern := regexp.MustCompile(`[a-zA-ZÀ-ÿ]`)
	matches := regexPattern.FindAllString(input, -1)
	return strings.Join(matches, "")
}

// extractAlphanumeric extracts letters and numbers.
func (e *DefaultExtractor) extractAlphanumeric(input string) string {
	regexPattern := regexp.MustCompile(`[a-zA-Z0-9À-ÿ]`)
	matches := regexPattern.FindAllString(input, -1)
	return strings.Join(matches, "")
}

// DefaultFormatter provides default document formatting implementation.
type DefaultFormatter struct{}

// Format formats input according to Brazilian document standards.
func (f *DefaultFormatter) Format(input string, sanitizationType persistence.SanitizationType) (string, error) {
	switch sanitizationType {
	case persistence.SanitizeCPF:
		return f.formatCPF(input)
	case persistence.SanitizeCEP:
		return f.formatCEP(input)
	case persistence.SanitizePhone:
		return f.formatPhone(input)
	case persistence.SanitizeBRL:
		return f.formatBRL(input)
	default:
		return input, nil
	}
}

// formatCPF extracts and formats CPF.
func (f *DefaultFormatter) formatCPF(input string) (string, error) {
	extractor := &DefaultExtractor{}
	numbers := extractor.extractNumbers(input)

	if len(numbers) != cpfDigits {
		return "", ErrInvalidCPFLength
	}

	// Format: XXX.XXX.XXX-XX
	return numbers[:3] + "." + numbers[3:6] + "." + numbers[6:9] + "-" + numbers[9:], nil
}

// formatCEP extracts and formats CEP.
func (f *DefaultFormatter) formatCEP(input string) (string, error) {
	extractor := &DefaultExtractor{}
	numbers := extractor.extractNumbers(input)

	if len(numbers) != cepDigits {
		return "", ErrInvalidCEPLength
	}

	// Format: XXXXX-XXX
	return numbers[:5] + "-" + numbers[5:], nil
}

// formatPhone extracts and formats Brazilian phone.
func (f *DefaultFormatter) formatPhone(input string) (string, error) {
	extractor := &DefaultExtractor{}
	numbers := extractor.extractNumbers(input)

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

// formatBRL extracts and formats monetary value in reais.
func (f *DefaultFormatter) formatBRL(input string) (string, error) {
	regexPattern := regexp.MustCompile(`[\d,.]+`)
	match := regexPattern.FindString(input)

	if match == "" {
		return "", ErrMonetaryNotFound
	}

	match = strings.ReplaceAll(match, ".", "")
	match = strings.ReplaceAll(match, ",", ".")

	var value float64
	_, err := fmt.Sscanf(match, "%f", &value)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidMonetary, err)
	}

	return fmt.Sprintf("R$ %.2f", value), nil
}

// DefaultValidator provides default validation implementation.
type DefaultValidator struct{}

// Validate validates and sanitizes input.
func (v *DefaultValidator) Validate(input string, sanitizationType persistence.SanitizationType) (string, error) {
	switch sanitizationType {
	case persistence.SanitizeEmail:
		return v.validateEmail(input)
	default:
		return input, nil
	}
}

// validateEmail validates basic email format.
func (v *DefaultValidator) validateEmail(input string) (string, error) {
	input = strings.TrimSpace(strings.ToLower(input))
	regexPattern := regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)

	if !regexPattern.MatchString(input) {
		return "", ErrInvalidEmailFormat
	}

	return input, nil
}

// DefaultNormalizer provides default text normalization implementation.
type DefaultNormalizer struct{}

// Normalize normalizes text format.
func (n *DefaultNormalizer) Normalize(input string, sanitizationType persistence.SanitizationType) string {
	switch sanitizationType {
	case persistence.SanitizeNameCase:
		return n.nameCase(input)
	case persistence.SanitizeUpperCase:
		return strings.ToUpper(strings.TrimSpace(input))
	case persistence.SanitizeLowerCase:
		return strings.ToLower(strings.TrimSpace(input))
	case persistence.SanitizeTrimSpaces:
		return n.trimExtraSpaces(input)
	default:
		return input
	}
}

// nameCase converts text to proper name format.
func (n *DefaultNormalizer) nameCase(input string) string {
	input = strings.TrimSpace(input)
	words := strings.Fields(input)

	n.processWords(words)
	n.ensureFirstWordCapitalized(words)

	return strings.Join(words, " ")
}

// processWords applies proper case to each word.
func (n *DefaultNormalizer) processWords(words []string) {
	for index, word := range words {
		if len(word) > 0 {
			words[index] = n.formatWord(word)
		}
	}
}

// formatWord formats a single word based on Portuguese rules.
func (n *DefaultNormalizer) formatWord(word string) string {
	if n.isPreposition(word) {
		return strings.ToLower(word)
	}

	return strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
}

// isPreposition checks if a word is a common Portuguese preposition.
func (n *DefaultNormalizer) isPreposition(word string) bool {
	prepositions := []string{"de", "da", "do", "e", "dos", "das"}
	lowerWord := strings.ToLower(word)

	for _, prep := range prepositions {
		if len(word) <= 3 && lowerWord == prep {
			return true
		}
	}

	return false
}

// ensureFirstWordCapitalized ensures the first word is always capitalized.
func (n *DefaultNormalizer) ensureFirstWordCapitalized(words []string) {
	if len(words) > 0 && len(words[0]) > 0 {
		words[0] = strings.ToUpper(words[0][:1]) + strings.ToLower(words[0][1:])
	}
}

// trimExtraSpaces removes extra spaces.
func (n *DefaultNormalizer) trimExtraSpaces(input string) string {
	// Remove spaces at beginning and end
	input = strings.TrimSpace(input)
	// Replace multiple spaces with single space
	regexPattern := regexp.MustCompile(`\s+`)

	return regexPattern.ReplaceAllString(input, " ")
}
