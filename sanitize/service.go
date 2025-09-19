// Package sanitize provides data sanitization functionality.
package sanitize

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"lib-bot/persistence"
)

// Static errors for better error handling.
var (
	ErrUnsupportedSanitizationType = errors.New("unsupported sanitization type")
	ErrExtractorNotFound           = errors.New("extractor not found for type")
	ErrFormatterNotFound           = errors.New("formatter not found for type")
	ErrValidatorNotFound           = errors.New("validator not found for type")
	ErrNormalizerNotFound          = errors.New("normalizer not found for type")
)

// Service provides sanitization functionality.
type Service struct {
	extractors  map[persistence.SanitizationType]TextExtractor
	formatters  map[persistence.SanitizationType]Formatter
	validators  map[persistence.SanitizationType]Validator
	normalizers map[persistence.SanitizationType]Normalizer
}

// NewService creates a new sanitization service with default implementations.
func NewService() *Service {
	service := &Service{
		extractors:  make(map[persistence.SanitizationType]TextExtractor),
		formatters:  make(map[persistence.SanitizationType]Formatter),
		validators:  make(map[persistence.SanitizationType]Validator),
		normalizers: make(map[persistence.SanitizationType]Normalizer),
	}

	service.registerDefaults()
	return service
}

// registerDefaults registers default implementations for all sanitization types.
func (s *Service) registerDefaults() {
	// Extractors
	extractor := &DefaultExtractor{}
	s.extractors[persistence.SanitizeNumbersOnly] = extractor
	s.extractors[persistence.SanitizeLettersOnly] = extractor
	s.extractors[persistence.SanitizeAlphanumeric] = extractor

	// Formatters
	formatter := &DefaultFormatter{}
	s.formatters[persistence.SanitizeCPF] = formatter
	s.formatters[persistence.SanitizeCEP] = formatter
	s.formatters[persistence.SanitizePhone] = formatter
	s.formatters[persistence.SanitizeBRL] = formatter

	// Validators
	validator := &DefaultValidator{}
	s.validators[persistence.SanitizeEmail] = validator

	// Normalizers
	normalizer := &DefaultNormalizer{}
	s.normalizers[persistence.SanitizeNameCase] = normalizer
	s.normalizers[persistence.SanitizeUpperCase] = normalizer
	s.normalizers[persistence.SanitizeLowerCase] = normalizer
	s.normalizers[persistence.SanitizeTrimSpaces] = normalizer
}

// Sanitize applies sanitization based on the configuration.
func (s *Service) Sanitize(input string, config persistence.SanitizationConfig) (string, error) {
	if input == "" {
		return input, nil
	}

	switch config.Type {
	case persistence.SanitizeNumbersOnly, persistence.SanitizeLettersOnly, persistence.SanitizeAlphanumeric:
		return s.handleExtraction(input, config)
	case persistence.SanitizeCPF, persistence.SanitizeCEP, persistence.SanitizePhone, persistence.SanitizeBRL:
		return s.handleFormatting(input, config)
	case persistence.SanitizeEmail:
		return s.handleValidation(input, config)
	case persistence.SanitizeNameCase, persistence.SanitizeUpperCase, persistence.SanitizeLowerCase, persistence.SanitizeTrimSpaces:
		return s.handleNormalization(input, config)
	case persistence.SanitizeDateTimezone:
		return s.handleDateTimezone(config)
	case persistence.SanitizeCustom:
		return s.handleCustomRegex(input, config)
	default:
		return input, fmt.Errorf("%w: %s", ErrUnsupportedSanitizationType, config.Type)
	}
}

// handleExtraction handles text extraction sanitization.
func (s *Service) handleExtraction(input string, config persistence.SanitizationConfig) (string, error) {
	extractor, exists := s.extractors[config.Type]
	if !exists {
		return input, fmt.Errorf("%w: %s", ErrExtractorNotFound, config.Type)
	}

	return extractor.Extract(input, config.Type), nil
}

// handleFormatting handles document formatting sanitization.
func (s *Service) handleFormatting(input string, config persistence.SanitizationConfig) (string, error) {
	formatter, exists := s.formatters[config.Type]
	if !exists {
		return input, fmt.Errorf("%w: %s", ErrFormatterNotFound, config.Type)
	}

	return formatter.Format(input, config.Type)
}

// handleValidation handles validation-based sanitization.
func (s *Service) handleValidation(input string, config persistence.SanitizationConfig) (string, error) {
	validator, exists := s.validators[config.Type]
	if !exists {
		return input, fmt.Errorf("%w: %s", ErrValidatorNotFound, config.Type)
	}

	return validator.Validate(input, config.Type)
}

// handleNormalization handles text normalization sanitization.
func (s *Service) handleNormalization(input string, config persistence.SanitizationConfig) (string, error) {
	normalizer, exists := s.normalizers[config.Type]
	if !exists {
		return input, fmt.Errorf("%w: %s", ErrNormalizerNotFound, config.Type)
	}

	return normalizer.Normalize(input, config.Type), nil
}

// handleDateTimezone handles date/timezone sanitization.
func (s *Service) handleDateTimezone(config persistence.SanitizationConfig) (string, error) {
	timezone := config.Replacement
	if timezone == "" {
		timezone = "America/Sao_Paulo"
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return "", fmt.Errorf("invalid timezone: %w", err)
	}

	now := time.Now().In(location)
	return now.Format("2006-01-02 15:04:05 MST"), nil
}

// handleCustomRegex handles custom regex sanitization.
func (s *Service) handleCustomRegex(input string, config persistence.SanitizationConfig) (string, error) {
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

// RegisterExtractor registers a custom text extractor.
func (s *Service) RegisterExtractor(sanitizationType persistence.SanitizationType, extractor TextExtractor) {
	s.extractors[sanitizationType] = extractor
}

// RegisterFormatter registers a custom formatter.
func (s *Service) RegisterFormatter(sanitizationType persistence.SanitizationType, formatter Formatter) {
	s.formatters[sanitizationType] = formatter
}

// RegisterValidator registers a custom validator.
func (s *Service) RegisterValidator(sanitizationType persistence.SanitizationType, validator Validator) {
	s.validators[sanitizationType] = validator
}

// RegisterNormalizer registers a custom normalizer.
func (s *Service) RegisterNormalizer(sanitizationType persistence.SanitizationType, normalizer Normalizer) {
	s.normalizers[sanitizationType] = normalizer
}
