// Package whatsapp implementa adapter específico para WhatsApp Business API
package whatsapp

import (
	"context"
	"errors"

	"lib-bot/adapter"
	"lib-bot/component"
)

// WhatsApp adapter para WhatsApp Business API
type WhatsApp struct{ caps adapter.Capabilities }

// New cria novo adapter WhatsApp com capabilities específicas
func New() *WhatsApp {
	c := adapter.NewCaps()
	c.SupportsHSM = true                                                      // WhatsApp suporta HSM
	c.SupportsRichText = false                                                // Sem formatação rica
	c.MaxTextLen = 1024                                                       // Limite de texto do WhatsApp
	c.MaxButtons = 3                                                          // Máximo 3 botões
	c.ButtonKinds = map[string]bool{"reply": true, "url": true, "call": true} // Tipos suportados
	c.SupportsListPicker = true                                               // Suporte a listas
	return &WhatsApp{caps: c}
}

func (w *WhatsApp) Name() string                       { return "whatsapp" }
func (w *WhatsApp) Capabilities() adapter.Capabilities { return w.caps }

// Transform aplica transformações específicas do WhatsApp aos specs (sem renderização)
func (w *WhatsApp) Transform(_ context.Context, spec component.ComponentSpec) (component.ComponentSpec, error) {
	// Verifica suporte a HSM
	if spec.HSM != nil && !w.caps.SupportsHSM {
		return component.ComponentSpec{}, errors.New("adapter: HSM not supported")
	}

	// Limita número de botões conforme capabilities
	if len(spec.Buttons) > w.caps.MaxButtons {
		spec.Buttons = spec.Buttons[:w.caps.MaxButtons]
	}

	// Filtra tipos de botão não suportados
	out := make([]component.Button, 0, len(spec.Buttons))
	for _, b := range spec.Buttons {
		if w.caps.ButtonKinds[b.Kind] {
			out = append(out, b)
		}
	}
	spec.Buttons = out
	return spec, nil
}
