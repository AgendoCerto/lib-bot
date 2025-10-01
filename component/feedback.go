package component

import (
	"context"

	"github.com/AgendoCerto/lib-bot/liquid"
	"github.com/AgendoCerto/lib-bot/persistence"
	"github.com/AgendoCerto/lib-bot/runtime"
)

// Feedback representa o componente de coleta de feedback/avaliação
// Por baixo dos panos funciona como texto, mas é especializado para avaliações
type Feedback struct {
	text        string          // Texto da pergunta de feedback
	placeholder string          // Placeholder para entrada do usuário
	scale       string          // Tipo de escala (ex: "1-5", "1-10", "emoji", "text")
	det         liquid.Detector // Detector para parsing de templates Liquid
}

// NewFeedback cria nova instância de componente feedback
func NewFeedback(det liquid.Detector) *Feedback {
	return &Feedback{
		det:         det,
		scale:       "text", // Padrão: texto livre
		placeholder: "Digite sua avaliação ou comentário...",
	}
}

func (f *Feedback) Kind() string { return "feedback" }

// WithText define o texto da pergunta de feedback
func (f *Feedback) WithText(s string) *Feedback {
	cp := *f
	cp.text = s
	return &cp
}

// WithPlaceholder define o placeholder
func (f *Feedback) WithPlaceholder(s string) *Feedback {
	cp := *f
	cp.placeholder = s
	return &cp
}

// WithScale define o tipo de escala de avaliação
func (f *Feedback) WithScale(s string) *Feedback {
	cp := *f
	cp.scale = s
	return &cp
}

// Spec gera o ComponentSpec para feedback (funciona como texto por baixo dos panos)
func (f *Feedback) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	// Processa o texto principal com templates Liquid
	var textVal *TextValue
	if f.text != "" {
		meta, err := f.det.Parse(ctx, f.text)
		if err != nil {
			return ComponentSpec{}, err
		}
		textVal = &TextValue{
			Raw:      f.text,
			Template: meta.IsTemplate,
			Liquid:   meta,
		}
	}

	// Metadados específicos do componente feedback
	metaData := map[string]any{
		"component_type": "feedback_collection",
		"scale":          f.scale,
		"placeholder":    f.placeholder,
		"input_type":     "text",       // Por baixo dos panos é texto
		"purpose":        "evaluation", // Indica que é para avaliação
	}

	return ComponentSpec{
		Kind: "feedback",
		Text: textVal,
		Meta: metaData,
	}, nil
}

// FeedbackFactory factory para criar componentes feedback
type FeedbackFactory struct{ det liquid.Detector }

func NewFeedbackFactory(det liquid.Detector) *FeedbackFactory {
	return &FeedbackFactory{det: det}
}

func (f *FeedbackFactory) New(_ string, props map[string]any) (Component, error) {
	fb := NewFeedback(f.det)

	// Configurar texto da pergunta
	if text, ok := props["text"].(string); ok && text != "" {
		fb = fb.WithText(text)
	}

	// Configurar placeholder
	if placeholder, ok := props["placeholder"].(string); ok && placeholder != "" {
		fb = fb.WithPlaceholder(placeholder)
	}

	// Configurar escala
	if scale, ok := props["scale"].(string); ok && scale != "" {
		fb = fb.WithScale(scale)
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

	return &FeedbackWithBehaviorAndPersistence{
		feedback:    fb,
		behavior:    behavior,
		persistence: persistence,
	}, nil
}

// FeedbackWithBehaviorAndPersistence é um wrapper que inclui behaviors e persistência
type FeedbackWithBehaviorAndPersistence struct {
	feedback    *Feedback
	behavior    *ComponentBehavior
	persistence *persistence.Config
}

func (fwbp *FeedbackWithBehaviorAndPersistence) Kind() string {
	return fwbp.feedback.Kind()
}

func (fwbp *FeedbackWithBehaviorAndPersistence) Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) {
	spec, err := fwbp.feedback.Spec(ctx, rctx)
	if err != nil {
		return spec, err
	}

	spec.Behavior = fwbp.behavior
	spec.Persistence = fwbp.persistence

	return spec, nil
}
