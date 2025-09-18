package component

import (
	"context"

	"lib-bot/liquid"
	"lib-bot/runtime"
)

// Message componente para mensagens de texto simples ou HSM
type Message struct {
	text string          // Texto da mensagem
	hsm  *HSMView        // Configuração de HSM (opcional)
	det  liquid.Detector // Detector para parsing de templates Liquid
}

// NewMessage cria nova instância de componente message
func NewMessage(det liquid.Detector) *Message { return &Message{det: det} }

// Kind retorna o tipo do componente
func (m *Message) Kind() string { return "message" }

// WithText define o texto da mensagem
func (m *Message) WithText(s string) *Message { cp := *m; cp.text = s; return &cp }

// WithHSM define configuração de HSM
func (m *Message) WithHSM(h *HSMView) *Message { cp := *m; cp.hsm = h; return &cp }

// Spec gera o ComponentSpec com parsing de templates (sem renderização)
func (m *Message) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	if m.hsm != nil {
		// Processa parâmetros da HSM para detectar templates Liquid
		for i := range m.hsm.Params {
			meta, err := m.det.Parse(ctx, m.hsm.Params[i].Raw)
			if err != nil {
				return ComponentSpec{}, err
			}
			m.hsm.Params[i].Template = meta.IsTemplate
			m.hsm.Params[i].Liquid = meta
		}
		// Botões do HSM (labels também podem ter Liquid)
		for i := range m.hsm.Buttons {
			meta, err := m.det.Parse(ctx, m.hsm.Buttons[i].Label.Raw)
			if err != nil {
				return ComponentSpec{}, err
			}
			m.hsm.Buttons[i].Label.Template = meta.IsTemplate
			m.hsm.Buttons[i].Label.Liquid = meta
		}
		return ComponentSpec{Kind: "message", HSM: m.hsm}, nil
	}
	meta, err := m.det.Parse(ctx, m.text)
	if err != nil {
		return ComponentSpec{}, err
	}
	tv := &TextValue{Raw: m.text, Template: meta.IsTemplate, Liquid: meta}
	return ComponentSpec{Kind: "message", Text: tv}, nil
}

// Factory

type MessageFactory struct{ det liquid.Detector }

func NewMessageFactory(det liquid.Detector) *MessageFactory { return &MessageFactory{det: det} }

func (f *MessageFactory) New(_ string, props map[string]any) (Component, error) {
	m := NewMessage(f.det)
	if t, _ := props["text"].(string); t != "" {
		m = m.WithText(t)
	}
	if raw, ok := props["hsm_ref"].(map[string]any); ok && raw != nil {
		h := decodeHSM(raw)
		m = m.WithHSM(h)
	}
	return m, nil
}

func decodeHSM(raw map[string]any) *HSMView {
	h := &HSMView{}
	if v, _ := raw["id"].(string); v != "" {
		h.ID = v
	}
	if v, _ := raw["locale"].(string); v != "" {
		h.Locale = v
	}
	if v, _ := raw["namespace"].(string); v != "" {
		h.Namespace = v
	}
	if v, _ := raw["policy"].(string); v != "" {
		h.Policy = v
	}
	if ps, ok := raw["params"].([]any); ok {
		for _, p := range ps {
			if s, _ := p.(string); s != "" {
				h.Params = append(h.Params, TextValue{Raw: s})
			}
		}
	}
	if bs, ok := raw["buttons"].([]any); ok {
		for _, b := range bs {
			if m, _ := b.(map[string]any); m != nil {
				lbl, _ := m["label"].(string)
				kind, _ := m["kind"].(string)
				data, _ := m["data"].(string)
				h.Buttons = append(h.Buttons, Button{
					Label:   TextValue{Raw: lbl},
					Payload: data,
					Kind:    kind,
				})
			}
		}
	}
	return h
}
