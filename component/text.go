package component

import (
	"context"

	"lib-bot/liquid"
	"lib-bot/persistence"
	"lib-bot/runtime"
)

// Text componente para mensagens de texto simples do WhatsApp
// Diferente do componente message que é para HSM, este é especificamente para texto puro
type Text struct {
	body       string          // Corpo da mensagem
	previewURL bool            // Se deve mostrar preview de URLs
	det        liquid.Detector // Detector para parsing de templates Liquid
}

// NewText cria nova instância de componente text
func NewText(det liquid.Detector) *Text {
	return &Text{det: det, previewURL: false}
}

// Kind retorna o tipo do componente
func (t *Text) Kind() string { return "text" }

// WithBody define o corpo da mensagem de texto
func (t *Text) WithBody(s string) *Text {
	cp := *t
	cp.body = s
	return &cp
}

// WithPreviewURL define se deve mostrar preview de URLs (default: false)
func (t *Text) WithPreviewURL(preview bool) *Text {
	cp := *t
	cp.previewURL = preview
	return &cp
}

// Spec gera o ComponentSpec com parsing de templates (sem renderização)
func (t *Text) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	// Mensagem de texto simples - processa templates Liquid
	if t.body != "" {
		meta, err := t.det.Parse(ctx, t.body)
		if err != nil {
			return ComponentSpec{}, err
		}
		textVal := &TextValue{
			Raw:      t.body,
			Template: meta.IsTemplate,
			Liquid:   meta,
		}

		// Adiciona metadados específicos do WhatsApp text
		metaData := map[string]any{
			"preview_url": t.previewURL,
		}

		return ComponentSpec{
			Kind: "text",
			Text: textVal,
			Meta: metaData,
		}, nil
	}

	// Se não tem corpo, retorna spec básico
	return ComponentSpec{
		Kind: "text",
		Meta: map[string]any{
			"preview_url": t.previewURL,
		},
	}, nil
}

// Factory

type TextFactory struct{ det liquid.Detector }

func NewTextFactory(det liquid.Detector) *TextFactory {
	return &TextFactory{det: det}
}

func (f *TextFactory) New(_ string, props map[string]any) (Component, error) {
	t := NewText(f.det)

	// Configura corpo da mensagem
	if body, _ := props["body"].(string); body != "" {
		t = t.WithBody(body)
	} else if text, _ := props["text"].(string); text != "" {
		t = t.WithBody(text)
	}

	// Configura preview de URL
	if preview, ok := props["preview_url"].(bool); ok {
		t = t.WithPreviewURL(preview)
	}

	// Parse persistence
	persistence, err := ParsePersistence(props)
	if err != nil {
		return nil, err
	}

	return &TextWithBehaviorAndPersistence{
		text:        t,
		behavior:    nil, // Pode ser expandido futuramente
		persistence: persistence,
	}, nil
}

// TextWithBehaviorAndPersistence é um wrapper que inclui behaviors e persistência
type TextWithBehaviorAndPersistence struct {
	text        *Text
	behavior    *ComponentBehavior
	persistence *persistence.PersistenceConfig
}

func (twbp *TextWithBehaviorAndPersistence) Kind() string {
	return twbp.text.Kind()
}

func (twbp *TextWithBehaviorAndPersistence) Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) {
	spec, err := twbp.text.Spec(ctx, rctx)
	if err != nil {
		return spec, err
	}

	spec.Behavior = twbp.behavior
	spec.Persistence = twbp.persistence

	return spec, nil
}
