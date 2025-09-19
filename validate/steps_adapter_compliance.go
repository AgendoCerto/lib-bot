package validate

import (
	"fmt"
	"regexp"
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

// containsRichText detecta se o texto contém formatação rica após remover templates Liquid
func containsRichText(text string) bool {
	// Remove todos os templates Liquid usando regex para maior eficiência
	cleanText := removeLiquidTemplatesAdvanced(text)

	// Verifica marcadores de rich text em ordem de frequência para otimização
	richMarkers := []richTextPattern{
		// Markdown básico
		{"**", "bold"},          // **bold**
		{"*", "italic"},         // *italic*
		{"_", "italic/bold"},    // _italic_ ou __bold__
		{"`", "code"},           // `code`
		{"~~", "strikethrough"}, // ~~strikethrough~~

		// Markdown links e referências
		{"[", "link_start"},   // [link](url)
		{"](", "link_middle"}, // [text](url)

		// Headers markdown
		{"# ", "header1"},    // # Header
		{"## ", "header2"},   // ## Header
		{"### ", "header3"},  // ### Header
		{"#### ", "header4"}, // #### Header

		// Outros marcadores comuns
		{"> ", "blockquote"},     // > quote
		{"- ", "list"},           // - item
		{"* ", "list"},           // * item
		{"+ ", "list"},           // + item
		{"1. ", "numbered_list"}, // 1. item
	}

	// Verifica cada marcador no texto limpo
	for _, marker := range richMarkers {
		if containsRichTextMarker(cleanText, marker.pattern) {
			return true
		}
	}

	return false
}

// richTextPattern define um padrão de texto rico
type richTextPattern struct {
	pattern string
	name    string
}

// containsRichTextMarker verifica se um marcador específico está presente de forma válida
func containsRichTextMarker(text, marker string) bool {
	if !strings.Contains(text, marker) {
		return false
	}

	// Verificações contextuais para reduzir falsos positivos
	switch {
	case marker == "*" || marker == "**":
		// Verifica se não é apenas um asterisco isolado em matemática
		return hasValidAsteriskFormatting(text)
	case marker == "* ":
		// Verifica se é realmente uma lista e não matemática
		return hasValidListFormatting(text)
	case marker == "_" || marker == "__":
		// Verifica se não é apenas underscore em identificadores
		return hasValidUnderscoreFormatting(text)
	case strings.HasPrefix(marker, "# "):
		// Verifica se é realmente um header e não um hashtag
		return hasValidHeaderFormatting(text)
	case marker == "[" || marker == "](":
		// Verifica se faz parte de uma estrutura de link válida
		return hasValidLinkFormatting(text)
	default:
		return true
	}
}

// hasValidAsteriskFormatting verifica se asteriscos são para formatação, não matemática
func hasValidAsteriskFormatting(text string) bool {
	// Primeiro, verifica se há expressões matemáticas óbvias
	mathPatterns := []string{
		`\d+\s*\*\s*\d+`, // 2 * 3
		`\d+\s*\*\s*\$`,  // 5 * $
		`\$\s*\*\s*\d+`,  // $ * 2
		`\w+\s*\*\s*\w+`, // var * var
	}

	for _, pattern := range mathPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			// Se encontrou padrão matemático, não considera como formatação
			return false
		}
	}

	// Procura por padrões **texto** ou *texto* (formatação)
	formatPatterns := []string{
		`\*\*[^*]+\*\*`, // **bold**
		`\*[^*]{2,}\*`,  // *italic* (pelo menos 2 caracteres)
	}

	for _, pattern := range formatPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			return true
		}
	}
	return false
}

// hasValidUnderscoreFormatting verifica se underscores são para formatação
func hasValidUnderscoreFormatting(text string) bool {
	// Procura por padrões __texto__ ou _texto_ (evitando identificadores como user_id)
	patterns := []string{
		`__[^_]+__`,       // __bold__
		`_[^_\s\w][^_]*_`, // _italic_ (não inicia com letra/número para evitar identificadores)
		`_[^_]*[^\w]_`,    // _italic_ (não termina com letra/número)
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			return true
		}
	}
	return false
}

// hasValidHeaderFormatting verifica se # é um header markdown
func hasValidHeaderFormatting(text string) bool {
	// Procura por padrões # Texto no início de linha
	patterns := []string{
		`^#{1,6}\s+.+`,  // # Header no início
		`\n#{1,6}\s+.+`, // # Header após quebra de linha
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			return true
		}
	}
	return false
}

// hasValidLinkFormatting verifica se [ faz parte de um link markdown
func hasValidLinkFormatting(text string) bool {
	// Procura por padrões [texto](url) ou [texto][ref]
	patterns := []string{
		`\[[^\]]+\]\([^)]+\)`,  // [texto](url)
		`\[[^\]]+\]\[[^\]]*\]`, // [texto][ref]
		`\[[^\]]+\]:`,          // [ref]: definição
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			return true
		}
	}
	return false
}

// hasValidListFormatting verifica se * é um marcador de lista
func hasValidListFormatting(text string) bool {
	// Procura por padrões de lista no início de linha
	patterns := []string{
		`^\*\s+`,  // * item no início
		`\n\*\s+`, // * item após quebra de linha
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			return true
		}
	}
	return false
}

// removeLiquidTemplatesAdvanced remove templates Liquid usando regex otimizada
func removeLiquidTemplatesAdvanced(text string) string {
	// Regex para templates Liquid mais robusta
	liquidPatterns := []*regexp.Regexp{
		// {{ variavel }}, {{ variavel | filtro }}, {{ variavel.propriedade }}
		regexp.MustCompile(`\{\{\s*[^}]+\s*\}\}`),

		// {% tag %}, {% if %}, {% for %}, etc.
		regexp.MustCompile(`\{%\s*[^%]+\s*%\}`),

		// {%- tag -%} (whitespace control)
		regexp.MustCompile(`\{%-?\s*[^%]+\s*-?%\}`),

		// {{- variavel -}} (whitespace control)
		regexp.MustCompile(`\{\{-?\s*[^}]+\s*-?\}\}`),
	}

	result := text
	for _, pattern := range liquidPatterns {
		result = pattern.ReplaceAllString(result, "")
	}

	return strings.TrimSpace(result)
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
