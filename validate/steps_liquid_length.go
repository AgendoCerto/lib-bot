package validate

import (
	"fmt"
	"strings"

	"github.com/AgendoCerto/lib-bot/flow"
	"github.com/AgendoCerto/lib-bot/io"
)

// LiquidLengthStep valida limites de caracteres considerando expansão de templates Liquid
type LiquidLengthStep struct{}

// NewLiquidLengthStep cria novo validador de limites considerando Liquid
func NewLiquidLengthStep() *LiquidLengthStep {
	return &LiquidLengthStep{}
}

// ValidateDesign valida limites de caracteres no design completo
func (s *LiquidLengthStep) ValidateDesign(design io.DesignDoc) []Issue {
	var issues []Issue

	for i, node := range design.Graph.Nodes {
		path := fmt.Sprintf("graph.nodes[%d]", i)
		issues = append(issues, s.validateNodeTextLimits(node, path, design.Props)...)
	}

	return issues
}

// validateNodeTextLimits valida limites de texto de um nó considerando templates Liquid
func (s *LiquidLengthStep) validateNodeTextLimits(node flow.Node, path string, globalProps map[string]any) []Issue {
	var issues []Issue

	switch node.Kind {
	case "message", "text":
		if text := s.extractText(node.Props, "text"); text != "" {
			issues = append(issues, s.validateTextLength(text, path+".props.text", 1024, globalProps)...)
		}
	case "buttons":
		if text := s.extractText(node.Props, "text"); text != "" {
			issues = append(issues, s.validateTextLength(text, path+".props.text", 1024, globalProps)...)
		}
		// Valida labels dos botões
		if buttons, ok := node.Props["buttons"].([]interface{}); ok {
			for i, btn := range buttons {
				if btnMap, ok := btn.(map[string]interface{}); ok {
					if label := s.extractText(btnMap, "label"); label != "" {
						btnPath := fmt.Sprintf("%s.props.buttons[%d].label", path, i)
						issues = append(issues, s.validateTextLength(label, btnPath, 20, globalProps)...)
					}
				}
			}
		}
	case "listpicker":
		if text := s.extractText(node.Props, "text"); text != "" {
			issues = append(issues, s.validateTextLength(text, path+".props.text", 1024, globalProps)...)
		}
		// Valida títulos e descrições dos itens
		if sections, ok := node.Props["sections"].([]interface{}); ok {
			for secIdx, section := range sections {
				if secMap, ok := section.(map[string]interface{}); ok {
					if items, ok := secMap["items"].([]interface{}); ok {
						for itemIdx, item := range items {
							if itemMap, ok := item.(map[string]interface{}); ok {
								if title := s.extractText(itemMap, "title"); title != "" {
									titlePath := fmt.Sprintf("%s.props.sections[%d].items[%d].title", path, secIdx, itemIdx)
									issues = append(issues, s.validateTextLength(title, titlePath, 24, globalProps)...)
								}
								if desc := s.extractText(itemMap, "description"); desc != "" {
									descPath := fmt.Sprintf("%s.props.sections[%d].items[%d].description", path, secIdx, itemIdx)
									issues = append(issues, s.validateTextLength(desc, descPath, 72, globalProps)...)
								}
							}
						}
					}
				}
			}
		}
	case "media":
		if caption := s.extractText(node.Props, "caption"); caption != "" {
			issues = append(issues, s.validateTextLength(caption, path+".props.caption", 1024, globalProps)...)
		}
	case "carousel":
		if text := s.extractText(node.Props, "text"); text != "" {
			issues = append(issues, s.validateTextLength(text, path+".props.text", 1024, globalProps)...)
		}
		// Valida cards do carousel
		if cards, ok := node.Props["cards"].([]interface{}); ok {
			for cardIdx, card := range cards {
				if cardMap, ok := card.(map[string]interface{}); ok {
					if title := s.extractText(cardMap, "title"); title != "" {
						titlePath := fmt.Sprintf("%s.props.cards[%d].title", path, cardIdx)
						issues = append(issues, s.validateTextLength(title, titlePath, 80, globalProps)...)
					}
					if desc := s.extractText(cardMap, "description"); desc != "" {
						descPath := fmt.Sprintf("%s.props.cards[%d].description", path, cardIdx)
						issues = append(issues, s.validateTextLength(desc, descPath, 300, globalProps)...)
					}
				}
			}
		}
	case "confirm":
		if title := s.extractText(node.Props, "title"); title != "" {
			issues = append(issues, s.validateTextLength(title, path+".props.title", 60, globalProps)...)
		}
		if desc := s.extractText(node.Props, "description"); desc != "" {
			issues = append(issues, s.validateTextLength(desc, path+".props.description", 72, globalProps)...)
		}
	}

	return issues
}

// extractText extrai texto de uma propriedade
func (s *LiquidLengthStep) extractText(props map[string]any, key string) string {
	if val, ok := props[key].(string); ok {
		return val
	}
	return ""
}

// validateTextLength valida o comprimento do texto considerando expansão de templates Liquid
func (s *LiquidLengthStep) validateTextLength(text string, path string, maxLen int, globalProps map[string]any) []Issue {
	var issues []Issue

	// Calcula o tamanho base (sem templates)
	staticLen := len(text)

	// Detecta templates Liquid e estima expansão
	liquidTemplates := s.findLiquidTemplates(text)
	estimatedExpansion := s.estimateExpansion(liquidTemplates, globalProps)

	// Tamanho estimado final
	estimatedLen := staticLen + estimatedExpansion

	// Se não há templates, validação direta
	if len(liquidTemplates) == 0 {
		if staticLen > maxLen {
			issues = append(issues, Issue{
				Code: "liquid.length.exceeded", Severity: Err,
				Path: path,
				Msg:  fmt.Sprintf("text exceeds maximum length: %d chars, max: %d", staticLen, maxLen),
			})
		}
		return issues
	}

	// Com templates Liquid, diferentes níveis de validação
	if estimatedLen > maxLen {
		if staticLen > maxLen {
			// Já excede apenas com texto estático
			issues = append(issues, Issue{
				Code: "liquid.length.critical", Severity: Err,
				Path: path,
				Msg:  fmt.Sprintf("CRITICAL: static text already exceeds limit: %d chars, max: %d (before template expansion)", staticLen, maxLen),
			})
		} else {
			// Pode exceder após expansão dos templates
			issues = append(issues, Issue{
				Code: "liquid.length.estimated_exceeded", Severity: Warn,
				Path: path,
				Msg:  fmt.Sprintf("WARNING: estimated length may exceed limit: ~%d chars, max: %d (static: %d + estimated expansion: %d)", estimatedLen, maxLen, staticLen, estimatedExpansion),
			})
		}
	} else if estimatedLen > int(float64(maxLen)*0.8) { // 80% do limite
		// Aviso preventivo quando próximo do limite
		issues = append(issues, Issue{
			Code: "liquid.length.approaching_limit", Severity: Info,
			Path: path,
			Msg:  fmt.Sprintf("INFO: estimated length approaching limit: ~%d chars, max: %d (static: %d + estimated expansion: %d)", estimatedLen, maxLen, staticLen, estimatedExpansion),
		})
	}

	return issues
}

// findLiquidTemplates encontra templates Liquid no texto
func (s *LiquidLengthStep) findLiquidTemplates(text string) []string {
	var templates []string

	// Procura por padrões {{ variable }}
	start := 0
	for {
		startIdx := strings.Index(text[start:], "{{")
		if startIdx == -1 {
			break
		}
		startIdx += start

		endIdx := strings.Index(text[startIdx:], "}}")
		if endIdx == -1 {
			break
		}
		endIdx += startIdx + 2

		template := text[startIdx:endIdx]
		templates = append(templates, template)
		start = endIdx
	}

	// Procura por padrões {% tag %}
	start = 0
	for {
		startIdx := strings.Index(text[start:], "{%")
		if startIdx == -1 {
			break
		}
		startIdx += start

		endIdx := strings.Index(text[startIdx:], "%}")
		if endIdx == -1 {
			break
		}
		endIdx += startIdx + 2

		template := text[startIdx:endIdx]
		templates = append(templates, template)
		start = endIdx
	}

	return templates
}

// estimateExpansion estima a expansão dos templates baseado nas propriedades globais
func (s *LiquidLengthStep) estimateExpansion(templates []string, globalProps map[string]any) int {
	totalExpansion := 0

	for _, template := range templates {
		expansion := s.estimateTemplateExpansion(template, globalProps)
		totalExpansion += expansion
	}

	return totalExpansion
}

// estimateTemplateExpansion estima a expansão de um template específico
func (s *LiquidLengthStep) estimateTemplateExpansion(template string, globalProps map[string]any) int {
	// Remove {{ }} ou {% %}
	template = strings.TrimSpace(template)
	template = strings.TrimPrefix(template, "{{")
	template = strings.TrimPrefix(template, "{%")
	template = strings.TrimSuffix(template, "}}")
	template = strings.TrimSuffix(template, "%}")
	template = strings.TrimSpace(template)

	// Estimativas baseadas em padrões conhecidos
	switch {
	case strings.Contains(template, "props.clinica_nome"):
		return 30 // Nome da clínica estimado
	case strings.Contains(template, "props.telefone_emergencia"):
		return 18 // Telefone formato +55 11 9999-9999
	case strings.Contains(template, "props.horario_funcionamento"):
		return 35 // Horário de funcionamento
	case strings.Contains(template, "user.name"):
		return 15 // Nome do usuário estimado
	case strings.Contains(template, "flow."):
		return 20 // Variáveis de fluxo genéricas
	case strings.Contains(template, "#each"):
		return 100 // Loops podem expandir muito
	case strings.Contains(template, "data"):
		return 10 // Data formato DD/MM/YYYY
	case strings.Contains(template, "horario"):
		return 5 // Horário formato HH:MM
	default:
		return 10 // Expansão padrão conservadora
	}
}
