package component

import (
	"context"
	"strconv"

	"lib-bot/runtime"
)

// DelayView representa um componente de delay/pausa
type DelayView struct {
	Duration   int    `json:"duration"`              // Duração em millisegundos
	Unit       string `json:"unit,omitempty"`        // Unidade (milliseconds, seconds)
	Reason     string `json:"reason,omitempty"`      // Motivo do delay
	ShowTyping bool   `json:"show_typing,omitempty"` // Mostrar indicador de digitação
	Message    string `json:"message,omitempty"`     // Mensagem opcional durante delay
}

// DelayComponent implementa um componente de delay
type DelayComponent struct {
	props map[string]any
}

// NewDelayComponent cria novo componente de delay
func NewDelayComponent(props map[string]any) *DelayComponent {
	return &DelayComponent{props: props}
}

// Kind retorna o tipo do componente
func (d *DelayComponent) Kind() string {
	return "delay"
}

// Spec gera o ComponentSpec para delay
func (d *DelayComponent) Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) {
	view := DelayView{}

	// Duration obrigatório
	if duration, ok := d.props["duration"]; ok {
		switch v := duration.(type) {
		case int:
			view.Duration = v
		case float64:
			view.Duration = int(v)
		case string:
			if parsed, err := strconv.Atoi(v); err == nil {
				view.Duration = parsed
			}
		}
	}

	// Unit opcional
	if unit, ok := d.props["unit"].(string); ok {
		view.Unit = unit
	} else {
		view.Unit = "milliseconds" // padrão
	}

	// Reason opcional
	if reason, ok := d.props["reason"].(string); ok {
		view.Reason = reason
	}

	// ShowTyping opcional
	if showTyping, ok := d.props["show_typing"].(bool); ok {
		view.ShowTyping = showTyping
	}

	// Message opcional
	if message, ok := d.props["message"].(string); ok {
		view.Message = message
	}

	return ComponentSpec{
		Kind: "delay",
		Meta: map[string]any{
			"duration":    view.Duration,
			"unit":        view.Unit,
			"reason":      view.Reason,
			"show_typing": view.ShowTyping,
			"message":     view.Message,
		},
	}, nil
}

// DelayFactory cria componentes de delay
func DelayFactory(props map[string]any) (Component, error) {
	return NewDelayComponent(props), nil
}
