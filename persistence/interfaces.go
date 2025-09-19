// Package persistence provides interfaces and types for data persistence and sanitization.
package persistence

import (
	"context"
)

// KeyReader provides read access to persistence keys.
type KeyReader interface {
	Get(ctx context.Context, scope Scope, key string) (string, error)
}

// KeyWriter provides write access to persistence keys.
type KeyWriter interface {
	Set(ctx context.Context, scope Scope, key, value string) error
}

// KeyStore combines read and write access to persistence keys.
type KeyStore interface {
	KeyReader
	KeyWriter
}

// Sanitizer applies data sanitization.
type Sanitizer interface {
	Sanitize(input string, config SanitizationConfig) (string, error)
}

// ConfigValidator validates persistence configurations.
type ConfigValidator interface {
	ValidateConfig(config Config) []ValidationIssue
}

// KeyValidator validates key references.
type KeyValidator interface {
	ValidateKeyReference(key string, scope Scope) ValidationIssue
}

// Validator combines configuration and key validation.
type Validator interface {
	ConfigValidator
	KeyValidator
}

// NumberExtractor extracts numbers from text.
type NumberExtractor interface {
	ExtractNumbers(input string) string
}

// LetterExtractor extracts letters from text.
type LetterExtractor interface {
	ExtractLetters(input string) string
}

// AlphanumericExtractor extracts alphanumeric characters from text.
type AlphanumericExtractor interface {
	ExtractAlphanumeric(input string) string
}

// DocumentFormatter formats Brazilian documents.
type DocumentFormatter interface {
	FormatCPF(numbers string) (string, error)
	FormatCEP(numbers string) (string, error)
	FormatPhone(numbers string) (string, error)
}

// TextFormatter formats text in various ways.
type TextFormatter interface {
	NameCase(input string) string
	TrimSpaces(input string) string
}

// EmailValidator validates email addresses.
type EmailValidator interface {
	ValidateEmail(input string) (string, error)
}

// MonetaryFormatter formats monetary values.
type MonetaryFormatter interface {
	FormatBRL(input string) (string, error)
}

// DateTimeProvider provides current date and time.
type DateTimeProvider interface {
	GetDateWithTimezone(timezone string) (string, error)
}

// RegexApplier applies custom regex patterns.
type RegexApplier interface {
	ApplyCustomRegex(input string, config SanitizationConfig) (string, error)
}
