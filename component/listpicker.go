package component

import (
	"context"

	"lib-bot/liquid"
	"lib-bot/runtime"
)

// ListPicker componente para listas de seleção (menu interativo)
type ListPicker struct {
	text       string          // Texto principal
	buttonText string          // Texto do botão da lista
	sections   []SectionData   // Seções da lista
	det        liquid.Detector // Detector para parsing de templates Liquid
}

// SectionData representa uma seção da lista
type SectionData struct {
	Title string     `json:"title"` // Título da seção
	Items []ItemData `json:"items"` // Itens da seção
}

// ItemData representa um item da lista
type ItemData struct {
	ID          string `json:"id"`                    // ID único do item
	Title       string `json:"title"`                 // Título do item
	Description string `json:"description,omitempty"` // Descrição opcional
}

// NewListPicker cria nova instância de componente listpicker
func NewListPicker(det liquid.Detector) *ListPicker {
	return &ListPicker{
		det:        det,
		sections:   make([]SectionData, 0),
		buttonText: "Ver opções", // Default
	}
}

// Kind retorna o tipo do componente
func (l *ListPicker) Kind() string { return "listpicker" }

// WithText define o texto principal
func (l *ListPicker) WithText(s string) *ListPicker {
	cp := *l
	cp.text = s
	return &cp
}

// WithButtonText define o texto do botão da lista
func (l *ListPicker) WithButtonText(s string) *ListPicker {
	cp := *l
	cp.buttonText = s
	return &cp
}

// AddSection adiciona uma seção com itens
func (l *ListPicker) AddSection(title string, items []ItemData) *ListPicker {
	cp := *l
	cp.sections = append(cp.sections, SectionData{
		Title: title,
		Items: items,
	})
	return &cp
}

// Spec gera o ComponentSpec com parsing de templates
func (l *ListPicker) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	// Processa texto principal
	var textVal *TextValue
	if l.text != "" {
		meta, err := l.det.Parse(ctx, l.text)
		if err != nil {
			return ComponentSpec{}, err
		}
		textVal = &TextValue{
			Raw:      l.text,
			Template: meta.IsTemplate,
			Liquid:   meta,
		}
	}

	// Processa texto do botão
	btnMeta, err := l.det.Parse(ctx, l.buttonText)
	if err != nil {
		return ComponentSpec{}, err
	}

	// Cria metadata específica para lista
	meta := map[string]any{
		"button_text": TextValue{
			Raw:      l.buttonText,
			Template: btnMeta.IsTemplate,
			Liquid:   btnMeta,
		},
		"sections": l.sections,
	}

	return ComponentSpec{
		Kind: "listpicker",
		Text: textVal,
		Meta: meta,
	}, nil
}

// Factory

type ListPickerFactory struct{ det liquid.Detector }

func NewListPickerFactory(det liquid.Detector) *ListPickerFactory {
	return &ListPickerFactory{det: det}
}

func (f *ListPickerFactory) New(_ string, props map[string]any) (Component, error) {
	l := NewListPicker(f.det)

	// Texto principal
	if text, _ := props["text"].(string); text != "" {
		l = l.WithText(text)
	}

	// Texto do botão
	if btnText, _ := props["button_text"].(string); btnText != "" {
		l = l.WithButtonText(btnText)
	}

	// Seções
	if sectionsRaw, ok := props["sections"].([]any); ok {
		for _, sectionRaw := range sectionsRaw {
			if sectionMap, ok := sectionRaw.(map[string]any); ok {
				title, _ := sectionMap["title"].(string)

				var items []ItemData
				if itemsRaw, ok := sectionMap["items"].([]any); ok {
					for _, itemRaw := range itemsRaw {
						if itemMap, ok := itemRaw.(map[string]any); ok {
							id, _ := itemMap["id"].(string)
							itemTitle, _ := itemMap["title"].(string)
							desc, _ := itemMap["description"].(string)

							items = append(items, ItemData{
								ID:          id,
								Title:       itemTitle,
								Description: desc,
							})
						}
					}
				}

				l = l.AddSection(title, items)
			}
		}
	}

	return l, nil
}
