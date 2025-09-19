package validate

import (
	"fmt"
	"strings"

	"lib-bot/adapter"
	"lib-bot/adapter/whatsapp"
	"lib-bot/component"
	"lib-bot/flow"
	"lib-bot/io"
)

// AdapterComplianceStep valida se specs estão em conformidade com capabilities do adapter
type AdapterComplianceStep struct{}

// NewAdapterComplianceStep cria novo validador de conformidade
func NewAdapterComplianceStep() *AdapterComplianceStep {
	return &AdapterComplianceStep{}
}

func (s *AdapterComplianceStep) Check(spec component.ComponentSpec, caps adapter.Capabilities, path string) []Issue {
	var issues []Issue

	// Valida conformidade baseada no tipo de componente
	switch spec.Kind {
	case "buttons":
		issues = append(issues, s.validateButtons(spec, caps, path)...)
	case "listpicker", "menu":
		issues = append(issues, s.validateListPicker(spec, caps, path)...)
	case "carousel":
		issues = append(issues, s.validateCarousel(spec, caps, path)...)
	case "media":
		issues = append(issues, s.validateMedia(spec, caps, path)...)
	case "message":
		issues = append(issues, s.validateMessage(spec, caps, path)...)
	}

	// Validações gerais
	issues = append(issues, s.validateGeneral(spec, caps, path)...)

	return issues
}

// validateButtons valida componentes de botões
func (s *AdapterComplianceStep) validateButtons(spec component.ComponentSpec, caps adapter.Capabilities, path string) []Issue {
	var issues []Issue

	// Verifica limite de botões
	if len(spec.Buttons) > caps.MaxButtons {
		issues = append(issues, Issue{
			Code: "adapter.buttons.limit_exceeded", Severity: Err,
			Path: path + ".buttons",
			Msg:  fmt.Sprintf("too many buttons: %d, max allowed: %d", len(spec.Buttons), caps.MaxButtons),
		})
	}

	// Valida tipos de botão suportados
	for i, btn := range spec.Buttons {
		btnPath := fmt.Sprintf("%s.buttons[%d]", path, i)

		if !caps.ButtonKinds[btn.Kind] {
			issues = append(issues, Issue{
				Code: "adapter.button.unsupported_kind", Severity: Err,
				Path: btnPath + ".kind",
				Msg:  fmt.Sprintf("button kind '%s' not supported by adapter", btn.Kind),
			})
		}

		// Valida tamanho do título do botão
		if len(btn.Label.Raw) > caps.MaxButtonTitleLen {
			issues = append(issues, Issue{
				Code: "adapter.button.title_too_long", Severity: Warn,
				Path: btnPath + ".label",
				Msg:  fmt.Sprintf("button title too long: %d chars, max: %d (will be truncated)", len(btn.Label.Raw), caps.MaxButtonTitleLen),
			})
		}
	}

	return issues
}

// validateListPicker valida componentes de lista
func (s *AdapterComplianceStep) validateListPicker(spec component.ComponentSpec, caps adapter.Capabilities, path string) []Issue {
	var issues []Issue

	if !caps.SupportsListPicker {
		issues = append(issues, Issue{
			Code: "adapter.listpicker.unsupported", Severity: Err,
			Path: path,
			Msg:  "list picker not supported by adapter",
		})
		return issues
	}

	// Valida seções e itens se presentes no metadata
	if sections, ok := spec.Meta["sections"].([]any); ok {
		if len(sections) > caps.MaxListSections {
			issues = append(issues, Issue{
				Code: "adapter.listpicker.too_many_sections", Severity: Warn,
				Path: path + ".sections",
				Msg:  fmt.Sprintf("too many sections: %d, max: %d (will be truncated)", len(sections), caps.MaxListSections),
			})
		}

		// Valida itens por seção
		for i, sectionRaw := range sections {
			if sectionMap, ok := sectionRaw.(map[string]any); ok {
				if items, ok := sectionMap["items"].([]any); ok {
					if len(items) > caps.MaxListItems {
						issues = append(issues, Issue{
							Code: "adapter.listpicker.too_many_items", Severity: Warn,
							Path: fmt.Sprintf("%s.sections[%d].items", path, i),
							Msg:  fmt.Sprintf("too many items in section: %d, max: %d (will be truncated)", len(items), caps.MaxListItems),
						})
					}
				}
			}
		}
	}

	return issues
}

// validateCarousel valida componentes de carrossel
func (s *AdapterComplianceStep) validateCarousel(spec component.ComponentSpec, caps adapter.Capabilities, path string) []Issue {
	var issues []Issue

	if !caps.SupportsCarousel {
		// WhatsApp suporta carrossel via product_list, não é erro fatal
		issues = append(issues, Issue{
			Code: "adapter.carousel.limited_support", Severity: Info,
			Path: path,
			Msg:  "carousel will be mapped to product list format",
		})
	}

	// Valida cards se presentes
	if cards, ok := spec.Meta["cards"].([]any); ok {
		// WhatsApp limita número de produtos em product_list
		maxCards := 30 // Limite típico do WhatsApp para product_list
		if len(cards) > maxCards {
			issues = append(issues, Issue{
				Code: "adapter.carousel.too_many_cards", Severity: Warn,
				Path: path + ".cards",
				Msg:  fmt.Sprintf("too many cards: %d, recommended max: %d", len(cards), maxCards),
			})
		}

		// Valida botões em cada card
		for i, cardRaw := range cards {
			if cardMap, ok := cardRaw.(map[string]any); ok {
				if buttons, ok := cardMap["buttons"].([]any); ok {
					if len(buttons) > caps.MaxButtons {
						issues = append(issues, Issue{
							Code: "adapter.carousel.card_too_many_buttons", Severity: Warn,
							Path: fmt.Sprintf("%s.cards[%d].buttons", path, i),
							Msg:  fmt.Sprintf("too many buttons in card: %d, max: %d", len(buttons), caps.MaxButtons),
						})
					}
				}
			}
		}
	}

	return issues
}

// validateMedia valida componentes de mídia
func (s *AdapterComplianceStep) validateMedia(spec component.ComponentSpec, caps adapter.Capabilities, path string) []Issue {
	var issues []Issue

	// Valida se tem URL de mídia
	if spec.MediaURL == "" {
		issues = append(issues, Issue{
			Code: "adapter.media.missing_url", Severity: Err,
			Path: path + ".media_url",
			Msg:  "media component requires media_url",
		})
	}

	// Valida caption se presente
	if spec.Text != nil && len(spec.Text.Raw) > caps.MaxTextLen {
		issues = append(issues, Issue{
			Code: "adapter.media.caption_too_long", Severity: Warn,
			Path: path + ".caption",
			Msg:  fmt.Sprintf("media caption too long: %d chars, max: %d (will be truncated)", len(spec.Text.Raw), caps.MaxTextLen),
		})
	}

	return issues
}

// validateMessage valida componentes de mensagem
func (s *AdapterComplianceStep) validateMessage(spec component.ComponentSpec, caps adapter.Capabilities, path string) []Issue {
	var issues []Issue

	// Valida HSM
	if spec.HSM != nil {
		if !caps.SupportsHSM {
			issues = append(issues, Issue{
				Code: "adapter.message.hsm_unsupported", Severity: Err,
				Path: path + ".hsm",
				Msg:  "HSM templates not supported by adapter",
			})
		}
	}

	// Valida texto
	if spec.Text != nil && len(spec.Text.Raw) > caps.MaxTextLen {
		issues = append(issues, Issue{
			Code: "adapter.message.text_too_long", Severity: Warn,
			Path: path + ".text",
			Msg:  fmt.Sprintf("message text too long: %d chars, max: %d (will be truncated)", len(spec.Text.Raw), caps.MaxTextLen),
		})
	}

	return issues
}

// validateGeneral valida aspectos gerais de qualquer componente
func (s *AdapterComplianceStep) validateGeneral(spec component.ComponentSpec, caps adapter.Capabilities, path string) []Issue {
	var issues []Issue

	// Valida texto rico se não suportado
	if spec.Text != nil && !caps.SupportsRichText {
		if containsRichText(spec.Text.Raw) {
			issues = append(issues, Issue{
				Code: "adapter.general.rich_text_unsupported", Severity: Warn,
				Path: path + ".text",
				Msg:  "rich text formatting will be stripped",
			})
		}
	}

	// Valida campos de metadata específicos
	if spec.Meta != nil {
		// Valida header
		if header, ok := spec.Meta["header"].(string); ok && len(header) > caps.MaxHeaderLen {
			issues = append(issues, Issue{
				Code: "adapter.general.header_too_long", Severity: Warn,
				Path: path + ".header",
				Msg:  fmt.Sprintf("header too long: %d chars, max: %d (will be truncated)", len(header), caps.MaxHeaderLen),
			})
		}

		// Valida footer
		if footer, ok := spec.Meta["footer"].(string); ok && len(footer) > caps.MaxFooterLen {
			issues = append(issues, Issue{
				Code: "adapter.general.footer_too_long", Severity: Warn,
				Path: path + ".footer",
				Msg:  fmt.Sprintf("footer too long: %d chars, max: %d (will be truncated)", len(footer), caps.MaxFooterLen),
			})
		}

		// Valida description
		if desc, ok := spec.Meta["description"].(string); ok && len(desc) > caps.MaxDescriptionLen {
			issues = append(issues, Issue{
				Code: "adapter.general.description_too_long", Severity: Warn,
				Path: path + ".description",
				Msg:  fmt.Sprintf("description too long: %d chars, max: %d (will be truncated)", len(desc), caps.MaxDescriptionLen),
			})
		}
	}

	return issues
}

// containsRichText detecta se o texto contém formatação rica
func containsRichText(text string) bool {
	// Remove templates Liquid antes de verificar rich text
	cleanText := removeLiquidTemplates(text)

	// Verifica marcadores comuns de texto rico
	richMarkers := []string{"**", "*", "_", "`", "~~", "[", "](", "##", "###"}
	for _, marker := range richMarkers {
		if strings.Contains(cleanText, marker) {
			return true
		}
	}
	return false
}

// removeLiquidTemplates remove templates Liquid do texto para análise de rich text
func removeLiquidTemplates(text string) string {
	// Remove {{ }} templates
	result := text
	for {
		start := strings.Index(result, "{{")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "}}")
		if end == -1 {
			break
		}
		end += start + 2
		result = result[:start] + result[end:]
	}

	// Remove {% %} templates
	for {
		start := strings.Index(result, "{%")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "%}")
		if end == -1 {
			break
		}
		end += start + 2
		result = result[:start] + result[end:]
	}

	return result
}

// ValidateDesign implementa DesignValidator interface
func (s *AdapterComplianceStep) ValidateDesign(design io.DesignDoc) []Issue {
	var issues []Issue

	// Cria adapter WhatsApp para obter capabilities
	whatsappAdapter := whatsapp.New()
	caps := whatsappAdapter.Capabilities()

	// Para cada nó no grafo, valida conformidade com adapter
	for i, node := range design.Graph.Nodes {
		path := fmt.Sprintf("graph.nodes[%d]", i)

		// Cria spec do componente baseado no nó
		spec := s.createComponentSpec(node)

		// Valida usando validação de adapter
		nodeIssues := s.Check(spec, caps, path)
		issues = append(issues, nodeIssues...)
	}

	return issues
}

// createComponentSpec cria ComponentSpec a partir de um nó do design
func (s *AdapterComplianceStep) createComponentSpec(node flow.Node) component.ComponentSpec {
	spec := component.ComponentSpec{
		Kind: node.Kind,
		Meta: node.Props,
	}

	// Extrai campos específicos dos props conforme o tipo
	if node.Props != nil {
		switch node.Kind {
		case "media":
			if url, ok := node.Props["url"].(string); ok {
				spec.MediaURL = url
			}
		case "message":
			if text, ok := node.Props["text"].(string); ok {
				spec.Text = &component.TextValue{
					Raw: text,
				}
			}
		}
	}

	return spec
}
