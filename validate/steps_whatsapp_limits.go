package validate

import (
	"fmt"

	"github.com/AgendoCerto/lib-bot/io"
)

// WhatsAppLimitsStep valida limites do WhatsApp Business API
type WhatsAppLimitsStep struct{}

// NewWhatsAppLimitsStep cria novo validador de limites WhatsApp
func NewWhatsAppLimitsStep() *WhatsAppLimitsStep {
	return &WhatsAppLimitsStep{}
}

func (w *WhatsAppLimitsStep) ValidateDesign(design io.DesignDoc) []Issue {
	var issues []Issue

	for i, node := range design.Graph.Nodes {
		path := fmt.Sprintf("graph.nodes[%d]", i)

		// Valida componente buttons
		if node.Kind == "buttons" {
			issues = append(issues, w.validateButtons(path, node.Props)...)
		}

		// Valida componente listpicker
		if node.Kind == "listpicker" {
			issues = append(issues, w.validateListPicker(path, node.Props)...)
		}

		// Valida componente media
		if node.Kind == "media" {
			issues = append(issues, w.validateMedia(path, node.Props)...)
		}

		// Valida componente text ou message
		if node.Kind == "text" || node.Kind == "message" {
			issues = append(issues, w.validateText(path, node.Props)...)
		}
	}

	return issues
}

// validateButtons valida limites do tipo interactive.button
func (w *WhatsAppLimitsStep) validateButtons(basePath string, props map[string]any) []Issue {
	var issues []Issue

	// WhatsApp: máximo 3 botões
	if buttons, ok := props["buttons"].([]any); ok {
		if len(buttons) > 3 {
			issues = append(issues, Issue{
				Code:     "whatsapp.buttons.max_count",
				Severity: Err,
				Path:     basePath + ".props.buttons",
				Msg:      fmt.Sprintf("WhatsApp permite no máximo 3 botões (encontrado: %d)", len(buttons)),
			})
		}

		// Valida cada botão
		for i, btnRaw := range buttons {
			if btnMap, ok := btnRaw.(map[string]any); ok {
				// Label: máximo 20 caracteres
				if label, ok := btnMap["label"].(string); ok && len(label) > 20 {
					issues = append(issues, Issue{
						Code:     "whatsapp.button.label.max_length",
						Severity: Err,
						Path:     fmt.Sprintf("%s.props.buttons[%d].label", basePath, i),
						Msg:      fmt.Sprintf("Label deve ter no máximo 20 caracteres (atual: %d)", len(label)),
					})
				}
			}
		}
	}

	// Header: máximo 60 caracteres
	if header, ok := props["header"].(string); ok && len(header) > 60 {
		issues = append(issues, Issue{
			Code:     "whatsapp.header.max_length",
			Severity: Err,
			Path:     basePath + ".props.header",
			Msg:      fmt.Sprintf("Header deve ter no máximo 60 caracteres (atual: %d)", len(header)),
		})
	}

	// Footer: máximo 60 caracteres
	if footer, ok := props["footer"].(string); ok && len(footer) > 60 {
		issues = append(issues, Issue{
			Code:     "whatsapp.footer.max_length",
			Severity: Err,
			Path:     basePath + ".props.footer",
			Msg:      fmt.Sprintf("Footer deve ter no máximo 60 caracteres (atual: %d)", len(footer)),
		})
	}

	// Texto (body): máximo 1024 caracteres (interactive message)
	if text, ok := props["text"].(string); ok && len(text) > 1024 {
		issues = append(issues, Issue{
			Code:     "whatsapp.text.max_length",
			Severity: Err,
			Path:     basePath + ".props.text",
			Msg:      fmt.Sprintf("Texto deve ter no máximo 1024 caracteres (atual: %d)", len(text)),
		})
	}

	return issues
}

// validateListPicker valida limites do tipo interactive.list
func (w *WhatsAppLimitsStep) validateListPicker(basePath string, props map[string]any) []Issue {
	var issues []Issue

	// WhatsApp: máximo 10 seções
	if sections, ok := props["sections"].([]any); ok {
		if len(sections) > 10 {
			issues = append(issues, Issue{
				Code:     "whatsapp.sections.max_count",
				Severity: Err,
				Path:     basePath + ".props.sections",
				Msg:      fmt.Sprintf("WhatsApp permite no máximo 10 seções (encontrado: %d)", len(sections)),
			})
		}

		totalItems := 0
		for i, sectionRaw := range sections {
			if sectionMap, ok := sectionRaw.(map[string]any); ok {
				sectionPath := fmt.Sprintf("%s.props.sections[%d]", basePath, i)

				// Título da seção: máximo 24 caracteres
				if title, ok := sectionMap["title"].(string); ok && len(title) > 24 {
					issues = append(issues, Issue{
						Code:     "whatsapp.section.title.max_length",
						Severity: Err,
						Path:     sectionPath + ".title",
						Msg:      fmt.Sprintf("Título da seção deve ter no máximo 24 caracteres (atual: %d)", len(title)),
					})
				}

				// Itens da seção: máximo 10 por seção
				if items, ok := sectionMap["items"].([]any); ok {
					if len(items) > 10 {
						issues = append(issues, Issue{
							Code:     "whatsapp.section.items.max_count",
							Severity: Err,
							Path:     sectionPath + ".items",
							Msg:      fmt.Sprintf("Máximo 10 itens por seção (encontrado: %d)", len(items)),
						})
					}

					totalItems += len(items)

					// Valida cada item
					for j, itemRaw := range items {
						if itemMap, ok := itemRaw.(map[string]any); ok {
							itemPath := fmt.Sprintf("%s.items[%d]", sectionPath, j)

							// Título do item: máximo 24 caracteres
							if title, ok := itemMap["title"].(string); ok && len(title) > 24 {
								issues = append(issues, Issue{
									Code:     "whatsapp.item.title.max_length",
									Severity: Err,
									Path:     itemPath + ".title",
									Msg:      fmt.Sprintf("Título do item deve ter no máximo 24 caracteres (atual: %d)", len(title)),
								})
							}

							// Descrição do item: máximo 72 caracteres
							if desc, ok := itemMap["description"].(string); ok && len(desc) > 72 {
								issues = append(issues, Issue{
									Code:     "whatsapp.item.description.max_length",
									Severity: Err,
									Path:     itemPath + ".description",
									Msg:      fmt.Sprintf("Descrição do item deve ter no máximo 72 caracteres (atual: %d)", len(desc)),
								})
							}
						}
					}
				}
			}
		}

		// Total de itens: máximo 90-100 (vamos usar 90 para ser conservador)
		if totalItems > 90 {
			issues = append(issues, Issue{
				Code:     "whatsapp.list.total_items.max_count",
				Severity: Warn,
				Path:     basePath + ".props.sections",
				Msg:      fmt.Sprintf("Total de itens muito alto: %d (WhatsApp recomenda no máximo 90)", totalItems),
			})
		}
	}

	// Button text: máximo 20 caracteres
	if btnText, ok := props["button_text"].(string); ok && len(btnText) > 20 {
		issues = append(issues, Issue{
			Code:     "whatsapp.button_text.max_length",
			Severity: Err,
			Path:     basePath + ".props.button_text",
			Msg:      fmt.Sprintf("Texto do botão deve ter no máximo 20 caracteres (atual: %d)", len(btnText)),
		})
	}

	// Header: máximo 60 caracteres
	if header, ok := props["header"].(string); ok && len(header) > 60 {
		issues = append(issues, Issue{
			Code:     "whatsapp.header.max_length",
			Severity: Err,
			Path:     basePath + ".props.header",
			Msg:      fmt.Sprintf("Header deve ter no máximo 60 caracteres (atual: %d)", len(header)),
		})
	}

	// Footer: máximo 60 caracteres
	if footer, ok := props["footer"].(string); ok && len(footer) > 60 {
		issues = append(issues, Issue{
			Code:     "whatsapp.footer.max_length",
			Severity: Err,
			Path:     basePath + ".props.footer",
			Msg:      fmt.Sprintf("Footer deve ter no máximo 60 caracteres (atual: %d)", len(footer)),
		})
	}

	// Texto (body): máximo 1024 caracteres (interactive message)
	if text, ok := props["text"].(string); ok && len(text) > 1024 {
		issues = append(issues, Issue{
			Code:     "whatsapp.text.max_length",
			Severity: Err,
			Path:     basePath + ".props.text",
			Msg:      fmt.Sprintf("Texto deve ter no máximo 1024 caracteres (atual: %d)", len(text)),
		})
	}

	return issues
}

// validateMedia valida limites dos tipos media
func (w *WhatsAppLimitsStep) validateMedia(basePath string, props map[string]any) []Issue {
	var issues []Issue

	mediaType, _ := props["type"].(string)

	// Caption: máximo 1024 caracteres
	if caption, ok := props["caption"].(string); ok && len(caption) > 1024 {
		issues = append(issues, Issue{
			Code:     "whatsapp.caption.max_length",
			Severity: Err,
			Path:     basePath + ".props.caption",
			Msg:      fmt.Sprintf("Caption deve ter no máximo 1024 caracteres (atual: %d)", len(caption)),
		})
	}

	// Filename: obrigatório para document
	if mediaType == "document" {
		if filename, ok := props["filename"].(string); !ok || filename == "" {
			issues = append(issues, Issue{
				Code:     "whatsapp.document.filename.required",
				Severity: Err,
				Path:     basePath + ".props.filename",
				Msg:      "Documento requer filename",
			})
		}
	}

	return issues
}

// validateText valida limites do tipo text
func (w *WhatsAppLimitsStep) validateText(basePath string, props map[string]any) []Issue {
	var issues []Issue

	// Texto: máximo 4096 caracteres
	if text, ok := props["text"].(string); ok && len(text) > 4096 {
		issues = append(issues, Issue{
			Code:     "whatsapp.text.max_length",
			Severity: Err,
			Path:     basePath + ".props.text",
			Msg:      fmt.Sprintf("Texto deve ter no máximo 4096 caracteres (atual: %d)", len(text)),
		})
	}

	return issues
}
