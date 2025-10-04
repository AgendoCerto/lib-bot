package component

import (
	"context"

	"github.com/AgendoCerto/lib-bot/liquid"
	"github.com/AgendoCerto/lib-bot/runtime"
)

// HSMTrigger componente para envio de HSM/template configurável (spec v2.2)
// Permite controlar quando enviar (imediato, agendado ou por condição)
type HSMTrigger struct {
	templateID      string            // ID do template
	language        string            // Código do idioma (ex: "pt_BR")
	variables       map[string]string // Variáveis do template
	triggerMode     string            // immediate|schedule|condition
	scheduleMinutes int               // Para mode=schedule
	conditionExpr   string            // Para mode=condition
	retryCount      int               // Número de tentativas
	retryInterval   int               // Intervalo mínimo entre tentativas (segundos)
	cooldownKey     string            // Chave para cooldown
	cooldownS       int               // Tempo de cooldown em segundos
	det             liquid.Detector   // Detector para parsing
}

// NewHSMTrigger cria nova instância
func NewHSMTrigger(det liquid.Detector) *HSMTrigger {
	return &HSMTrigger{
		det:           det,
		triggerMode:   "immediate",
		retryCount:    0,
		retryInterval: 0,
	}
}

func (h *HSMTrigger) Kind() string { return "hsm_trigger" }

// WithTemplate define o template e idioma
func (h *HSMTrigger) WithTemplate(id, language string) *HSMTrigger {
	cp := *h
	cp.templateID = id
	cp.language = language
	return &cp
}

// WithVariables define as variáveis do template
func (h *HSMTrigger) WithVariables(vars map[string]string) *HSMTrigger {
	cp := *h
	cp.variables = vars
	return &cp
}

// WithTriggerMode define quando enviar
func (h *HSMTrigger) WithTriggerMode(mode string, param int) *HSMTrigger {
	cp := *h
	cp.triggerMode = mode
	if mode == "schedule" {
		cp.scheduleMinutes = param
	}
	return &cp
}

// WithCondition define expressão condicional
func (h *HSMTrigger) WithCondition(expr string) *HSMTrigger {
	cp := *h
	cp.conditionExpr = expr
	return &cp
}

// WithRetry define configuração de retry
func (h *HSMTrigger) WithRetry(count, interval int) *HSMTrigger {
	cp := *h
	cp.retryCount = count
	cp.retryInterval = interval
	return &cp
}

// WithCooldown define cooldown
func (h *HSMTrigger) WithCooldown(key string, seconds int) *HSMTrigger {
	cp := *h
	cp.cooldownKey = key
	cp.cooldownS = seconds
	return &cp
}

// Spec gera o ComponentSpec
func (h *HSMTrigger) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	metaData := map[string]any{
		"template_id":    h.templateID,
		"language":       h.language,
		"variables":      h.variables,
		"trigger_mode":   h.triggerMode,
		"component_type": "hsm_trigger",
	}

	if h.triggerMode == "schedule" {
		metaData["schedule"] = map[string]int{
			"in_minutes": h.scheduleMinutes,
		}
	}

	if h.conditionExpr != "" {
		metaData["condition_expr"] = h.conditionExpr
	}

	if h.retryCount > 0 {
		metaData["retries"] = map[string]int{
			"count":          h.retryCount,
			"min_interval_s": h.retryInterval,
		}
	}

	if h.cooldownKey != "" {
		metaData["cooldown_key"] = h.cooldownKey
		metaData["cooldown_s"] = h.cooldownS
	}

	return ComponentSpec{
		Kind: "hsm_trigger",
		Meta: metaData,
	}, nil
}

// HSMTriggerFactory factory
type HSMTriggerFactory struct{ det liquid.Detector }

func NewHSMTriggerFactory(det liquid.Detector) *HSMTriggerFactory {
	return &HSMTriggerFactory{det: det}
}

func (f *HSMTriggerFactory) New(_ string, props map[string]any) (Component, error) {
	h := NewHSMTrigger(f.det)

	// Template
	if templateID, ok := props["template_id"].(string); ok {
		language, _ := props["language"].(string)
		if language == "" {
			language = "pt_BR"
		}
		h = h.WithTemplate(templateID, language)
	}

	// Variáveis
	if varsRaw, ok := props["variables"].(map[string]any); ok {
		vars := make(map[string]string)
		for k, v := range varsRaw {
			if str, ok := v.(string); ok {
				vars[k] = str
			}
		}
		h = h.WithVariables(vars)
	}

	// Trigger mode
	if mode, ok := props["trigger_mode"].(string); ok {
		var param int
		if scheduleRaw, ok := props["schedule"].(map[string]any); ok {
			if minutes, ok := scheduleRaw["in_minutes"].(float64); ok {
				param = int(minutes)
			} else if minutes, ok := scheduleRaw["in_minutes"].(int); ok {
				param = minutes
			}
		}
		h = h.WithTriggerMode(mode, param)
	}

	// Condition
	if condExpr, ok := props["condition_expr"].(string); ok {
		h = h.WithCondition(condExpr)
	}

	// Retries
	if retriesRaw, ok := props["retries"].(map[string]any); ok {
		count, _ := retriesRaw["count"].(float64)
		interval, _ := retriesRaw["min_interval_s"].(float64)
		h = h.WithRetry(int(count), int(interval))
	}

	// Cooldown
	if cooldownKey, ok := props["cooldown_key"].(string); ok {
		cooldownS, _ := props["cooldown_s"].(float64)
		h = h.WithCooldown(cooldownKey, int(cooldownS))
	}

	// Parse behaviors
	behavior, err := ParseBehavior(props, f.det)
	if err != nil {
		return nil, err
	}

	return &HSMTriggerWithBehavior{
		hsmTrigger: h,
		behavior:   behavior,
	}, nil
}

// HSMTriggerWithBehavior wrapper
type HSMTriggerWithBehavior struct {
	hsmTrigger *HSMTrigger
	behavior   *ComponentBehavior
}

func (hwb *HSMTriggerWithBehavior) Kind() string {
	return hwb.hsmTrigger.Kind()
}

func (hwb *HSMTriggerWithBehavior) Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) {
	spec, err := hwb.hsmTrigger.Spec(ctx, rctx)
	if err != nil {
		return spec, err
	}

	spec.Behavior = hwb.behavior
	return spec, nil
}
