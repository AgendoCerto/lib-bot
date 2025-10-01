package component

import (
	"context"

	"github.com/AgendoCerto/lib-bot/liquid"
	"github.com/AgendoCerto/lib-bot/persistence"
	"github.com/AgendoCerto/lib-bot/runtime"
)

// Terms representa o componente de aceite de termos
// Por baixo dos panos funciona como texto, mas tem outputs específicos para aceito/não aceito
type Terms struct {
	text        string          // Texto dos termos a serem aceitos
	acceptText  string          // Texto para aceitar (ex: "Aceito", "Sim", "OK")
	rejectText  string          // Texto para rejeitar (ex: "Não aceito", "Não", "Cancelar")
	placeholder string          // Placeholder para entrada do usuário
	det         liquid.Detector // Detector para parsing de templates Liquid
}

// NewTerms cria nova instância de componente terms
func NewTerms(det liquid.Detector) *Terms {
	return &Terms{
		det:         det,
		acceptText:  "Aceito",
		rejectText:  "Não aceito",
		placeholder: "Digite 'Aceito' para concordar ou 'Não aceito' para recusar",
	}
}

func (t *Terms) Kind() string { return "terms" }

// WithText define o texto dos termos
func (t *Terms) WithText(s string) *Terms {
	cp := *t
	cp.text = s
	return &cp
}

// WithAcceptText define o texto de aceitação
func (t *Terms) WithAcceptText(s string) *Terms {
	cp := *t
	cp.acceptText = s
	return &cp
}

// WithRejectText define o texto de rejeição
func (t *Terms) WithRejectText(s string) *Terms {
	cp := *t
	cp.rejectText = s
	return &cp
}

// WithPlaceholder define o placeholder
func (t *Terms) WithPlaceholder(s string) *Terms {
	cp := *t
	cp.placeholder = s
	return &cp
}

// Spec gera o ComponentSpec para terms (funciona como texto por baixo dos panos)
func (t *Terms) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	// Processa o texto principal com templates Liquid
	var textVal *TextValue
	if t.text != "" {
		meta, err := t.det.Parse(ctx, t.text)
		if err != nil {
			return ComponentSpec{}, err
		}
		textVal = &TextValue{
			Raw:      t.text,
			Template: meta.IsTemplate,
			Liquid:   meta,
		}
	}

	// Metadados específicos do componente terms
	metaData := map[string]any{
		"component_type": "terms_acceptance",
		"accept_text":    t.acceptText,
		"reject_text":    t.rejectText,
		"placeholder":    t.placeholder,
		"input_type":     "text", // Por baixo dos panos é texto
	}

	return ComponentSpec{
		Kind: "terms",
		Text: textVal,
		Meta: metaData,
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

	// Configurar texto de aceitação
	if acceptText, ok := props["accept_text"].(string); ok && acceptText != "" {
		t = t.WithAcceptText(acceptText)
	}

	// Configurar texto de rejeição
	if rejectText, ok := props["reject_text"].(string); ok && rejectText != "" {
		t = t.WithRejectText(rejectText)
	}

	// Configurar placeholder
	if placeholder, ok := props["placeholder"].(string); ok && placeholder != "" {
		t = t.WithPlaceholder(placeholder)
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
