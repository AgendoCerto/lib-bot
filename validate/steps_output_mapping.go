package validate

import (
	"fmt"

	"lib-bot/flow"
	"lib-bot/io"
)

// OutputMappingStep valida que todos os outputs mapeiam corretamente para elementos interativos
type OutputMappingStep struct{}

// NewOutputMappingStep cria novo validador de mapeamento de outputs
func NewOutputMappingStep() *OutputMappingStep {
	return &OutputMappingStep{}
}

// ValidateDesign valida mapeamento de outputs no design completo
func (s *OutputMappingStep) ValidateDesign(design io.DesignDoc) []Issue {
	var issues []Issue

	for i, node := range design.Graph.Nodes {
		path := fmt.Sprintf("graph.nodes[%d]", i)
		issues = append(issues, s.validateNodeOutputs(node, path)...)
	}

	return issues
}

// validateNodeOutputs valida que outputs do nó mapeiam corretamente para elementos interativos
func (s *OutputMappingStep) validateNodeOutputs(node flow.Node, path string) []Issue {
	var issues []Issue

	switch node.Kind {
	case "buttons":
		issues = append(issues, s.validateButtonOutputs(node, path)...)
	case "listpicker":
		issues = append(issues, s.validateListPickerOutputs(node, path)...)
	case "carousel":
		issues = append(issues, s.validateCarouselOutputs(node, path)...)
	case "confirm":
		issues = append(issues, s.validateConfirmOutputs(node, path)...)
	case "message", "media", "delay", "text":
		issues = append(issues, s.validateNonInteractiveOutputs(node, path)...)
	default:
		issues = append(issues, Issue{
			Code: "output.unknown_component", Severity: Err,
			Path: path + ".kind",
			Msg:  fmt.Sprintf("unknown component kind: %s", node.Kind),
		})
	}

	return issues
}

// validateButtonOutputs valida que outputs de buttons mapeiam para IDs dos botões
func (s *OutputMappingStep) validateButtonOutputs(node flow.Node, path string) []Issue {
	var issues []Issue

	// Extrai IDs dos botões das propriedades
	buttonIDs := s.extractButtonIDs(node.Props)
	if len(buttonIDs) == 0 {
		issues = append(issues, Issue{
			Code: "output.buttons.no_buttons", Severity: Err,
			Path: path + ".props.buttons",
			Msg:  "buttons component must have at least one button",
		})
		return issues
	}

	// Verifica limite do WhatsApp (máximo 3 botões)
	if len(buttonIDs) > 3 {
		issues = append(issues, Issue{
			Code: "output.buttons.too_many_buttons", Severity: Err,
			Path: path + ".props.buttons",
			Msg:  fmt.Sprintf("WhatsApp adapter supports maximum 3 buttons, got %d", len(buttonIDs)),
		})
	}

	outputs := node.Outputs
	standardOutputs := []string{"timeout", "invalid", "fallback"}

	// CRÍTICO: Cada botão deve ter um output correspondente
	for _, buttonID := range buttonIDs {
		if !contains(outputs, buttonID) {
			issues = append(issues, Issue{
				Code: "output.buttons.missing_button_output", Severity: Err,
				Path: path + ".outputs",
				Msg:  fmt.Sprintf("CRITICAL: missing output for button ID '%s' - this will cause engine failure", buttonID),
			})
		}
	}

	// Verifica outputs obrigatórios para robustez
	for _, required := range standardOutputs {
		if !contains(outputs, required) {
			issues = append(issues, Issue{
				Code: "output.buttons.missing_standard", Severity: Warn,
				Path: path + ".outputs",
				Msg:  fmt.Sprintf("missing standard output '%s' - recommended for fallback handling", required),
			})
		}
	}

	// CRÍTICO: Verifica outputs inválidos que causarão falha na engine
	for _, output := range outputs {
		if !contains(buttonIDs, output) && !contains(standardOutputs, output) {
			issues = append(issues, Issue{
				Code: "output.buttons.invalid_output", Severity: Err,
				Path: path + ".outputs",
				Msg:  fmt.Sprintf("CRITICAL: output '%s' does not map to any button ID or standard output - engine will fail", output),
			})
		}
	}

	return issues
}

// validateListPickerOutputs valida que outputs de listpicker mapeiam para IDs dos itens
func (s *OutputMappingStep) validateListPickerOutputs(node flow.Node, path string) []Issue {
	var issues []Issue

	// Extrai IDs dos itens da lista
	itemIDs := s.extractListItemIDs(node.Props)
	if len(itemIDs) == 0 {
		issues = append(issues, Issue{
			Code: "output.listpicker.no_items", Severity: Err,
			Path: path + ".props.sections",
			Msg:  "listpicker component must have at least one item",
		})
		return issues
	}

	outputs := node.Outputs
	standardOutputs := []string{"timeout", "invalid", "cancelled"}

	// CRÍTICO: Cada item deve ter um output correspondente
	for _, itemID := range itemIDs {
		if !contains(outputs, itemID) {
			issues = append(issues, Issue{
				Code: "output.listpicker.missing_item_output", Severity: Err,
				Path: path + ".outputs",
				Msg:  fmt.Sprintf("CRITICAL: missing output for list item ID '%s' - this will cause engine failure", itemID),
			})
		}
	}

	// Verifica outputs obrigatórios
	for _, required := range standardOutputs {
		if !contains(outputs, required) {
			issues = append(issues, Issue{
				Code: "output.listpicker.missing_standard", Severity: Warn,
				Path: path + ".outputs",
				Msg:  fmt.Sprintf("missing standard output '%s' - recommended for error handling", required),
			})
		}
	}

	// CRÍTICO: Verifica outputs inválidos
	for _, output := range outputs {
		if !contains(itemIDs, output) && !contains(standardOutputs, output) {
			issues = append(issues, Issue{
				Code: "output.listpicker.invalid_output", Severity: Err,
				Path: path + ".outputs",
				Msg:  fmt.Sprintf("CRITICAL: output '%s' does not map to any list item ID or standard output - engine will fail", output),
			})
		}
	}

	return issues
}

// validateCarouselOutputs valida que outputs de carousel mapeiam para IDs dos botões dos cards
func (s *OutputMappingStep) validateCarouselOutputs(node flow.Node, path string) []Issue {
	var issues []Issue

	// Extrai IDs dos botões de todos os cards
	carouselButtonIDs := s.extractCarouselButtonIDs(node.Props)
	if len(carouselButtonIDs) == 0 {
		issues = append(issues, Issue{
			Code: "output.carousel.no_buttons", Severity: Warn,
			Path: path + ".props.cards",
			Msg:  "carousel component has no interactive buttons",
		})
	}

	outputs := node.Outputs
	standardOutputs := []string{"timeout", "invalid"}

	// CRÍTICO: Cada botão de card deve ter um output correspondente
	for _, buttonID := range carouselButtonIDs {
		if !contains(outputs, buttonID) {
			issues = append(issues, Issue{
				Code: "output.carousel.missing_button_output", Severity: Err,
				Path: path + ".outputs",
				Msg:  fmt.Sprintf("CRITICAL: missing output for carousel button ID '%s' - this will cause engine failure", buttonID),
			})
		}
	}

	// Verifica outputs obrigatórios
	for _, required := range standardOutputs {
		if !contains(outputs, required) {
			issues = append(issues, Issue{
				Code: "output.carousel.missing_standard", Severity: Warn,
				Path: path + ".outputs",
				Msg:  fmt.Sprintf("missing standard output '%s' - recommended for error handling", required),
			})
		}
	}

	return issues
}

// validateConfirmOutputs valida que outputs de confirm mapeiam para botões yes/no
func (s *OutputMappingStep) validateConfirmOutputs(node flow.Node, path string) []Issue {
	var issues []Issue

	// Para confirm, verifica se os IDs dos botões estão presentes
	confirmButtonIDs := s.extractConfirmButtonIDs(node.Props)
	expectedButtons := 2

	if len(confirmButtonIDs) == 0 {
		// Se não há IDs, usa comportamento padrão
		expectedOutputs := []string{"confirmed", "cancelled", "timeout"}
		outputs := node.Outputs

		for _, required := range expectedOutputs {
			if !contains(outputs, required) {
				issues = append(issues, Issue{
					Code: "output.confirm.missing_default_output", Severity: Err,
					Path: path + ".outputs",
					Msg:  fmt.Sprintf("CRITICAL: confirm component missing default output '%s'", required),
				})
			}
		}
	} else {
		// Se há IDs específicos, deve mapear para eles
		if len(confirmButtonIDs) != expectedButtons {
			issues = append(issues, Issue{
				Code: "output.confirm.invalid_buttons", Severity: Err,
				Path: path + ".props",
				Msg:  fmt.Sprintf("confirm component must have exactly %d buttons (yes/no), got %d", expectedButtons, len(confirmButtonIDs)),
			})
		}

		outputs := node.Outputs
		standardOutputs := []string{"timeout"}

		// CRÍTICO: Cada botão confirm deve ter um output correspondente
		for _, buttonID := range confirmButtonIDs {
			if !contains(outputs, buttonID) {
				issues = append(issues, Issue{
					Code: "output.confirm.missing_button_output", Severity: Err,
					Path: path + ".outputs",
					Msg:  fmt.Sprintf("CRITICAL: missing output for confirm button ID '%s' - this will cause engine failure", buttonID),
				})
			}
		}

		// Verifica outputs obrigatórios
		for _, required := range standardOutputs {
			if !contains(outputs, required) {
				issues = append(issues, Issue{
					Code: "output.confirm.missing_standard", Severity: Warn,
					Path: path + ".outputs",
					Msg:  fmt.Sprintf("missing standard output '%s' - recommended for timeout handling", required),
				})
			}
		}
	}

	return issues
}

// validateNonInteractiveOutputs valida componentes não-interativos (message, media, delay)
func (s *OutputMappingStep) validateNonInteractiveOutputs(node flow.Node, path string) []Issue {
	var issues []Issue

	outputs := node.Outputs

	switch node.Kind {
	case "message":
		expectedOutputs := []string{"complete"}
		if !outputsMatch(outputs, expectedOutputs) {
			issues = append(issues, Issue{
				Code: "output.message.invalid_outputs", Severity: Err,
				Path: path + ".outputs",
				Msg:  fmt.Sprintf("CRITICAL: message component must have outputs: %v, got: %v", expectedOutputs, outputs),
			})
		}
	case "text":
		expectedOutputs := []string{"sent", "failed", "timeout"}
		for _, required := range expectedOutputs {
			if !contains(outputs, required) {
				issues = append(issues, Issue{
					Code: "output.text.missing_output", Severity: Err,
					Path: path + ".outputs",
					Msg:  fmt.Sprintf("CRITICAL: text component missing required output '%s'", required),
				})
			}
		}
	case "media":
		expectedOutputs := []string{"sent", "failed", "timeout"}
		for _, required := range expectedOutputs {
			if !contains(outputs, required) {
				issues = append(issues, Issue{
					Code: "output.media.missing_output", Severity: Err,
					Path: path + ".outputs",
					Msg:  fmt.Sprintf("CRITICAL: media component missing required output '%s'", required),
				})
			}
		}
	case "delay":
		expectedOutputs := []string{"complete"}
		if !outputsMatch(outputs, expectedOutputs) {
			issues = append(issues, Issue{
				Code: "output.delay.invalid_outputs", Severity: Err,
				Path: path + ".outputs",
				Msg:  fmt.Sprintf("CRITICAL: delay component must have outputs: %v, got: %v", expectedOutputs, outputs),
			})
		}
	}

	return issues
}

// Funções auxiliares para extrair IDs dos elementos

func (s *OutputMappingStep) extractButtonIDs(props map[string]any) []string {
	var buttonIDs []string

	if buttons, ok := props["buttons"].([]interface{}); ok {
		for _, btn := range buttons {
			if btnMap, ok := btn.(map[string]interface{}); ok {
				if id, ok := btnMap["id"].(string); ok && id != "" {
					buttonIDs = append(buttonIDs, id)
				}
			}
		}
	}

	return buttonIDs
}

func (s *OutputMappingStep) extractListItemIDs(props map[string]any) []string {
	var itemIDs []string

	if sections, ok := props["sections"].([]interface{}); ok {
		for _, section := range sections {
			if sectionMap, ok := section.(map[string]interface{}); ok {
				if items, ok := sectionMap["items"].([]interface{}); ok {
					for _, item := range items {
						if itemMap, ok := item.(map[string]interface{}); ok {
							if id, ok := itemMap["id"].(string); ok && id != "" {
								itemIDs = append(itemIDs, id)
							}
						}
					}
				}
			}
		}
	}

	return itemIDs
}

func (s *OutputMappingStep) extractCarouselButtonIDs(props map[string]any) []string {
	var buttonIDs []string

	if cards, ok := props["cards"].([]interface{}); ok {
		for _, card := range cards {
			if cardMap, ok := card.(map[string]interface{}); ok {
				if buttons, ok := cardMap["buttons"].([]interface{}); ok {
					for _, btn := range buttons {
						if btnMap, ok := btn.(map[string]interface{}); ok {
							if id, ok := btnMap["id"].(string); ok && id != "" {
								buttonIDs = append(buttonIDs, id)
							}
						}
					}
				}
			}
		}
	}

	return buttonIDs
}

func (s *OutputMappingStep) extractConfirmButtonIDs(props map[string]any) []string {
	var buttonIDs []string

	// Para confirm, verificar se existem IDs nos botões yes/no
	if yes, ok := props["yes"].(map[string]interface{}); ok {
		if id, ok := yes["id"].(string); ok && id != "" {
			buttonIDs = append(buttonIDs, id)
		}
	}

	if no, ok := props["no"].(map[string]interface{}); ok {
		if id, ok := no["id"].(string); ok && id != "" {
			buttonIDs = append(buttonIDs, id)
		}
	}

	return buttonIDs
}

// Funções utilitárias

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func outputsMatch(actual, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}

	for _, exp := range expected {
		if !contains(actual, exp) {
			return false
		}
	}

	return true
}
