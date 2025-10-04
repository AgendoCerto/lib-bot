package validate

import (
	"fmt"

	"github.com/AgendoCerto/lib-bot/flow"
	"github.com/AgendoCerto/lib-bot/io"
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
	case "message", "media", "delay":
		issues = append(issues, s.validateNonInteractiveOutputs(node, path)...)
	case "terms":
		issues = append(issues, s.validateTermsOutputs(node, path)...)
	case "feedback":
		issues = append(issues, s.validateFeedbackOutputs(node, path)...)
	case "global_start":
		issues = append(issues, s.validateGlobalStartOutputs(node, path)...)
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

	// SPEC V2.2: Suporta output único "selected" OU outputs individuais (v2.1)
	hasSelectedOutput := contains(outputs, "selected")

	if hasSelectedOutput {
		// Modo v2.2: output único "selected" é válido
		// Não precisa de outputs individuais

		// Verifica se tem outputs extras além de "selected" e standardOutputs
		validOutputsV22 := append([]string{"selected"}, standardOutputs...)
		for _, output := range outputs {
			if !contains(validOutputsV22, output) {
				issues = append(issues, Issue{
					Code: "output.buttons.invalid_output_v22", Severity: Warn,
					Path: path + ".outputs",
					Msg:  fmt.Sprintf("buttons (v2.2) has unexpected output '%s' - only 'selected' and standard outputs are valid", output),
				})
			}
		}
	} else {
		// Modo v2.1: CRÍTICO: Cada botão deve ter um output correspondente
		for _, buttonID := range buttonIDs {
			if !contains(outputs, buttonID) {
				issues = append(issues, Issue{
					Code: "output.buttons.missing_button_output", Severity: Err,
					Path: path + ".outputs",
					Msg:  fmt.Sprintf("CRITICAL: missing output for button ID '%s' - this will cause engine failure", buttonID),
				})
			}
		}

		// Verifica se outputs extras são válidos (permite outputs padrão)
		validOutputs := append(buttonIDs, standardOutputs...)
		for _, output := range outputs {
			if !contains(validOutputs, output) {
				issues = append(issues, Issue{
					Code: "output.buttons.invalid_output", Severity: Warn,
					Path: path + ".outputs",
					Msg:  fmt.Sprintf("buttons component has unexpected output '%s'", output),
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
	standardOutputs := []string{"timeout", "invalid", "fallback"}

	// SPEC V2.2: Suporta output único "selected" OU outputs individuais (v2.1)
	hasSelectedOutput := contains(outputs, "selected")

	if hasSelectedOutput {
		// Modo v2.2: output único "selected" é válido
		// Não precisa de outputs individuais

		// Verifica se tem outputs extras além de "selected" e standardOutputs
		validOutputsV22 := append([]string{"selected"}, standardOutputs...)
		for _, output := range outputs {
			if !contains(validOutputsV22, output) {
				issues = append(issues, Issue{
					Code: "output.listpicker.invalid_output_v22", Severity: Warn,
					Path: path + ".outputs",
					Msg:  fmt.Sprintf("listpicker (v2.2) has unexpected output '%s' - only 'selected' and standard outputs are valid", output),
				})
			}
		}
	} else {
		// Modo v2.1: CRÍTICO: Cada item deve ter um output correspondente
		for _, itemID := range itemIDs {
			if !contains(outputs, itemID) {
				issues = append(issues, Issue{
					Code: "output.listpicker.missing_item_output", Severity: Err,
					Path: path + ".outputs",
					Msg:  fmt.Sprintf("CRITICAL: missing output for list item ID '%s' - this will cause engine failure", itemID),
				})
			}
		}

		// Verificar se outputs extras são válidos (permite outputs padrão)
		validOutputs := append(itemIDs, standardOutputs...)
		for _, output := range outputs {
			if !contains(validOutputs, output) {
				issues = append(issues, Issue{
					Code: "output.listpicker.invalid_output", Severity: Warn,
					Path: path + ".outputs",
					Msg:  fmt.Sprintf("listpicker component has unexpected output '%s'", output),
				})
			}
		}

		// CRÍTICO: Verifica outputs inválidos que causarão falha na engine
		for _, output := range outputs {
			if !contains(itemIDs, output) && !contains(standardOutputs, output) {
				issues = append(issues, Issue{
					Code: "output.listpicker.invalid_output", Severity: Err,
					Path: path + ".outputs",
					Msg:  fmt.Sprintf("CRITICAL: output '%s' does not map to any list item ID or standard output - engine will fail", output),
				})
			}
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
	standardOutputs := []string{"timeout", "invalid", "fallback"}

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

	// Verificar se outputs extras são válidos (permite outputs padrão)
	validOutputs := append(carouselButtonIDs, standardOutputs...)
	validOutputs = append(validOutputs, "complete") // carousel pode ter complete também
	for _, output := range outputs {
		if !contains(validOutputs, output) {
			issues = append(issues, Issue{
				Code: "output.carousel.invalid_output", Severity: Warn,
				Path: path + ".outputs",
				Msg:  fmt.Sprintf("carousel component has unexpected output '%s'", output),
			})
		}
	}

	return issues
}

// validateNonInteractiveOutputs valida componentes não-interativos (message, media, delay)
func (s *OutputMappingStep) validateNonInteractiveOutputs(node flow.Node, path string) []Issue {
	var issues []Issue

	outputs := node.Outputs
	standardOutputs := []string{"timeout", "invalid", "fallback"}

	switch node.Kind {
	case "message":
		// message pode ter "complete" OU "response" dependendo do behavior.await
		// Se await.enabled=true → usa "response" (aguarda resposta do usuário)
		// Se await.enabled=false → usa "complete" (apenas envia)

		// Extrair await do behavior
		hasAwait := false
		props := node.Props
		if props != nil {
			if behavior, ok := props["behavior"].(map[string]interface{}); ok {
				if awaitCfg, ok := behavior["await"].(map[string]interface{}); ok {
					if enabled, ok := awaitCfg["enabled"].(bool); ok {
						hasAwait = enabled
					}
				}
			}
		}

		if hasAwait {
			// Modo interativo: espera resposta
			requiredOutputs := []string{"response"}
			for _, required := range requiredOutputs {
				if !contains(outputs, required) {
					issues = append(issues, Issue{
						Code: "output.message.missing_required", Severity: Err,
						Path: path + ".outputs",
						Msg:  fmt.Sprintf("CRITICAL: message component with await=true missing required output '%s'", required),
					})
				}
			}
			// Verificar se outputs extras são válidos
			validOutputs := append(requiredOutputs, standardOutputs...)
			for _, output := range outputs {
				if !contains(validOutputs, output) {
					// Se tem "complete" mas deveria ter "response", avisar
					if output == "complete" {
						issues = append(issues, Issue{
							Code: "output.message.invalid_output", Severity: Warn,
							Path: path + ".outputs",
							Msg:  "message component with await=true should use 'response' output instead of 'complete'",
						})
					} else {
						issues = append(issues, Issue{
							Code: "output.message.invalid_output", Severity: Warn,
							Path: path + ".outputs",
							Msg:  fmt.Sprintf("message component has unexpected output '%s'", output),
						})
					}
				}
			}
		} else {
			// Modo não-interativo: apenas envia
			requiredOutputs := []string{"complete"}
			for _, required := range requiredOutputs {
				if !contains(outputs, required) {
					issues = append(issues, Issue{
						Code: "output.message.missing_required", Severity: Err,
						Path: path + ".outputs",
						Msg:  fmt.Sprintf("CRITICAL: message component missing required output '%s'", required),
					})
				}
			}
			// Verificar se outputs extras são válidos
			validOutputs := append(requiredOutputs, standardOutputs...)
			for _, output := range outputs {
				if !contains(validOutputs, output) {
					// Se tem "response" mas deveria ter "complete", avisar
					if output == "response" {
						issues = append(issues, Issue{
							Code: "output.message.invalid_output", Severity: Warn,
							Path: path + ".outputs",
							Msg:  "message component with await=false should use 'complete' output instead of 'response'",
						})
					} else {
						issues = append(issues, Issue{
							Code: "output.message.invalid_output", Severity: Warn,
							Path: path + ".outputs",
							Msg:  fmt.Sprintf("message component has unexpected output '%s'", output),
						})
					}
				}
			}
		}
	case "media":
		// media deve ter "sent" + outputs padrão opcionais
		requiredOutputs := []string{"sent"}
		for _, required := range requiredOutputs {
			if !contains(outputs, required) {
				issues = append(issues, Issue{
					Code: "output.media.missing_required", Severity: Err,
					Path: path + ".outputs",
					Msg:  fmt.Sprintf("CRITICAL: media component missing required output '%s'", required),
				})
			}
		}
		// Verificar se outputs extras são válidos
		validOutputs := append(requiredOutputs, standardOutputs...)
		for _, output := range outputs {
			if !contains(validOutputs, output) {
				issues = append(issues, Issue{
					Code: "output.media.invalid_output", Severity: Warn,
					Path: path + ".outputs",
					Msg:  fmt.Sprintf("media component has unexpected output '%s'", output),
				})
			}
		}
	case "delay":
		// delay deve ter "complete" + outputs padrão opcionais
		requiredOutputs := []string{"complete"}
		for _, required := range requiredOutputs {
			if !contains(outputs, required) {
				issues = append(issues, Issue{
					Code: "output.delay.missing_required", Severity: Err,
					Path: path + ".outputs",
					Msg:  fmt.Sprintf("CRITICAL: delay component missing required output '%s'", required),
				})
			}
		}
		// Verificar se outputs extras são válidos
		validOutputs := append(requiredOutputs, standardOutputs...)
		for _, output := range outputs {
			if !contains(validOutputs, output) {
				issues = append(issues, Issue{
					Code: "output.delay.invalid_output", Severity: Warn,
					Path: path + ".outputs",
					Msg:  fmt.Sprintf("delay component has unexpected output '%s'", output),
				})
			}
		}
	}

	return issues
}

// validateTermsOutputs valida outputs do componente terms
func (s *OutputMappingStep) validateTermsOutputs(node flow.Node, path string) []Issue {
	var issues []Issue

	outputs := node.Outputs
	standardOutputs := []string{"timeout", "invalid", "fallback"}
	requiredOutputs := []string{"accepted", "rejected"}

	// CRÍTICO: terms deve ter outputs de aceito e rejeitado
	for _, required := range requiredOutputs {
		if !contains(outputs, required) {
			issues = append(issues, Issue{
				Code: "output.terms.missing_required", Severity: Err,
				Path: path + ".outputs",
				Msg:  fmt.Sprintf("CRITICAL: terms component missing required output '%s'", required),
			})
		}
	}

	// Verificar se outputs extras são válidos (permite outputs padrão)
	validOutputs := append(requiredOutputs, standardOutputs...)
	for _, output := range outputs {
		if !contains(validOutputs, output) {
			issues = append(issues, Issue{
				Code: "output.terms.invalid_output", Severity: Warn,
				Path: path + ".outputs",
				Msg:  fmt.Sprintf("terms component has unexpected output '%s'", output),
			})
		}
	}

	return issues
}

// validateFeedbackOutputs valida outputs do componente feedback
func (s *OutputMappingStep) validateFeedbackOutputs(node flow.Node, path string) []Issue {
	var issues []Issue

	outputs := node.Outputs
	standardOutputs := []string{"timeout", "invalid", "fallback"}
	requiredOutputs := []string{"submitted"}

	// CRÍTICO: feedback deve ter output de submissão
	for _, required := range requiredOutputs {
		if !contains(outputs, required) {
			issues = append(issues, Issue{
				Code: "output.feedback.missing_required", Severity: Err,
				Path: path + ".outputs",
				Msg:  fmt.Sprintf("CRITICAL: feedback component missing required output '%s'", required),
			})
		}
	}

	// Verificar se outputs extras são válidos (permite outputs padrão)
	validOutputs := append(requiredOutputs, standardOutputs...)
	for _, output := range outputs {
		if !contains(validOutputs, output) {
			issues = append(issues, Issue{
				Code: "output.feedback.invalid_output", Severity: Warn,
				Path: path + ".outputs",
				Msg:  fmt.Sprintf("feedback component has unexpected output '%s'", output),
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
				// Suporta tanto "id" (v2.1) quanto "payload" (v2.2)
				var id string
				if idVal, ok := btnMap["id"].(string); ok && idVal != "" {
					id = idVal
				} else if payloadVal, ok := btnMap["payload"].(string); ok && payloadVal != "" {
					id = payloadVal
				}

				if id != "" {
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

// Funções utilitárias

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// validateGlobalStartOutputs valida outputs do componente global_start
func (s *OutputMappingStep) validateGlobalStartOutputs(node flow.Node, path string) []Issue {
	var issues []Issue

	outputs := node.Outputs
	expectedOutputs := []string{"start"}

	// global_start deve ter exatamente o output "start"
	if !outputsMatch(outputs, expectedOutputs) {
		issues = append(issues, Issue{
			Code: "output.global_start.invalid_outputs", Severity: Err,
			Path: path + ".outputs",
			Msg:  fmt.Sprintf("global_start must have exactly ['start'] output, got %v", outputs),
		})
	}

	return issues
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
