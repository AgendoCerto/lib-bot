package component

import (
	"context"

	"github.com/AgendoCerto/lib-bot/liquid"
	"github.com/AgendoCerto/lib-bot/runtime"
)

// ListPicker componente para listas de seleção (menu interativo)
type ListPicker struct {
	text       string          // Texto principal
	buttonText string          // Texto do botão da lista
	sections   []SectionData   // Seções da lista
	header     string          // WhatsApp: header opcional (≤60 chars)
	footer     string          // WhatsApp: footer opcional (≤60 chars)
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

// WithHeader define o header (WhatsApp: ≤60 chars)
func (l *ListPicker) WithHeader(s string) *ListPicker {
	cp := *l
	cp.header = s
	return &cp
}

// WithFooter define o footer (WhatsApp: ≤60 chars)
func (l *ListPicker) WithFooter(s string) *ListPicker {
	cp := *l
	cp.footer = s
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

	// Cria metadata específica para lista com spec v2.2
	meta := map[string]any{
		"output_mode": "single", // v2.2: usa output único "selected" ao invés de item IDs
		"button_text": TextValue{
			Raw:      l.buttonText,
			Template: btnMeta.IsTemplate,
			Liquid:   btnMeta,
		},
		"sections": l.sections,
	}

	// WhatsApp: header e footer opcionais
	if l.header != "" {
		meta["header"] = l.header
	}
	if l.footer != "" {
		meta["footer"] = l.footer
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

	// WhatsApp: header e footer opcionais
	if header, _ := props["header"].(string); header != "" {
		l = l.WithHeader(header)
	}
	if footer, _ := props["footer"].(string); footer != "" {
		l = l.WithFooter(footer)
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

	// Parse behaviors
	behavior, err := ParseBehavior(props, f.det)
	if err != nil {
		return nil, err
	}

	return &ListPickerWithBehavior{
		listPicker: l,
		behavior:   behavior,
	}, nil
}

// ListPickerWithBehavior é um wrapper que inclui behaviors
type ListPickerWithBehavior struct {
	listPicker *ListPicker
	behavior   *ComponentBehavior
}

func (lpwb *ListPickerWithBehavior) Kind() string {
	return lpwb.listPicker.Kind()
}

func (lpwb *ListPickerWithBehavior) Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) {
	spec, err := lpwb.listPicker.Spec(ctx, rctx)
	if err != nil {
		return spec, err
	}
	spec.Behavior = lpwb.behavior
	return spec, nil
}
