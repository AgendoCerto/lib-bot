package component

import (
	"context"

	"github.com/AgendoCerto/lib-bot/liquid"
	"github.com/AgendoCerto/lib-bot/runtime"
)

// TermsGate componente server-driven para aceite de termos (spec v2.2)
// Só mostra o termo se o backend determinar que é necessário
type TermsGate struct {
	versionID    string          // ID da versão dos termos (ex: "v2")
	text         string          // Texto dos termos
	remindAfterS int             // Lembrete automático após X segundos (opcional)
	det          liquid.Detector // Detector para parsing de templates Liquid
}

// NewTermsGate cria nova instância de componente terms_gate
func NewTermsGate(det liquid.Detector) *TermsGate {
	return &TermsGate{det: det}
}

func (tg *TermsGate) Kind() string { return "terms_gate" }

// WithVersionID define o ID da versão dos termos
func (tg *TermsGate) WithVersionID(id string) *TermsGate {
	cp := *tg
	cp.versionID = id
	return &cp
}

// WithText define o texto dos termos
func (tg *TermsGate) WithText(s string) *TermsGate {
	cp := *tg
	cp.text = s
	return &cp
}

// WithReminder define lembrete automático
func (tg *TermsGate) WithReminder(seconds int) *TermsGate {
	cp := *tg
	cp.remindAfterS = seconds
	return &cp
}

// Spec gera o ComponentSpec para terms_gate
func (tg *TermsGate) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	// Processa o texto com templates Liquid
	var textVal *TextValue
	if tg.text != "" {
		meta, err := tg.det.Parse(ctx, tg.text)
		if err != nil {
			return ComponentSpec{}, err
		}
		textVal = &TextValue{
			Raw:      tg.text,
			Template: meta.IsTemplate,
			Liquid:   meta,
		}
	}

	// Metadados específicos do terms_gate
	metaData := map[string]any{
		"version_id":     tg.versionID,
		"server_driven":  true, // Backend decide se mostra ou não
		"remind_after_s": tg.remindAfterS,
		"component_type": "terms_gate",
	}

	return ComponentSpec{
		Kind: "terms_gate",
		Text: textVal,
		Meta: metaData,
	}, nil
}

// TermsGateFactory factory para criar componentes terms_gate
type TermsGateFactory struct{ det liquid.Detector }

func NewTermsGateFactory(det liquid.Detector) *TermsGateFactory {
	return &TermsGateFactory{det: det}
}

func (f *TermsGateFactory) New(_ string, props map[string]any) (Component, error) {
	tg := NewTermsGate(f.det)

	// Configurar versão
	if versionID, ok := props["version_id"].(string); ok && versionID != "" {
		tg = tg.WithVersionID(versionID)
	}

	// Configurar texto
	if text, ok := props["text"].(string); ok && text != "" {
		tg = tg.WithText(text)
	}

	// Configurar lembrete
	if remindAfter, ok := props["remind_after_s"].(float64); ok {
		tg = tg.WithReminder(int(remindAfter))
	} else if remindAfter, ok := props["remind_after_s"].(int); ok {
		tg = tg.WithReminder(remindAfter)
	}

	// Parse behaviors
	behavior, err := ParseBehavior(props, f.det)
	if err != nil {
		return nil, err
	}

	return &TermsGateWithBehavior{
		termsGate: tg,
		behavior:  behavior,
	}, nil
}

// TermsGateWithBehavior wrapper que inclui behaviors
type TermsGateWithBehavior struct {
	termsGate *TermsGate
	behavior  *ComponentBehavior
}

func (tgwb *TermsGateWithBehavior) Kind() string {
	return tgwb.termsGate.Kind()
}

func (tgwb *TermsGateWithBehavior) Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) {
	spec, err := tgwb.termsGate.Spec(ctx, rctx)
	if err != nil {
		return spec, err
	}

	spec.Behavior = tgwb.behavior
	return spec, nil
}
