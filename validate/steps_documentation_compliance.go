package validate

import (
	"fmt"
	"strings"

	"github.com/AgendoCerto/lib-bot/flow"
	"github.com/AgendoCerto/lib-bot/io"
)

// DocumentationComplianceStep valida conformidade com especificações da documentação
type DocumentationComplianceStep struct{}

// NewDocumentationComplianceStep cria novo validador de conformidade com documentação
func NewDocumentationComplianceStep() *DocumentationComplianceStep {
	return &DocumentationComplianceStep{}
}

// ValidateDesign valida design completo contra especificações da documentação
func (s *DocumentationComplianceStep) ValidateDesign(design io.DesignDoc) []Issue {
	var issues []Issue

	// 1. Valida estrutura de versão
	issues = append(issues, s.validateVersion(design.Version)...)

	// 2. Valida entradas obrigatórias
	issues = append(issues, s.validateEntries(design.Entries)...)

	// 3. Valida bot e channels
	issues = append(issues, s.validateBot(design.Bot)...)

	// 4. Valida propriedades globais
	issues = append(issues, s.validateProps(design.Props)...)

	// 5. Valida estrutura do grafo
	issues = append(issues, s.validateGraphWithEntries(design.Graph, design.Entries)...)

	return issues
}

// validateVersion valida estrutura de versão conforme documentação
func (s *DocumentationComplianceStep) validateVersion(version io.Version) []Issue {
	var issues []Issue

	if version.ID == "" {
		issues = append(issues, Issue{
			Code: "doc.version.missing_id", Severity: Err,
			Path: "version.id",
			Msg:  "version ID is required",
		})
	}

	if version.Status != "development" && version.Status != "production" {
		issues = append(issues, Issue{
			Code: "doc.version.invalid_status", Severity: Err,
			Path: "version.status",
			Msg:  "version status must be 'development' or 'production'",
		})
	}

	return issues
}

// validateEntries valida pontos de entrada conforme documentação
func (s *DocumentationComplianceStep) validateEntries(entries []flow.Entry) []Issue {
	var issues []Issue

	if len(entries) == 0 {
		issues = append(issues, Issue{
			Code: "doc.entries.missing", Severity: Err,
			Path: "entries",
			Msg:  "at least one entry point is required",
		})
		return issues
	}

	hasGlobalStart := false
	for i, entry := range entries {
		path := fmt.Sprintf("entries[%d]", i)

		// Valida kind
		switch entry.Kind {
		case flow.EntryGlobalStart:
			hasGlobalStart = true
		case flow.EntryChannelStart:
			if entry.ChannelID == "" {
				issues = append(issues, Issue{
					Code: "doc.entries.channel_missing_id", Severity: Err,
					Path: path + ".channel_id",
					Msg:  "channel_start entry requires channel_id",
				})
			}
		case flow.EntryForced:
			// Entrada forçada é válida
		default:
			issues = append(issues, Issue{
				Code: "doc.entries.invalid_kind", Severity: Err,
				Path: path + ".kind",
				Msg:  fmt.Sprintf("invalid entry kind: %s", entry.Kind),
			})
		}

		// Valida target
		if entry.Target == "" {
			issues = append(issues, Issue{
				Code: "doc.entries.missing_target", Severity: Err,
				Path: path + ".target",
				Msg:  "entry target is required",
			})
		}
	}

	// Global start obrigatório
	if !hasGlobalStart {
		issues = append(issues, Issue{
			Code: "doc.entries.missing_global_start", Severity: Err,
			Path: "entries",
			Msg:  "global_start entry is required",
		})
	}

	return issues
}

// validateBot valida configuração do bot
func (s *DocumentationComplianceStep) validateBot(bot io.Bot) []Issue {
	var issues []Issue

	if len(bot.Channels) == 0 {
		issues = append(issues, Issue{
			Code: "doc.bot.missing_channels", Severity: Err,
			Path: "bot.channels",
			Msg:  "at least one channel is required",
		})
	}

	// Valida channels suportados
	supportedChannels := map[string]bool{
		"whatsapp":  true,
		"telegram":  true,
		"webchat":   true,
		"facebook":  true,
		"instagram": true,
	}

	for i, channel := range bot.Channels {
		if !supportedChannels[channel] {
			issues = append(issues, Issue{
				Code: "doc.bot.unsupported_channel", Severity: Warn,
				Path: fmt.Sprintf("bot.channels[%d]", i),
				Msg:  fmt.Sprintf("channel '%s' may not be fully supported", channel),
			})
		}
	}

	return issues
}

// validateProps valida propriedades globais
func (s *DocumentationComplianceStep) validateProps(props map[string]any) []Issue {
	var issues []Issue

	// Props são opcionais, mas se presentes devem seguir convenções
	for key, value := range props {
		// Valida naming convention
		if strings.Contains(key, " ") || strings.Contains(key, "-") {
			issues = append(issues, Issue{
				Code: "doc.props.invalid_key_format", Severity: Warn,
				Path: fmt.Sprintf("props.%s", key),
				Msg:  "property keys should use snake_case format",
			})
		}

		// Valida tipos suportados
		switch value.(type) {
		case string, int, float64, bool:
			// Tipos básicos são válidos
		default:
			issues = append(issues, Issue{
				Code: "doc.props.complex_type", Severity: Info,
				Path: fmt.Sprintf("props.%s", key),
				Msg:  "complex property types may not be accessible in all contexts",
			})
		}
	}

	return issues
}

// validateGraphWithEntries valida estrutura do grafo considerando entry points
func (s *DocumentationComplianceStep) validateGraphWithEntries(graph io.Graph, entries []flow.Entry) []Issue {
	var issues []Issue

	if len(graph.Nodes) == 0 {
		issues = append(issues, Issue{
			Code: "doc.graph.no_nodes", Severity: Err,
			Path: "graph.nodes",
			Msg:  "graph must contain at least one node",
		})
		return issues
	}

	// Valida nós
	nodeIDs := make(map[flow.ID]bool)
	for i, node := range graph.Nodes {
		path := fmt.Sprintf("graph.nodes[%d]", i)
		issues = append(issues, s.validateNode(node, path)...)
		nodeIDs[node.ID] = true
	}

	// Valida arestas
	for i, edge := range graph.Edges {
		path := fmt.Sprintf("graph.edges[%d]", i)
		issues = append(issues, s.validateEdge(edge, nodeIDs, path)...)
	}

	// Valida conectividade e terminais
	issues = append(issues, s.validateConnectivityWithEntries(graph, entries)...)

	return issues
}

// validateNode valida um nó individual
func (s *DocumentationComplianceStep) validateNode(node flow.Node, path string) []Issue {
	var issues []Issue

	// Valida ID
	if node.ID == "" {
		issues = append(issues, Issue{
			Code: "doc.node.missing_id", Severity: Err,
			Path: path + ".id",
			Msg:  "node ID is required",
		})
	}

	// Valida Kind
	if node.Kind == "" {
		issues = append(issues, Issue{
			Code: "doc.node.missing_kind", Severity: Err,
			Path: path + ".kind",
			Msg:  "node kind is required",
		})
	}

	// Valida kinds conhecidos
	knownKinds := map[string]bool{
		"message":    true,
		"text":       true,
		"buttons":    true,
		"listpicker": true,
		"carousel":   true,
		"confirm":    true,
		"media":      true,
		"delay":      true,
		"router":     true,
		"terminal":   true,
		"action":     true,
	}

	if node.Kind != "" && !knownKinds[node.Kind] {
		issues = append(issues, Issue{
			Code: "doc.node.unknown_kind", Severity: Warn,
			Path: path + ".kind",
			Msg:  fmt.Sprintf("unknown node kind: %s", node.Kind),
		})
	}

	// Valida props
	if len(node.Props) == 0 {
		issues = append(issues, Issue{
			Code: "doc.node.missing_props", Severity: Warn,
			Path: path + ".props",
			Msg:  "node should have properties defined",
		})
	}

	// Valida posicionamento para editor visual
	if node.X == nil || node.Y == nil {
		issues = append(issues, Issue{
			Code: "doc.node.missing_position", Severity: Info,
			Path: path + ".position",
			Msg:  "node position not set (will affect visual editor)",
		})
	}

	return issues
}

// validateEdge valida uma aresta individual
func (s *DocumentationComplianceStep) validateEdge(edge flow.Edge, nodeIDs map[flow.ID]bool, path string) []Issue {
	var issues []Issue

	// Valida From
	if edge.From == "" {
		issues = append(issues, Issue{
			Code: "doc.edge.missing_from", Severity: Err,
			Path: path + ".from",
			Msg:  "edge from is required",
		})
	} else if !nodeIDs[edge.From] {
		issues = append(issues, Issue{
			Code: "doc.edge.invalid_from", Severity: Err,
			Path: path + ".from",
			Msg:  fmt.Sprintf("edge references non-existent from node: %s", edge.From),
		})
	}

	// Valida To
	if edge.To == "" {
		issues = append(issues, Issue{
			Code: "doc.edge.missing_to", Severity: Err,
			Path: path + ".to",
			Msg:  "edge to is required",
		})
	} else if !nodeIDs[edge.To] {
		issues = append(issues, Issue{
			Code: "doc.edge.invalid_to", Severity: Err,
			Path: path + ".to",
			Msg:  fmt.Sprintf("edge references non-existent to node: %s", edge.To),
		})
	}

	// Valida label
	if edge.Label == "" {
		issues = append(issues, Issue{
			Code: "doc.edge.missing_label", Severity: Info,
			Path: path + ".label",
			Msg:  "edge should have a descriptive label",
		})
	}

	// Valida guard se presente
	if edge.Guard != nil && edge.Guard.Expr != "" {
		// Guard deve seguir sintaxe básica
		if !isValidGuardExpression(edge.Guard.Expr) {
			issues = append(issues, Issue{
				Code: "doc.edge.invalid_guard", Severity: Warn,
				Path: path + ".guard",
				Msg:  "guard expression may have invalid syntax",
			})
		}
	}

	return issues
}

// validateConnectivityWithEntries valida conectividade do grafo considerando entry points
func (s *DocumentationComplianceStep) validateConnectivityWithEntries(graph io.Graph, entries []flow.Entry) []Issue {
	var issues []Issue

	// Constrói mapa de adjacência
	outgoing := make(map[flow.ID]int)
	incoming := make(map[flow.ID]int)
	nodeExists := make(map[flow.ID]bool)

	for _, node := range graph.Nodes {
		nodeExists[node.ID] = true
		outgoing[node.ID] = 0
		incoming[node.ID] = 0
	}

	for _, edge := range graph.Edges {
		if nodeExists[edge.From] && nodeExists[edge.To] {
			outgoing[edge.From]++
			incoming[edge.To]++
		}
	}

	// Cria mapa de nós que são entry points
	entryNodes := make(map[flow.ID]bool)
	for _, entry := range entries {
		entryNodes[entry.Target] = true
	}

	// Valida nós sem conexões de entrada (exceto entry points e nós HSM)
	for nodeID := range nodeExists {
		if incoming[nodeID] == 0 && !entryNodes[nodeID] {
			// Verifica se é um nó HSM (que pode ser iniciado externamente)
			isHSMNode := false
			for _, node := range graph.Nodes {
				if node.ID == nodeID && node.Kind == "hsm" {
					isHSMNode = true
					break
				}
			}

			if !isHSMNode {
				issues = append(issues, Issue{
					Code: "doc.connectivity.unreachable_node", Severity: Warn,
					Path: fmt.Sprintf("graph.nodes[%s]", nodeID),
					Msg:  fmt.Sprintf("node %s has no incoming edges (may be unreachable)", nodeID),
				})
			}
		}

		if outgoing[nodeID] == 0 {
			// Nó terminal - isso pode ser intencional
			issues = append(issues, Issue{
				Code: "doc.connectivity.terminal_node", Severity: Info,
				Path: fmt.Sprintf("graph.nodes[%s]", nodeID),
				Msg:  fmt.Sprintf("node %s has no outgoing edges (terminal node)", nodeID),
			})
		}
	}

	return issues
}

// isValidGuardExpression valida sintaxe básica de expressões guard
func isValidGuardExpression(guard string) bool {
	// Validação básica - pode ser expandida
	if strings.TrimSpace(guard) == "" {
		return false
	}

	// Verifica estruturas comuns
	validPatterns := []string{
		"payload ==",
		"payload !=",
		"user.",
		"flow.",
		"props.",
		"&&",
		"||",
		"!",
		"==",
		"!=",
		">",
		"<",
		">=",
		"<=",
	}

	hasValidPattern := false
	for _, pattern := range validPatterns {
		if strings.Contains(guard, pattern) {
			hasValidPattern = true
			break
		}
	}

	return hasValidPattern
}
