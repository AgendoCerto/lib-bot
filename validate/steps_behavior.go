package validate

import (
	"lib-bot/adapter"
	"lib-bot/component"
)

// BehaviorValidationStep valida configurações de behavior dos componentes
type BehaviorValidationStep struct{}

// NewBehaviorValidationStep cria novo validador de behaviors
func NewBehaviorValidationStep() *BehaviorValidationStep {
	return &BehaviorValidationStep{}
}

func (s *BehaviorValidationStep) Check(spec component.ComponentSpec, caps adapter.Capabilities, path string) []Issue {
	var issues []Issue

	if spec.Behavior == nil {
		return issues
	}

	// Valida timeout behavior
	if spec.Behavior.Timeout != nil {
		issues = append(issues, s.validateTimeoutBehavior(*spec.Behavior.Timeout, path)...)
	}

	// Valida validation behavior
	if spec.Behavior.Validation != nil {
		issues = append(issues, s.validateValidationBehavior(*spec.Behavior.Validation, path)...)
	}

	// Valida delay behavior
	if spec.Behavior.Delay != nil {
		issues = append(issues, s.validateDelayBehavior(*spec.Behavior.Delay, path)...)
	}

	return issues
}

func (s *BehaviorValidationStep) validateTimeoutBehavior(timeout component.TimeoutBehavior, path string) []Issue {
	var issues []Issue

	// Duration deve ser positivo
	if timeout.Duration <= 0 {
		issues = append(issues, Issue{
			Code: "behavior.timeout.invalid_duration", Severity: Err,
			Path: path + ".behavior.timeout.duration",
			Msg:  "timeout duration must be positive",
		})
	}

	// Duration não deve ser muito longo (mais de 1 hora = 3600 segundos)
	if timeout.Duration > 3600 {
		issues = append(issues, Issue{
			Code: "behavior.timeout.duration_too_long", Severity: Warn,
			Path: path + ".behavior.timeout.duration",
			Msg:  "timeout duration is very long (over 1 hour), consider shorter values",
		})
	}

	// Action deve ser válida
	validActions := map[string]bool{"retry": true, "escalate": true, "continue": true}
	if !validActions[timeout.Action] {
		issues = append(issues, Issue{
			Code: "behavior.timeout.invalid_action", Severity: Err,
			Path: path + ".behavior.timeout.action",
			Msg:  "timeout action must be one of: retry, escalate, continue",
		})
	}

	// MaxAttempts deve ser positivo se especificado
	if timeout.MaxAttempts < 0 {
		issues = append(issues, Issue{
			Code: "behavior.timeout.invalid_max_attempts", Severity: Err,
			Path: path + ".behavior.timeout.max_attempts",
			Msg:  "max_attempts must be non-negative",
		})
	}

	// Se MaxAttempts for 0 e action for retry, é problemático
	if timeout.MaxAttempts == 0 && timeout.Action == "retry" {
		issues = append(issues, Issue{
			Code: "behavior.timeout.infinite_retry", Severity: Warn,
			Path: path + ".behavior.timeout",
			Msg:  "retry action with max_attempts=0 may cause infinite loops",
		})
	}

	return issues
}

func (s *BehaviorValidationStep) validateValidationBehavior(validation component.ValidationBehavior, path string) []Issue {
	var issues []Issue

	// OnInvalid deve ser válida
	validActions := map[string]bool{"retry": true, "escalate": true, "continue": true}
	if !validActions[validation.OnInvalid] {
		issues = append(issues, Issue{
			Code: "behavior.validation.invalid_action", Severity: Err,
			Path: path + ".behavior.validation.on_invalid",
			Msg:  "on_invalid action must be one of: retry, escalate, continue",
		})
	}

	// MaxAttempts deve ser positivo se especificado
	if validation.MaxAttempts < 0 {
		issues = append(issues, Issue{
			Code: "behavior.validation.invalid_max_attempts", Severity: Err,
			Path: path + ".behavior.validation.max_attempts",
			Msg:  "max_attempts must be non-negative",
		})
	}

	// Se MaxAttempts for 0 e action for retry, é problemático
	if validation.MaxAttempts == 0 && validation.OnInvalid == "retry" {
		issues = append(issues, Issue{
			Code: "behavior.validation.infinite_retry", Severity: Warn,
			Path: path + ".behavior.validation",
			Msg:  "retry action with max_attempts=0 may cause infinite loops",
		})
	}

	return issues
}

func (s *BehaviorValidationStep) validateDelayBehavior(delay component.DelayBehavior, path string) []Issue {
	var issues []Issue

	// Before e After devem ser não-negativos
	if delay.Before < 0 {
		issues = append(issues, Issue{
			Code: "behavior.delay.invalid_before", Severity: Err,
			Path: path + ".behavior.delay.before",
			Msg:  "before delay must be non-negative",
		})
	}

	if delay.After < 0 {
		issues = append(issues, Issue{
			Code: "behavior.delay.invalid_after", Severity: Err,
			Path: path + ".behavior.delay.after",
			Msg:  "after delay must be non-negative",
		})
	}

	// Delays muito longos podem ser problemáticos (mais de 30 segundos)
	if delay.Before > 30000 {
		issues = append(issues, Issue{
			Code: "behavior.delay.before_too_long", Severity: Warn,
			Path: path + ".behavior.delay.before",
			Msg:  "before delay is very long (over 30 seconds), may impact user experience",
		})
	}

	if delay.After > 30000 {
		issues = append(issues, Issue{
			Code: "behavior.delay.after_too_long", Severity: Warn,
			Path: path + ".behavior.delay.after",
			Msg:  "after delay is very long (over 30 seconds), may impact user experience",
		})
	}

	return issues
}
