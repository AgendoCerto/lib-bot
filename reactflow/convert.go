// Package reactflow oferece conversores entre o formato interno e React Flow
package reactflow

import (
	"fmt"
	"strings"

	"github.com/AgendoCerto/lib-bot/component"
	"github.com/AgendoCerto/lib-bot/flow"
	"github.com/AgendoCerto/lib-bot/io"
	"github.com/AgendoCerto/lib-bot/layout"
)

// isFallbackEdge detecta se uma edge é de fallback/retry/timeout
func isFallbackEdge(e flow.Edge) bool {
	guardContains := func(text string) bool {
		if e.Guard != nil && e.Guard.Expr != "" {
			return strings.Contains(e.Guard.Expr, text)
		}
		return false
	}

	return e.Label == "timeout" || e.Label == "invalid" || e.Label == "fallback" ||
		guardContains("timeout") || guardContains("fallback") || guardContains("invalid")
}

// getEdgeHandles determina handles específicos para uma edge baseado na direção
func getEdgeHandles(e flow.Edge, direction layout.Direction) (sourceHandle, targetHandle string) {
	isFallback := isFallbackEdge(e)

	if direction == layout.DirectionHorizontal {
		// Layout horizontal: fluxo normal left->right, fallbacks top/bottom
		if isFallback {
			sourceHandle = "bottom"
			targetHandle = "top"
		} else {
			sourceHandle = "right"
			targetHandle = "left"
		}
	} else {
		// Layout vertical: fluxo normal top->bottom, fallbacks left/right
		if isFallback {
			sourceHandle = "right"
			targetHandle = "left"
		} else {
			sourceHandle = "bottom"
			targetHandle = "top"
		}
	}

	return sourceHandle, targetHandle
}

// getHandlePositions determina posições dos handles baseado na direção e se tem fallbacks
func getHandlePositions(edges []flow.Edge, nodeID flow.ID, direction layout.Direction) (sourcePos, targetPos string) {
	// Verifica se este nó tem edges de fallback
	hasOutgoingFallback := false
	hasIncomingFallback := false
	for _, e := range edges {
		if string(e.From) == string(nodeID) && isFallbackEdge(e) {
			hasOutgoingFallback = true
		}
		if string(e.To) == string(nodeID) && isFallbackEdge(e) {
			hasIncomingFallback = true
		}
	}

	if direction == layout.DirectionHorizontal {
		// Layout horizontal: fluxo normal left->right, fallbacks top/bottom
		sourcePos, targetPos = "right", "left"
		if hasOutgoingFallback {
			sourcePos = "bottom"
		}
		if hasIncomingFallback {
			targetPos = "top"
		}
	} else {
		// Layout vertical: fluxo normal top->bottom, fallbacks left/right
		sourcePos, targetPos = "bottom", "top"
		if hasOutgoingFallback {
			sourcePos = "right"
		}
		if hasIncomingFallback {
			targetPos = "left"
		}
	}

	return sourcePos, targetPos
}

// DesignToReactFlow converte um design interno para formato React Flow
// - Node.Type  = flow.Node.Kind
// - Node.Data  = { "props_ref" | "props", "final": bool, "kind": string, "title": string }
// - Edge.Data  = { "label", "priority", "guard" }
// - Position   = usa coordenadas persistidas (X,Y) se disponíveis, senão (0,0) por padrão
func DesignToReactFlow(d io.DesignDoc) (nodes []Node, edges []Edge) {
	return DesignToReactFlowWithDirection(d, layout.DirectionVertical) // Padrão vertical
}

// DesignToReactFlowWithDirection converte design para React Flow com direção específica
func DesignToReactFlowWithDirection(d io.DesignDoc, direction layout.Direction) (nodes []Node, edges []Edge) {
	// Primeiro passo: análise de edges para detectar fallbacks e loops
	fallbackMap := make(map[string]bool)

	for _, e := range d.Graph.Edges {
		if isFallbackEdge(e) {
			edgeKey := string(e.From) + "->" + string(e.To)
			fallbackMap[edgeKey] = true
		}
	}

	// Adiciona nós de entrada especiais baseado nos entries
	startNodeCount := 0
	for _, entry := range d.Entries {
		if entry.Kind == flow.EntryGlobalStart {
			startNodeCount++

			// Cria nó de entrada visual
			startNodeID := "__start_global"

			// Posição inicial do nó de entrada
			startPos := Position{X: 0, Y: 0}
			if direction == layout.DirectionHorizontal {
				startPos.X = 50 // Mais à esquerda em layout horizontal
				startPos.Y = 100
			} else {
				startPos.X = 100
				startPos.Y = 50 // Mais acima em layout vertical
			}

			startNode := Node{
				ID:   startNodeID,
				Type: "start", // Tipo especial para nó de entrada
				Data: map[string]any{
					"kind":       "start",
					"final":      false,
					"title":      "Início",
					"entry_kind": string(entry.Kind),
					"target":     string(entry.Target),
				},
				Position:       startPos,
				Draggable:      boolPtr(true),
				Selectable:     boolPtr(true),
				Deletable:      boolPtr(false), // Nó de início não pode ser deletado
				SourcePosition: "right",        // Sempre aponta para a direita inicialmente
				TargetPosition: "left",         // Não recebe conexões
			}

			// Ajusta handles baseado na direção
			if direction == layout.DirectionHorizontal {
				startNode.SourcePosition = "right"
			} else {
				startNode.SourcePosition = "bottom"
			}

			// Dimensões fixas para nó de entrada

			nodes = append(nodes, startNode)

			// Cria edge do nó de entrada para o nó alvo
			startEdge := Edge{
				ID:         "__edge_start_" + string(entry.Target),
				Source:     startNodeID,
				Target:     string(entry.Target),
				Type:       "default",
				Label:      "início",
				MarkerEnd:  "arrowclosed",
				Animated:   boolPtr(false),
				Deletable:  boolPtr(false), // Edge de início não pode ser deletada
				Selectable: boolPtr(true),
			}

			// Handles da edge de entrada
			if direction == layout.DirectionHorizontal {
				sourceHandle := "right"
				targetHandle := "left"
				startEdge.SourceHandle = &sourceHandle
				startEdge.TargetHandle = &targetHandle
			} else {
				sourceHandle := "bottom"
				targetHandle := "top"
				startEdge.SourceHandle = &sourceHandle
				startEdge.TargetHandle = &targetHandle
			}

			edges = append(edges, startEdge)
		}
	}

	// Continua com os nós regulares

	for _, n := range d.Graph.Nodes {
		data := map[string]any{
			"kind":  n.Kind,
			"final": n.Final,
		}

		// Adiciona título se disponível
		if n.Title != "" {
			data["title"] = n.Title
		}

		if n.PropsRef != "" {
			data["props_ref"] = n.PropsRef
		} else if len(n.Props) > 0 {
			data["props"] = n.Props
		}

		// Adiciona informações de persistência se disponíveis
		var availableKeys []string
		// WhatsApp defaults
		availableKeys = append(availableKeys, "context.wa_phone", "context.wa_name")

		if n.PropsRef != "" {
			if refProps, exists := d.Props[n.PropsRef]; exists {
				if refPropsMap, ok := refProps.(map[string]any); ok {
					persistenceConfig, err := component.ParsePersistence(refPropsMap)
					if err == nil && persistenceConfig != nil {
						persistenceData := map[string]any{
							"enabled": persistenceConfig.Enabled,
							"scope":   string(persistenceConfig.Scope),
							"key":     persistenceConfig.Key,
						}
						if persistenceConfig.Sanitization != nil {
							persistenceData["sanitization"] = map[string]any{
								"type":         string(persistenceConfig.Sanitization.Type),
								"custom_regex": persistenceConfig.Sanitization.CustomRegex,
								"replacement":  persistenceConfig.Sanitization.Replacement,
								"description":  persistenceConfig.Sanitization.Description,
								"strict_mode":  persistenceConfig.Sanitization.StrictMode,
							}
						}
						if persistenceConfig.Required {
							persistenceData["required"] = persistenceConfig.Required
						}
						if persistenceConfig.DefaultValue != "" {
							persistenceData["default_value"] = persistenceConfig.DefaultValue
						}
						data["persistence"] = persistenceData
						// Add key to availableKeys
						if persistenceConfig.Key != "" {
							if persistenceConfig.Scope == "context" {
								availableKeys = append(availableKeys, "context."+persistenceConfig.Key)
							} else if persistenceConfig.Scope == "profile" {
								availableKeys = append(availableKeys, "profile."+persistenceConfig.Key)
							}
						}
					}
				}
			}
		} else if len(n.Props) > 0 {
			persistenceConfig, err := component.ParsePersistence(n.Props)
			if err == nil && persistenceConfig != nil {
				persistenceData := map[string]any{
					"enabled": persistenceConfig.Enabled,
					"scope":   string(persistenceConfig.Scope),
					"key":     persistenceConfig.Key,
				}
				if persistenceConfig.Sanitization != nil {
					persistenceData["sanitization"] = map[string]any{
						"type":         string(persistenceConfig.Sanitization.Type),
						"custom_regex": persistenceConfig.Sanitization.CustomRegex,
						"replacement":  persistenceConfig.Sanitization.Replacement,
						"description":  persistenceConfig.Sanitization.Description,
						"strict_mode":  persistenceConfig.Sanitization.StrictMode,
					}
				}
				if persistenceConfig.Required {
					persistenceData["required"] = persistenceConfig.Required
				}
				if persistenceConfig.DefaultValue != "" {
					persistenceData["default_value"] = persistenceConfig.DefaultValue
				}
				data["persistence"] = persistenceData
				// Add key to availableKeys
				if persistenceConfig.Key != "" {
					if persistenceConfig.Scope == "context" {
						availableKeys = append(availableKeys, "context."+persistenceConfig.Key)
					} else if persistenceConfig.Scope == "profile" {
						availableKeys = append(availableKeys, "profile."+persistenceConfig.Key)
					}
				}
			}
		}
		// Expose availableKeys for frontend
		data["available_keys"] = availableKeys

		// Usa coordenadas persistidas se disponíveis, senão padrão (0,0)
		pos := Position{X: 0, Y: 0}
		if n.X != nil && n.Y != nil {
			pos.X = *n.X
			pos.Y = *n.Y
		}

		// Determina posições dos handles baseado na direção e contexto das conexões
		sourcePos, targetPos := getHandlePositions(d.Graph.Edges, n.ID, direction)

		node := Node{
			ID:             string(n.ID),
			Type:           n.Kind, // React Flow "type" = seu "kind" (message, confirm, etc.)
			Data:           data,
			Position:       pos,
			Draggable:      boolPtr(true),     // Permite arrastar nós por padrão
			Selectable:     boolPtr(true),     // Permite selecionar nós por padrão
			Deletable:      boolPtr(!n.Final), // Nós finais não podem ser deletados por segurança
			SourcePosition: sourcePos,         // Handle de saída dinâmico
			TargetPosition: targetPos,         // Handle de entrada dinâmico
		}

		nodes = append(nodes, node)
	}

	// Adiciona edges regulares às edges existentes (que já podem incluir edges de início)
	regularEdges := make([]Edge, 0, len(d.Graph.Edges))
	for i, e := range d.Graph.Edges {
		data := map[string]any{}
		if e.Label != "" {
			data["label"] = e.Label
		}
		if e.Priority != 0 {
			data["priority"] = e.Priority
		}
		if e.Guard != nil && e.Guard.Expr != "" {
			data["guard"] = e.Guard.Expr
		}
		// Prepara o label direto da edge se disponível
		var directLabel string
		if e.Label != "" {
			directLabel = e.Label
		}

		// Determina estilo da edge baseado no tipo
		edgeType := "default"
		animated := false
		markerEnd := "arrowclosed"

		// Personaliza edge baseado em guards/labels especiais
		if e.Guard != nil && e.Guard.Expr != "" {
			// Edges com guards são condicionais
			if strings.Contains(e.Guard.Expr, "timeout") {
				edgeType = "step"
				animated = true
				markerEnd = "arrow"
			} else if strings.Contains(e.Guard.Expr, "fallback") || strings.Contains(e.Guard.Expr, "invalid") {
				edgeType = "step"
				markerEnd = "arrow"
			}
		}

		// Edges de timeout/error podem ser animadas para destaque
		if e.Label == "timeout" || e.Label == "invalid" || e.Label == "fallback" {
			animated = true
		}

		// Determina handles específicos para esta edge baseado na direção
		sourceHandle, targetHandle := getEdgeHandles(e, direction)

		id := fmt.Sprintf("e%d_%s_%s", i, e.From, e.To)
		edge := Edge{
			ID:           id,
			Source:       string(e.From),
			Target:       string(e.To),
			Type:         edgeType, // Tipo baseado no contexto
			Data:         data,
			Label:        directLabel,   // Label direta na edge
			MarkerEnd:    markerEnd,     // Marcador baseado no tipo
			Animated:     &animated,     // Animação para destacar fluxos especiais
			Deletable:    boolPtr(true), // Permite deletar edges por padrão
			Selectable:   boolPtr(true), // Permite selecionar edges por padrão
			SourceHandle: &sourceHandle, // Handle específico da fonte
			TargetHandle: &targetHandle, // Handle específico do destino
		}

		regularEdges = append(regularEdges, edge)
	}

	// Combina edges de início com edges regulares
	edges = append(edges, regularEdges...)
	return nodes, edges
}

// ReactFlowToDesign converte nós/arestas do React Flow de volta para DesignDoc interno
// - Lê "data.props" (inline) OU "data.props_ref" (preferência por props_ref se existir)
// - Preserva título, coordenadas e propriedades dos nós
// - Reconstrói arestas com labels, prioridades e guards
// - "base" permite reaproveitar Bot/Version/Entries/Props existentes (para atualizações incrementais)
func ReactFlowToDesign(nodes []Node, edges []Edge, base io.DesignDoc) io.DesignDoc {
	out := base
	// Define valores padrão se não existirem no design base
	if out.Schema == "" {
		out.Schema = "flowkit/1.0"
	}
	if out.Bot.ID == "" {
		out.Bot = io.Bot{ID: "bot_unknown", Channels: []string{}}
	}
	if out.Version.ID == "" {
		out.Version = io.Version{ID: "01NEW", Status: "development"}
	}
	if out.Props == nil {
		out.Props = map[string]any{}
	}

	// Nodes - converte nós do React Flow de volta para estrutura interna
	out.Graph.Nodes = make([]flow.Node, 0, len(nodes))
	for _, n := range nodes {
		kind := n.Type
		if k, ok := n.Data["kind"].(string); ok && k != "" {
			kind = k
		}
		fn := flow.Node{
			ID:   flow.ID(n.ID),
			Kind: kind,
		}

		// Preserva título se disponível
		if title, ok := n.Data["title"].(string); ok && title != "" {
			fn.Title = title
		}

		// props_ref tem prioridade caso exista
		if pr, ok := n.Data["props_ref"].(string); ok && pr != "" {
			fn.PropsRef = pr

			// Se há informações de persistência no data para um props_ref,
			// atualizamos as props referenciais no design base
			if persistenceData, hasPersistence := n.Data["persistence"].(map[string]any); hasPersistence {
				if out.Props == nil {
					out.Props = make(map[string]any)
				}
				if refProps, exists := out.Props[pr]; exists {
					if refPropsMap, ok := refProps.(map[string]any); ok {
						refPropsMap["persistence"] = persistenceData
						out.Props[pr] = refPropsMap
					}
				}
			}
		} else if p, ok := n.Data["props"].(map[string]any); ok && p != nil {
			fn.Props = p

			// Se há informações de persistência no data, inclui no props
			if persistenceData, hasPersistence := n.Data["persistence"].(map[string]any); hasPersistence {
				if fn.Props == nil {
					fn.Props = make(map[string]any)
				}
				fn.Props["persistence"] = persistenceData
			}
		} else {
			// Se não há props mas há persistência, cria props apenas com persistência
			if persistenceData, hasPersistence := n.Data["persistence"].(map[string]any); hasPersistence {
				fn.Props = map[string]any{
					"persistence": persistenceData,
				}
			}
		}
		if f, ok := n.Data["final"].(bool); ok {
			fn.Final = f
		}

		// Preserva as coordenadas de posição do React Flow
		fn.X = &n.Position.X
		fn.Y = &n.Position.Y

		out.Graph.Nodes = append(out.Graph.Nodes, fn)
	}

	// Edges
	out.Graph.Edges = make([]flow.Edge, 0, len(edges))
	for _, e := range edges {
		label, _ := e.Data["label"].(string)
		guard, _ := e.Data["guard"].(string)
		priority := 0
		if p, ok := e.Data["priority"].(float64); ok {
			priority = int(p)
		} else if p2, ok := e.Data["priority"].(int); ok {
			priority = p2
		}
		var guardPtr *flow.Guard
		if guard != "" {
			guardPtr = &flow.Guard{Expr: guard}
		}

		out.Graph.Edges = append(out.Graph.Edges, flow.Edge{
			From:     flow.ID(e.Source),
			To:       flow.ID(e.Target),
			Label:    label,
			Priority: priority,
			Guard:    guardPtr,
		})
	}

	return out
}

// DesignToReactFlowWithAutoLayout converte design para React Flow aplicando auto-layout
func DesignToReactFlowWithAutoLayout(d io.DesignDoc, direction layout.Direction) (nodes []Node, edges []Edge) {
	// Primeiro converte com direção específica para handles corretos
	nodes, edges = DesignToReactFlowWithDirection(d, direction)

	// Sempre aplica auto-layout quando solicitado explicitamente
	// (ignorando posições existentes)
	nodes = ApplyAutoLayout(d.Graph.Nodes, d.Graph.Edges, nodes, direction)

	return nodes, edges
}

// ApplyAutoLayout aplica algoritmo de auto-layout aos nós React Flow
func ApplyAutoLayout(flowNodes []flow.Node, flowEdges []flow.Edge, reactNodes []Node, direction layout.Direction) []Node {
	config := layout.DefaultConfig()
	config.Direction = direction

	// Cria grafo temporário para aplicar layout
	tempGraph := &flow.Graph{
		Nodes: flowNodes,
		Edges: flowEdges,
	}

	// Aplica layout ao grafo temporário
	err := layout.ApplyAutoLayout(tempGraph, config)
	if err != nil {
		// Em caso de erro, retorna nós sem modificar posições
		return reactNodes
	}

	// Mapeia resultados de volta para os nós React Flow
	nodeMap := make(map[string]*Node)
	for i := range reactNodes {
		nodeMap[reactNodes[i].ID] = &reactNodes[i]
	}

	// Aplica posições calculadas
	for _, flowNode := range tempGraph.Nodes {
		if node, exists := nodeMap[string(flowNode.ID)]; exists && flowNode.X != nil && flowNode.Y != nil {
			node.Position.X = *flowNode.X
			node.Position.Y = *flowNode.Y
		}
	}

	return reactNodes
}

// ApplyAutoLayoutVertical aplica layout vertical aos nós
func ApplyAutoLayoutVertical(d io.DesignDoc) (nodes []Node, edges []Edge) {
	return DesignToReactFlowWithAutoLayout(d, layout.DirectionVertical)
}

// ApplyAutoLayoutHorizontal aplica layout horizontal aos nós
func ApplyAutoLayoutHorizontal(d io.DesignDoc) (nodes []Node, edges []Edge) {
	return DesignToReactFlowWithAutoLayout(d, layout.DirectionHorizontal)
}

// CreateReactFlowDocument cria um documento ReactFlow completo com configurações otimizadas
func CreateReactFlowDocument(nodes []Node, edges []Edge, layoutInfo *LayoutInfo) ReactFlowDocument {
	return ReactFlowDocument{
		Nodes: nodes,
		Edges: edges,
		DefaultEdgeOptions: &DefaultEdgeOptions{
			Type:       "default",
			Animated:   false,
			MarkerEnd:  "arrowclosed",
			Deletable:  true,
			Selectable: true,
		},
		DefaultNodeOptions: &DefaultNodeOptions{
			Draggable:      true,
			Selectable:     true,
			SourcePosition: "right",
			TargetPosition: "left",
		},
		Layout: layoutInfo,
	}
}

// boolPtr retorna um ponteiro para um valor booleano
func boolPtr(b bool) *bool {
	return &b
}
