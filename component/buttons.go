package component

import (
	"context"

	"lib-bot/liquid"
	"lib-bot/runtime"
)

// Buttons componente para mensagens com botões interativos
type Buttons struct {
	text    string          // Texto principal
	buttons []ButtonData    // Lista de botões
	det     liquid.Detector // Detector para parsing de templates Liquid
}

// ButtonData representa dados de um botão
type ButtonData struct {
	Label   string `json:"label"`         // Texto do botão
	Payload string `json:"payload"`       // Payload/ID do botão
	Kind    string `json:"kind"`          // Tipo: reply, url, call
	URL     string `json:"url,omitempty"` // URL para botões do tipo url
}

// NewButtons cria nova instância de componente buttons
func NewButtons(det liquid.Detector) *Buttons {
	return &Buttons{det: det, buttons: make([]ButtonData, 0)}
}

// Kind retorna o tipo do componente
func (b *Buttons) Kind() string { return "buttons" }

// WithText define o texto principal
func (b *Buttons) WithText(s string) *Buttons {
	cp := *b
	cp.text = s
	return &cp
}

// AddButton adiciona um botão
func (b *Buttons) AddButton(label, payload, kind string) *Buttons {
	cp := *b
	cp.buttons = append(cp.buttons, ButtonData{
		Label:   label,
		Payload: payload,
		Kind:    kind,
	})
	return &cp
}

// AddURLButton adiciona um botão com URL
func (b *Buttons) AddURLButton(label, url string) *Buttons {
	cp := *b
	cp.buttons = append(cp.buttons, ButtonData{
		Label:   label,
		Payload: "url_" + url,
		Kind:    "url",
		URL:     url,
	})
	return &cp
}

// Spec gera o ComponentSpec com parsing de templates
func (b *Buttons) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	// Processa texto principal
	var textVal *TextValue
	if b.text != "" {
		meta, err := b.det.Parse(ctx, b.text)
		if err != nil {
			return ComponentSpec{}, err
		}
		textVal = &TextValue{
			Raw:      b.text,
			Template: meta.IsTemplate,
			Liquid:   meta,
		}
	}

	// Processa botões
	buttons := make([]Button, 0, len(b.buttons))
	for _, btn := range b.buttons {
		labelMeta, err := b.det.Parse(ctx, btn.Label)
		if err != nil {
			return ComponentSpec{}, err
		}

		buttons = append(buttons, Button{
			Label: TextValue{
				Raw:      btn.Label,
				Template: labelMeta.IsTemplate,
				Liquid:   labelMeta,
			},
			Payload: btn.Payload,
			Kind:    btn.Kind,
		})
	}

	return ComponentSpec{
		Kind:    "buttons",
		Text:    textVal,
		Buttons: buttons,
	}, nil
}

// Factory

type ButtonsFactory struct{ det liquid.Detector }

func NewButtonsFactory(det liquid.Detector) *ButtonsFactory {
	return &ButtonsFactory{det: det}
}

func (f *ButtonsFactory) New(_ string, props map[string]any) (Component, error) {
	b := NewButtons(f.det)

	// Texto principal
	if text, _ := props["text"].(string); text != "" {
		b = b.WithText(text)
	}

	// Botões
	if btnList, ok := props["buttons"].([]any); ok {
		for _, btnRaw := range btnList {
			if btnMap, ok := btnRaw.(map[string]any); ok {
				label, _ := btnMap["label"].(string)
				payload, _ := btnMap["payload"].(string)
				kind, _ := btnMap["kind"].(string)
				url, _ := btnMap["url"].(string)

				if kind == "" {
					kind = "reply" // Default
				}

				if kind == "url" && url != "" {
					b = b.AddURLButton(label, url)
				} else {
					b = b.AddButton(label, payload, kind)
				}
			}
		}
	}

	return b, nil
}
