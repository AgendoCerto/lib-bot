// Package sanitize provides interfaces for data sanitization.
package sanitize

import "github.com/AgendoCerto/lib-bot/persistence"

// TextExtractor extracts specific types of text from input.
type TextExtractor interface {
	Extract(input string, sanitizationType persistence.SanitizationType) string
}

// Formatter formats text according to specific rules.
type Formatter interface {
	Format(input string, sanitizationType persistence.SanitizationType) (string, error)
}

// Validator validates and sanitizes text.
type Validator interface {
	Validate(input string, sanitizationType persistence.SanitizationType) (string, error)
}

// Normalizer normalizes text format.
type Normalizer interface {
	Normalize(input string, sanitizationType persistence.SanitizationType) string
}
