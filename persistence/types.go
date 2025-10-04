// Package persistence provides types for data persistence configuration.
package persistence

// Scope defines where information will be persisted.
type Scope string

const (
	ScopeContext Scope = "context" // Temporary flow data (session-scoped)
	ScopeState   Scope = "state"   // Persistent user data (user-scoped)
	ScopeGlobal  Scope = "global"  // Global shared data (bot-scoped, shared across all users)
)

// SanitizationType defines predefined sanitization types.
type SanitizationType string

const (
	// Number extractors.
	SanitizeNumbersOnly  SanitizationType = "numbers_only" // Extract numbers only
	SanitizeLettersOnly  SanitizationType = "letters_only" // Extract letters only
	SanitizeAlphanumeric SanitizationType = "alphanumeric" // Extract letters and numbers

	// Specific formatters.
	SanitizeCPF   SanitizationType = "cpf"   // Extract and format CPF
	SanitizeCEP   SanitizationType = "cep"   // Extract and format CEP
	SanitizePhone SanitizationType = "phone" // Extract and format Brazilian phone

	// Monetary.
	SanitizeBRL SanitizationType = "monetary_brl" // Extract and format monetary value in reais

	// Text normalizers.
	SanitizeNameCase   SanitizationType = "name_case"   // Convert to proper name format
	SanitizeUpperCase  SanitizationType = "uppercase"   // Convert to uppercase
	SanitizeLowerCase  SanitizationType = "lowercase"   // Convert to lowercase
	SanitizeTrimSpaces SanitizationType = "trim_spaces" // Remove extra spaces

	// Email (simple validation).
	SanitizeEmail SanitizationType = "email" // Validate email format

	// Date with timezone.
	SanitizeDateTimezone SanitizationType = "get_date_timezone" // Extract date/time with configurable timezone

	// Custom regex.
	SanitizeCustom SanitizationType = "custom" // Custom regex
)

// SanitizationConfig configures input data sanitization.
type SanitizationConfig struct {
	Type        SanitizationType `json:"type"`                  // Sanitization type
	CustomRegex string           `json:"customRegex,omitempty"` // Custom regex (when type=custom)
	Replacement string           `json:"replacement,omitempty"` // Replacement string
	Description string           `json:"description,omitempty"` // Sanitization description
	StrictMode  bool             `json:"strictMode,omitempty"`  // If true, fail if cannot sanitize
}

// Config configures data persistence for a match.
type Config struct {
	Enabled      bool                `json:"enabled"`                // If persistence is enabled
	Scope        Scope               `json:"scope"`                  // Where to persist: context, state, or global
	Key          string              `json:"key"`                    // Storage key (e.g., "phone_number")
	Sanitization *SanitizationConfig `json:"sanitization,omitempty"` // Sanitization configuration
	Required     bool                `json:"required,omitempty"`     // If field is required
	DefaultValue string              `json:"defaultValue,omitempty"` // Default value if empty
}

// MatchConfig extends match configuration with persistence.
type MatchConfig struct {
	Pattern     string  `json:"pattern"`               // Match pattern (regex, exact, etc.)
	Type        string  `json:"type"`                  // Match type: "exact", "regex", "contains"
	Persistence *Config `json:"persistence,omitempty"` // Persistence configuration
}

// Info contains information about available persistence keys in the flow.
type Info struct {
	ContextKeys []string `json:"contextKeys"` // Available keys in context
	StateKeys   []string `json:"stateKeys"`   // Available keys in state (persistent user data)
	GlobalKeys  []string `json:"globalKeys"`  // Available keys in global (bot-wide shared data)
}

// ValidationIssue represents a validation problem.
type ValidationIssue struct {
	Code     string `json:"code"`
	Severity string `json:"severity"` // "error", "warn", "info"
	Message  string `json:"message"`
	Path     string `json:"path"`
}
