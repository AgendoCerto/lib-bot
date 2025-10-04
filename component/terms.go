package component

import (
	"context"
	"fmt"

	"github.com/AgendoCerto/lib-bot/liquid"
	"github.com/AgendoCerto/lib-bot/persistence"
	"github.com/AgendoCerto/lib-bot/runtime"
)

// Terms representa o componente de aceite de termos
// Usa BUTTONS (Aceitar/Rejeitar) + suporte a link no texto para visualizar termos completos
type Terms struct {
	text        string          // Texto dos termos (pode incluir link usando markdown [texto](url))
	linkURL     string          // URL opcional para termos completos
	linkText    string          // Texto do link (ex: "Ver termos completos")
	acceptLabel string          // Label do botão aceitar (ex: "Aceito")
	rejectLabel string          // Label do botão rejeitar (ex: "Não aceito")
	det         liquid.Detector // Detector para parsing de templates Liquid
}

// NewTerms cria nova instância de componente terms
func NewTerms(det liquid.Detector) *Terms {
	return &Terms{
		det:         det,
		acceptLabel: "Aceito",
		rejectLabel: "Não aceito",
		linkText:    "Ver termos completos",
	}
}

func (t *Terms) Kind() string { return "terms" }

// WithText define o texto dos termos
func (t *Terms) WithText(s string) *Terms {
	cp := *t
	cp.text = s
	return &cp
}

// WithLink define link para termos completos
func (t *Terms) WithLink(url, text string) *Terms {
	cp := *t
	cp.linkURL = url
	if text != "" {
		cp.linkText = text
	}
	return &cp
}

// WithLabels define os labels dos botões
func (t *Terms) WithLabels(accept, reject string) *Terms {
	cp := *t
	if accept != "" {
		cp.acceptLabel = accept
	}
	if reject != "" {
		cp.rejectLabel = reject
	}
	return &cp
}

// Spec gera o ComponentSpec para terms (usa BUTTONS para aceitar/rejeitar)
func (t *Terms) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	// Processa o texto principal com templates Liquid
	var textVal *TextValue
	if t.text != "" {
		// Se tem link, adiciona ao final do texto
		fullText := t.text
		if t.linkURL != "" {
			fullText = fmt.Sprintf("%s\n\n[%s](%s)", t.text, t.linkText, t.linkURL)
		}

		meta, err := t.det.Parse(ctx, fullText)
		if err != nil {
			return ComponentSpec{}, err
		}
		textVal = &TextValue{
			Raw:      fullText,
			Template: meta.IsTemplate,
			Liquid:   meta,
		}
	}

	// Parse labels dos botões
	acceptMeta, err := t.det.Parse(ctx, t.acceptLabel)
	if err != nil {
		return ComponentSpec{}, err
	}

	rejectMeta, err := t.det.Parse(ctx, t.rejectLabel)
	if err != nil {
		return ComponentSpec{}, err
	}

	// Cria botões de aceitar/rejeitar
	buttons := []Button{
		{
			Label:   TextValue{Raw: t.acceptLabel, Template: acceptMeta.IsTemplate, Liquid: acceptMeta},
			Payload: "accept",
			Kind:    "reply",
		},
		{
			Label:   TextValue{Raw: t.rejectLabel, Template: rejectMeta.IsTemplate, Liquid: rejectMeta},
			Payload: "reject",
			Kind:    "reply",
		},
	}

	// Metadados específicos do componente terms
	metaData := map[string]any{
		"component_type": "terms_acceptance",
		"output_mode":    "single", // Output único: selected
		"has_link":       t.linkURL != "",
	}

	if t.linkURL != "" {
		metaData["link_url"] = t.linkURL
	}

	return ComponentSpec{
		Kind:    "terms",
		Text:    textVal,
		Buttons: buttons,
		Meta:    metaData,
	}, nil
}

// TermsFactory factory para criar componentes terms
type TermsFactory struct{ det liquid.Detector }

func NewTermsFactory(det liquid.Detector) *TermsFactory {
	return &TermsFactory{det: det}
}

func (f *TermsFactory) New(_ string, props map[string]any) (Component, error) {
	t := NewTerms(f.det)

	// Configurar texto dos termos
	if text, ok := props["text"].(string); ok && text != "" {
		t = t.WithText(text)
	}

	// Configurar link (URL + texto do link)
	if linkURL, ok := props["link_url"].(string); ok && linkURL != "" {
		linkText, _ := props["link_text"].(string)
		t = t.WithLink(linkURL, linkText)
	}

	// Configurar labels dos botões
	acceptLabel, _ := props["accept_label"].(string)
	rejectLabel, _ := props["reject_label"].(string)
	if acceptLabel != "" || rejectLabel != "" {
		t = t.WithLabels(acceptLabel, rejectLabel)
	}

	// Backward compatibility: accept_text/reject_text
	if acceptText, ok := props["accept_text"].(string); ok && acceptText != "" && acceptLabel == "" {
		t = t.WithLabels(acceptText, "")
	}
	if rejectText, ok := props["reject_text"].(string); ok && rejectText != "" && rejectLabel == "" {
		currentAccept := t.acceptLabel
		t = t.WithLabels(currentAccept, rejectText)
	}

	// Parse behaviors
	behavior, err := ParseBehavior(props, f.det)
	if err != nil {
		return nil, err
	}

	// Parse persistence
	persistence, err := ParsePersistence(props)
	if err != nil {
		return nil, err
	}

	return &TermsWithBehaviorAndPersistence{
		terms:       t,
		behavior:    behavior,
		persistence: persistence,
	}, nil
}

// TermsWithBehaviorAndPersistence é um wrapper que inclui behaviors e persistência
type TermsWithBehaviorAndPersistence struct {
	terms       *Terms
	behavior    *ComponentBehavior
	persistence *persistence.Config
}

func (twbp *TermsWithBehaviorAndPersistence) Kind() string {
	return twbp.terms.Kind()
}

func (twbp *TermsWithBehaviorAndPersistence) Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) {
	spec, err := twbp.terms.Spec(ctx, rctx)
	if err != nil {
		return spec, err
	}

	spec.Behavior = twbp.behavior
	spec.Persistence = twbp.persistence

	return spec, nil
}
