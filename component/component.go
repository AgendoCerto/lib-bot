// Package component define tipos e interfaces para componentes de conversação
package component

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AgendoCerto/lib-bot/hsm"
	"github.com/AgendoCerto/lib-bot/liquid"
	"github.com/AgendoCerto/lib-bot/persistence"
	"github.com/AgendoCerto/lib-bot/runtime"
	"github.com/AgendoCerto/lib-bot/validator"
)

// TextValue armazena texto com suporte a templates Liquid (sem renderização)
type TextValue struct {
	Raw      string      `json:"raw"`      // Texto original com possíveis templates
	Template bool        `json:"template"` // Indica se contém templates Liquid
	Liquid   liquid.Meta `json:"liquid"`   // Metadados de parsing do Liquid
}

// HSMView representa uma HSM (Highly Structured Message) com parâmetros templated
type HSMView struct {
	ID        string      `json:"id"`                  // Identificador da HSM
	Locale    string      `json:"locale"`              // Localização (ex: pt_BR)
	Namespace string      `json:"namespace,omitempty"` // Namespace da HSM
	Params    []TextValue `json:"params"`              // Parâmetros (podem conter Liquid)
	Buttons   []Button    `json:"buttons,omitempty"`   // Botões interativos
	Policy    string      `json:"policy,omitempty"`    // Política de fallback: error_on_missing|fallback_to_text|fallback_to_menu
}

// Button representa um botão interativo
type Button struct {
	Label   TextValue `json:"label"`   // Texto do botão (pode ter templates)
	Payload string    `json:"payload"` // Dados enviados ao clicar
	Kind    string    `json:"kind"`    // Tipo: reply|url|call
}

// REMOVIDO: AwaitBehavior - substituído por validator.enabled=true + validator.timeout_seconds

// TimeoutBehavior configura comportamento de timeout
type TimeoutBehavior struct {
	Duration    int               `json:"duration"`               // Timeout em segundos
	Action      string            `json:"action"`                 // retry|escalate|continue
	MaxAttempts int               `json:"max_attempts,omitempty"` // Máximo de tentativas
	Message     *TextValue        `json:"message,omitempty"`      // Mensagem customizada de timeout
	Escalation  *EscalationConfig `json:"escalation,omitempty"`   // Configuração de escalação
}

// ValidationBehavior configura validação de entradas
type ValidationBehavior struct {
	OnInvalid    string            `json:"on_invalid"`              // retry|escalate|continue
	MaxAttempts  int               `json:"max_attempts,omitempty"`  // Máximo de tentativas
	FallbackText *TextValue        `json:"fallback_text,omitempty"` // Texto para entrada inválida
	Escalation   *EscalationConfig `json:"escalation,omitempty"`    // Configuração de escalação
}

// DelayBehavior configura delays
type DelayBehavior struct {
	Before     int    `json:"before,omitempty"`      // Delay antes (ms)
	After      int    `json:"after,omitempty"`       // Delay depois (ms)
	ShowTyping bool   `json:"show_typing,omitempty"` // Mostrar indicador de digitação
	Reason     string `json:"reason,omitempty"`      // Motivo do delay
}

// EscalationConfig configura escalação para humano
type EscalationConfig struct {
	Action    string     `json:"action"`               // transfer_human|end_conversation
	Message   *TextValue `json:"message,omitempty"`    // Mensagem antes da escalação
	TriggerAt int        `json:"trigger_at,omitempty"` // Número de tentativas para escalar
}

// ExperimentBehavior configura A/B testing
type ExperimentBehavior struct {
	Enabled   bool                `json:"enabled"`
	StickyKey string              `json:"sticky_key"` // Chave para manter consistência (ex: "profile.user_id")
	Variants  []ExperimentVariant `json:"variants"`
}

// ExperimentVariant representa uma variante do experimento
type ExperimentVariant struct {
	ID         string `json:"id"`          // ID da variante (ex: "A", "B")
	Weight     int    `json:"weight"`      // Peso para distribuição (0-100)
	TargetNode string `json:"target_node"` // Nó de destino para esta variante
}

// ComponentBehavior agrupa todos os behaviors de um componente
// IMPORTANTE: behavior.validator substitui behavior.await
// Quando validator.enabled=true, automaticamente aguarda resposta do usuário
type ComponentBehavior struct {
	// REMOVIDO: Await - funcionalidade absorvida pelo Validator
	// Use validator.enabled=true para aguardar resposta

	Timeout    *TimeoutBehavior    `json:"timeout,omitempty"`    // Configuração de timeout (complementa validator)
	Validation *ValidationBehavior `json:"validation,omitempty"` // Configuração de validação (LEGADO - usar Validator)
	Validator  *validator.Config   `json:"validator,omitempty"`  // Validator 2.0 (substitui Await + Validation)
	Delay      *DelayBehavior      `json:"delay,omitempty"`      // Configuração de delays
	Experiment *ExperimentBehavior `json:"experiment,omitempty"` // A/B testing
	// REMOVIDO: Persistence - usar ComponentSpec.Persistence ao invés de behavior.persistence
}

// ComponentSpec é o modelo canônico de um componente (sem renderização final)
type ComponentSpec struct {
	Kind        string              `json:"kind"`                  // Tipo do componente (message, confirm, etc.)
	Text        *TextValue          `json:"text,omitempty"`        // Texto principal
	MediaURL    string              `json:"media_url,omitempty"`   // URL de mídia (imagem, vídeo, etc.)
	Buttons     []Button            `json:"buttons,omitempty"`     // Botões interativos
	HSM         *hsm.HSMTemplate    `json:"hsm,omitempty"`         // Configuração de HSM simplificado
	Behavior    *ComponentBehavior  `json:"behavior,omitempty"`    // Configurações de comportamento
	Persistence *persistence.Config `json:"persistence,omitempty"` // Configuração de persistência
	Meta        map[string]any      `json:"meta,omitempty"`        // Metadados adicionais
}

// Component interface para geração de specs canônicos (apenas parsing, sem render)
type Component interface {
	Kind() string                                                          // Retorna o tipo do componente
	Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) // Gera spec canônico
}

// Factory interface para criação de componentes a partir de propriedades
type Factory interface {
	New(kind string, props map[string]any) (Component, error) // Cria componente das props do design
}

// Registry gerencia fábricas de componentes por tipo
type Registry struct{ factories map[string]Factory }

// NewRegistry cria um novo registry vazio
func NewRegistry() *Registry { return &Registry{factories: map[string]Factory{}} }

// Register registra uma factory para um tipo específico
func (r *Registry) Register(kind string, f Factory) { r.factories[kind] = f }

// New cria um componente do tipo especificado com as propriedades fornecidas
func (r *Registry) New(kind string, props map[string]any) (Component, error) {
	f, ok := r.factories[kind]
	if !ok {
		return nil, ErrUnknownKind{Kind: kind}
	}
	return f.New(kind, props)
}

// ErrUnknownKind erro retornado quando um tipo de componente não é encontrado
type ErrUnknownKind struct{ Kind string }

func (e ErrUnknownKind) Error() string { return "unknown component kind: " + e.Kind }

// ParseBehavior extrai configurações de behavior das props do componente
func ParseBehavior(props map[string]any, det liquid.Detector) (*ComponentBehavior, error) {
	behavior := &ComponentBehavior{}
	hasAnyBehavior := false

	// REMOVIDO: Parse await behavior - funcionalidade absorvida pelo validator
	// Se props["await"] existir, ignorar silenciosamente (compatibilidade)

	// Parse timeout behavior
	if timeoutRaw, ok := props["timeout"].(map[string]any); ok {
		timeout, err := parseTimeoutBehavior(timeoutRaw, det)
		if err != nil {
			return nil, err
		}
		behavior.Timeout = timeout
		hasAnyBehavior = true
	}

	// Parse validation behavior (legado)
	if validationRaw, ok := props["validation"].(map[string]any); ok {
		validation, err := parseValidationBehavior(validationRaw, det)
		if err != nil {
			return nil, err
		}
		behavior.Validation = validation
		hasAnyBehavior = true
	}

	// Parse validator 2.0 (novo - spec v2.2)
	if validatorRaw, ok := props["validator"].(map[string]any); ok {
		validatorConfig, err := parseValidator(validatorRaw)
		if err != nil {
			return nil, err
		}
		behavior.Validator = validatorConfig
		hasAnyBehavior = true
	}

	// Parse delay behavior
	if delayRaw, ok := props["delay"].(map[string]any); ok {
		delay := parseDelayBehavior(delayRaw)
		behavior.Delay = delay
		hasAnyBehavior = true
	}

	// Parse experiment behavior
	if experimentRaw, ok := props["experiment"].(map[string]any); ok {
		experiment, err := parseExperimentBehavior(experimentRaw)
		if err != nil {
			return nil, err
		}
		behavior.Experiment = experiment
		hasAnyBehavior = true
	}

	// REMOVIDO: Parse persistence behavior - usar ParsePersistence() separadamente
	// A persistência deve ser configurada via ComponentSpec.Persistence, não behavior.persistence

	// Compatibilidade com formato antigo de fallback
	if fallbackRaw, ok := props["fallback"].(map[string]any); ok {
		timeout, validation, err := parseCompatFallback(fallbackRaw, det)
		if err != nil {
			return nil, err
		}
		if timeout != nil {
			behavior.Timeout = timeout
			hasAnyBehavior = true
		}
		if validation != nil {
			behavior.Validation = validation
			hasAnyBehavior = true
		}
	}

	if !hasAnyBehavior {
		return nil, nil
	}
	return behavior, nil
}

// REMOVIDO: parseAwaitBehavior - substituído por validator com timeout_seconds

func parseTimeoutBehavior(raw map[string]any, det liquid.Detector) (*TimeoutBehavior, error) {
	timeout := &TimeoutBehavior{}

	if duration, ok := raw["duration"].(float64); ok {
		timeout.Duration = int(duration)
	} else if duration, ok := raw["duration"].(int); ok {
		timeout.Duration = duration
	}

	if action, ok := raw["action"].(string); ok {
		timeout.Action = action
	} else {
		timeout.Action = "retry" // padrão
	}

	if maxAttempts, ok := raw["max_attempts"].(float64); ok {
		timeout.MaxAttempts = int(maxAttempts)
	} else if maxAttempts, ok := raw["max_attempts"].(int); ok {
		timeout.MaxAttempts = maxAttempts
	}

	if messageText, ok := raw["message"].(string); ok {
		meta, err := det.Parse(context.Background(), messageText)
		if err != nil {
			return nil, err
		}
		timeout.Message = &TextValue{
			Raw:      messageText,
			Template: meta.IsTemplate,
			Liquid:   meta,
		}
	}

	if escalationRaw, ok := raw["escalation"].(map[string]any); ok {
		escalation, err := parseEscalationConfig(escalationRaw, det)
		if err != nil {
			return nil, err
		}
		timeout.Escalation = escalation
	}

	return timeout, nil
}

func parseValidationBehavior(raw map[string]any, det liquid.Detector) (*ValidationBehavior, error) {
	validation := &ValidationBehavior{}

	if onInvalid, ok := raw["on_invalid"].(string); ok {
		validation.OnInvalid = onInvalid
	} else {
		validation.OnInvalid = "retry" // padrão
	}

	if maxAttempts, ok := raw["max_attempts"].(float64); ok {
		validation.MaxAttempts = int(maxAttempts)
	} else if maxAttempts, ok := raw["max_attempts"].(int); ok {
		validation.MaxAttempts = maxAttempts
	}

	if fallbackText, ok := raw["fallback_text"].(string); ok {
		meta, err := det.Parse(context.Background(), fallbackText)
		if err != nil {
			return nil, err
		}
		validation.FallbackText = &TextValue{
			Raw:      fallbackText,
			Template: meta.IsTemplate,
			Liquid:   meta,
		}
	}

	if escalationRaw, ok := raw["escalation"].(map[string]any); ok {
		escalation, err := parseEscalationConfig(escalationRaw, det)
		if err != nil {
			return nil, err
		}
		validation.Escalation = escalation
	}

	return validation, nil
}

func parseDelayBehavior(raw map[string]any) *DelayBehavior {
	delay := &DelayBehavior{}

	if before, ok := raw["before"].(float64); ok {
		delay.Before = int(before)
	} else if before, ok := raw["before"].(int); ok {
		delay.Before = before
	}

	if after, ok := raw["after"].(float64); ok {
		delay.After = int(after)
	} else if after, ok := raw["after"].(int); ok {
		delay.After = after
	}

	if showTyping, ok := raw["show_typing"].(bool); ok {
		delay.ShowTyping = showTyping
	}

	if reason, ok := raw["reason"].(string); ok {
		delay.Reason = reason
	}

	return delay
}

func parseEscalationConfig(raw map[string]any, det liquid.Detector) (*EscalationConfig, error) {
	escalation := &EscalationConfig{}

	if action, ok := raw["action"].(string); ok {
		escalation.Action = action
	}

	if triggerAt, ok := raw["trigger_at"].(float64); ok {
		escalation.TriggerAt = int(triggerAt)
	} else if triggerAt, ok := raw["trigger_at"].(int); ok {
		escalation.TriggerAt = triggerAt
	}

	if messageText, ok := raw["message"].(string); ok {
		meta, err := det.Parse(context.Background(), messageText)
		if err != nil {
			return nil, err
		}
		escalation.Message = &TextValue{
			Raw:      messageText,
			Template: meta.IsTemplate,
			Liquid:   meta,
		}
	}

	return escalation, nil
}

// parseCompatFallback converte formato antigo de fallback para nova estrutura
func parseCompatFallback(raw map[string]any, det liquid.Detector) (*TimeoutBehavior, *ValidationBehavior, error) {
	var timeout *TimeoutBehavior
	var validation *ValidationBehavior

	// Se tem timeout, cria TimeoutBehavior
	if timeoutVal, ok := raw["timeout"].(float64); ok {
		timeout = &TimeoutBehavior{
			Duration: int(timeoutVal),
			Action:   "retry",
		}
		if maxAttempts, ok := raw["max_attempts"].(float64); ok {
			timeout.MaxAttempts = int(maxAttempts)
		} else if maxAttempts, ok := raw["max_attempts"].(int); ok {
			timeout.MaxAttempts = maxAttempts
		}
	} else if timeoutVal, ok := raw["timeout"].(int); ok {
		timeout = &TimeoutBehavior{
			Duration: timeoutVal,
			Action:   "retry",
		}
		if maxAttempts, ok := raw["max_attempts"].(float64); ok {
			timeout.MaxAttempts = int(maxAttempts)
		} else if maxAttempts, ok := raw["max_attempts"].(int); ok {
			timeout.MaxAttempts = maxAttempts
		}
	}

	// Se tem texto de fallback, cria ValidationBehavior
	if fallbackText, ok := raw["text"].(string); ok {
		meta, err := det.Parse(context.Background(), fallbackText)
		if err != nil {
			return nil, nil, err
		}
		validation = &ValidationBehavior{
			OnInvalid: "retry",
			FallbackText: &TextValue{
				Raw:      fallbackText,
				Template: meta.IsTemplate,
				Liquid:   meta,
			},
		}
		if maxAttempts, ok := raw["max_attempts"].(float64); ok {
			validation.MaxAttempts = int(maxAttempts)
		} else if maxAttempts, ok := raw["max_attempts"].(int); ok {
			validation.MaxAttempts = maxAttempts
		}
	}

	return timeout, validation, nil
}

// ParsePersistence extrai configuração de persistência das props
func ParsePersistence(props map[string]any) (*persistence.Config, error) {
	persistenceRaw, hasPersistence := props["persistence"]
	if !hasPersistence {
		return nil, nil
	}

	persistenceMap, ok := persistenceRaw.(map[string]any)
	if !ok {
		return nil, nil
	}

	config := &persistence.Config{}

	// Enabled
	if enabled, ok := persistenceMap["enabled"].(bool); ok {
		config.Enabled = enabled
	}

	// Scope
	if scope, ok := persistenceMap["scope"].(string); ok {
		config.Scope = persistence.Scope(scope)
	}

	// Key
	if key, ok := persistenceMap["key"].(string); ok {
		config.Key = key
	}

	// Required
	if required, ok := persistenceMap["required"].(bool); ok {
		config.Required = required
	}

	// Default value
	if defaultValue, ok := persistenceMap["default_value"].(string); ok {
		config.DefaultValue = defaultValue
	}

	// Sanitization
	if sanitizationRaw, hasSanitization := persistenceMap["sanitization"]; hasSanitization {
		if sanitizationMap, ok := sanitizationRaw.(map[string]any); ok {
			sanitization := &persistence.SanitizationConfig{}

			if sanitizationType, ok := sanitizationMap["type"].(string); ok {
				sanitization.Type = persistence.SanitizationType(sanitizationType)
			}

			if customRegex, ok := sanitizationMap["custom_regex"].(string); ok {
				sanitization.CustomRegex = customRegex
			}

			if replacement, ok := sanitizationMap["replacement"].(string); ok {
				sanitization.Replacement = replacement
			}

			if description, ok := sanitizationMap["description"].(string); ok {
				sanitization.Description = description
			}

			if strictMode, ok := sanitizationMap["strict_mode"].(bool); ok {
				sanitization.StrictMode = strictMode
			}

			config.Sanitization = sanitization
		}
	}

	return config, nil
}

// parseValidator converte map para validator.Config
func parseValidator(raw map[string]any) (*validator.Config, error) {
	// Usar JSON para conversão
	jsonData, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal validator config: %w", err)
	}

	var config validator.Config
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal validator config: %w", err)
	}

	return &config, nil
}

// parseExperimentBehavior converte map para ExperimentBehavior
func parseExperimentBehavior(raw map[string]any) (*ExperimentBehavior, error) {
	experiment := &ExperimentBehavior{}

	if enabled, ok := raw["enabled"].(bool); ok {
		experiment.Enabled = enabled
	}

	if stickyKey, ok := raw["sticky_key"].(string); ok {
		experiment.StickyKey = stickyKey
	}

	if variantsRaw, ok := raw["variants"].([]any); ok {
		variants := make([]ExperimentVariant, 0, len(variantsRaw))
		for _, vRaw := range variantsRaw {
			if vMap, ok := vRaw.(map[string]any); ok {
				variant := ExperimentVariant{}
				if id, ok := vMap["id"].(string); ok {
					variant.ID = id
				}
				if weight, ok := vMap["weight"].(float64); ok {
					variant.Weight = int(weight)
				} else if weight, ok := vMap["weight"].(int); ok {
					variant.Weight = weight
				}
				if targetNode, ok := vMap["target_node"].(string); ok {
					variant.TargetNode = targetNode
				}
				variants = append(variants, variant)
			}
		}
		experiment.Variants = variants
	}

	return experiment, nil
}

// REMOVIDO: parsePersistenceWriteConfig - sistema de persistence com writes foi descontinuado
// Usar ParsePersistence() ao invés, que retorna *persistence.Config
