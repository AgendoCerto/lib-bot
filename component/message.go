package component

import (
	"context"

	"lib-bot/hsm"
	"lib-bot/liquid"
	"lib-bot/runtime"
)

// Message componente para mensagens de texto simples ou HSM
type Message struct {
	text string           // Texto da mensagem
	hsm  *hsm.HSMTemplate // Configuração de HSM simplificado
	det  liquid.Detector  // Detector para parsing de templates Liquid
}

// NewMessage cria nova instância de componente message
func NewMessage(det liquid.Detector) *Message { return &Message{det: det} }

// Kind retorna o tipo do componente
func (m *Message) Kind() string { return "message" }

// WithText define o texto da mensagem
func (m *Message) WithText(s string) *Message { cp := *m; cp.text = s; return &cp }

// WithHSM define configuração de HSM
func (m *Message) WithHSM(h *hsm.HSMTemplate) *Message { cp := *m; cp.hsm = h; return &cp }

// Spec gera o ComponentSpec com parsing de templates (sem renderização)
func (m *Message) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	if m.hsm != nil {
		// HSM simplificado - apenas validar se o nome está presente
		if err := m.hsm.Validate(); err != nil {
			return ComponentSpec{}, err
		}
		return ComponentSpec{Kind: "message", HSM: m.hsm}, nil
	}

	// Mensagem de texto simples - processa templates Liquid
	if m.text != "" {
		meta, err := m.det.Parse(ctx, m.text)
		if err != nil {
			return ComponentSpec{}, err
		}
		textVal := &TextValue{
			Raw:      m.text,
			Template: meta.IsTemplate,
			Liquid:   meta,
		}
		return ComponentSpec{Kind: "message", Text: textVal}, nil
	}

	return ComponentSpec{Kind: "message"}, nil
}

// Factory

type MessageFactory struct{ det liquid.Detector }

func NewMessageFactory(det liquid.Detector) *MessageFactory { return &MessageFactory{det: det} }

func (f *MessageFactory) New(_ string, props map[string]any) (Component, error) {
	m := NewMessage(f.det)
	if t, _ := props["text"].(string); t != "" {
		m = m.WithText(t)
	}
	if raw, ok := props["hsm"].(map[string]any); ok && raw != nil {
		h := decodeHSM(raw)
		m = m.WithHSM(h)
	}
	return m, nil
}

func decodeHSM(raw map[string]any) *hsm.HSMTemplate {
	h := &hsm.HSMTemplate{}
	if name, _ := raw["name"].(string); name != "" {
		h.Name = name
	}
	return h
}
