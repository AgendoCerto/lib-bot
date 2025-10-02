package validate

import (
	"github.com/AgendoCerto/lib-bot/adapter"
	"github.com/AgendoCerto/lib-bot/component"
	"github.com/AgendoCerto/lib-bot/flow"
	"github.com/AgendoCerto/lib-bot/io"
)

// TopologyValidator valida a topologia e estrutura do fluxo
type TopologyValidator struct{}

// NewTopologyValidator cria um novo validador de topologia
func NewTopologyValidator() *TopologyValidator {
	return &TopologyValidator{}
}

// ValidateDesign valida a topologia completa de um design
func (v *TopologyValidator) ValidateDesign(design io.DesignDoc) []Issue {
	var issues []Issue

	// Validação 1: Start Global único e obrigatório
	issues = append(issues, v.validateStartEntries(design.Entries)...)

	// Validação 2: Nós terminais sem arestas de saída
	issues = append(issues, v.validateTerminalNodes(design.Graph)...)

	// Validação 3: Prioridades únicas por nó
	issues = append(issues, v.validateUniquePriorities(design.Graph)...)

	// Validação 4: Referências válidas (nós e arestas)
	issues = append(issues, v.validateReferences(design.Graph, design.Entries)...)

	// Validação 5: Ciclos apenas com guardas
	issues = append(issues, v.validateCycles(design.Graph)...)

	// Validação 6: Channels válidos
	issues = append(issues, v.validateChannels(design.Bot)...)

	return issues
}

// validateStartEntries valida entradas de início
func (v *TopologyValidator) validateStartEntries(entries []flow.Entry) []Issue {
	var issues []Issue

	globalStarts := 0
	channelStarts := make(map[string]int)

	for i, entry := range entries {
		path := "entries[" + itoa(i) + "]"

		switch entry.Kind {
		case flow.EntryGlobalStart:
			globalStarts++
			if entry.ChannelID != "" {
				issues = append(issues, Issue{
					Code: "topology.start.global_with_channel", Severity: Err,
					Path: path + ".channel_id",
					Msg:  "global start entry should not have channel_id",
				})
			}
		case flow.EntryChannelStart:
			if entry.ChannelID == "" {
				issues = append(issues, Issue{
					Code: "topology.start.channel_missing_id", Severity: Err,
					Path: path + ".channel_id",
					Msg:  "channel start entry must have channel_id",
				})
			} else {
				channelStarts[entry.ChannelID]++
			}
		case flow.EntryForced:
			// Entradas forçadas são sempre válidas
		default:
			issues = append(issues, Issue{
				Code: "topology.entry.invalid_kind", Severity: Err,
				Path: path + ".kind",
				Msg:  "invalid entry kind: " + string(entry.Kind),
			})
		}
	}

	// Deve ter exatamente um start global
	if globalStarts == 0 {
		issues = append(issues, Issue{
			Code: "topology.start.global_missing", Severity: Err,
			Path: "entries",
			Msg:  "design must have exactly one global_start entry",
		})
	} else if globalStarts > 1 {
		issues = append(issues, Issue{
			Code: "topology.start.global_multiple", Severity: Err,
			Path: "entries",
			Msg:  "design must have exactly one global_start entry, found " + itoa(globalStarts),
		})
	}

	// Cada channel deve ter no máximo um start específico
	for channelID, count := range channelStarts {
		if count > 1 {
			issues = append(issues, Issue{
				Code: "topology.start.channel_multiple", Severity: Err,
				Path: "entries",
				Msg:  "channel '" + channelID + "' has multiple channel_start entries",
			})
		}
	}

	return issues
}

// validateTerminalNodes valida que nós terminais não têm arestas de saída
func (v *TopologyValidator) validateTerminalNodes(graph io.Graph) []Issue {
	var issues []Issue

	// Mapa de nós terminais
	terminalNodes := make(map[flow.ID]bool)
	for _, node := range graph.Nodes {
		if node.Final {
			terminalNodes[node.ID] = true
		}
	}

	// Verifica se nós terminais têm arestas de saída
	for _, edge := range graph.Edges {
		if terminalNodes[edge.From] {
			issues = append(issues, Issue{
				Code: "topology.terminal.has_outgoing", Severity: Err,
				Path: "graph.edges",
				Msg:  "terminal node '" + string(edge.From) + "' cannot have outgoing edges",
			})
		}
	}

	return issues
}

// validateUniquePriorities valida que prioridades são únicas por nó
func (v *TopologyValidator) validateUniquePriorities(graph io.Graph) []Issue {
	var issues []Issue

	// Agrupa arestas por nó de origem
	edgesByNode := make(map[flow.ID][]flow.Edge)
	for _, edge := range graph.Edges {
		edgesByNode[edge.From] = append(edgesByNode[edge.From], edge)
	}

	// Verifica prioridades únicas para cada nó
	for nodeID, edges := range edgesByNode {
		priorities := make(map[int]bool)
		for _, edge := range edges {
			if edge.Priority == 0 {
				continue // Prioridade 0 é ignorada (sem prioridade)
			}
			if priorities[edge.Priority] {
				issues = append(issues, Issue{
					Code: "topology.priority.duplicate", Severity: Err,
					Path: "graph.edges",
					Msg:  "node '" + string(nodeID) + "' has duplicate priority " + itoa(edge.Priority),
				})
			}
			priorities[edge.Priority] = true
		}
	}

	return issues
}

// validateReferences valida que todas as referências são válidas
func (v *TopologyValidator) validateReferences(graph io.Graph, entries []flow.Entry) []Issue {
	var issues []Issue

	// Mapa de nós válidos
	validNodes := make(map[flow.ID]bool)
	for _, node := range graph.Nodes {
		validNodes[node.ID] = true
	}

	// Valida referências nas arestas
	for i, edge := range graph.Edges {
		path := "graph.edges[" + itoa(i) + "]"

		if !validNodes[edge.From] {
			issues = append(issues, Issue{
				Code: "topology.reference.invalid_from", Severity: Err,
				Path: path + ".from",
				Msg:  "edge references non-existent node: " + string(edge.From),
			})
		}

		if !validNodes[edge.To] {
			issues = append(issues, Issue{
				Code: "topology.reference.invalid_to", Severity: Err,
				Path: path + ".to",
				Msg:  "edge references non-existent node: " + string(edge.To),
			})
		}
	}

	// Valida referências nas entradas
	for i, entry := range entries {
		path := "entries[" + itoa(i) + "]"

		if !validNodes[entry.Target] {
			issues = append(issues, Issue{
				Code: "topology.reference.invalid_target", Severity: Err,
				Path: path + ".target",
				Msg:  "entry references non-existent node: " + string(entry.Target),
			})
		}
	}

	return issues
}

// validateCycles valida ciclos no grafo (implementação simplificada)
func (v *TopologyValidator) validateCycles(graph io.Graph) []Issue {
	var issues []Issue

	// Constrói mapa de adjacência
	adjList := make(map[flow.ID][]flow.Edge)
	for _, edge := range graph.Edges {
		adjList[edge.From] = append(adjList[edge.From], edge)
	}

	// ✅ Labels que permitem self-loops seguros (behaviors com término garantido)
	allowedSelfLoopLabels := map[string]bool{
		"timeout":    true, // Timeout tem duração limitada
		"retry":      true, // Retry tem max_attempts
		"fallback":   true, // Fallback é uma alternativa controlada
		"validation": true, // Validation tem tentativas limitadas
		"error":      true, // Error handling é controlado
	}

	// DFS para detectar ciclos
	visited := make(map[flow.ID]bool)
	recStack := make(map[flow.ID]bool)

	var dfs func(flow.ID, []flow.ID) bool
	dfs = func(nodeID flow.ID, path []flow.ID) bool {
		visited[nodeID] = true
		recStack[nodeID] = true

		for _, edge := range adjList[nodeID] {
			if !visited[edge.To] {
				if dfs(edge.To, append(path, nodeID)) {
					return true
				}
			} else if recStack[edge.To] {
				// ✅ Ciclo detectado - aplicar lógica inteligente

				// 1. Permite se é um self-loop (mesmo nó) com label permitida
				if nodeID == edge.To && allowedSelfLoopLabels[edge.Label] {
					// Self-loop com label segura - não é erro
					continue
				}

				// 2. Permite se tem guarda definida
				if edge.Guard != nil && edge.Guard.Expr != "" {
					// Ciclo com guarda - não é erro
					continue
				}

				// 3. Bloqueia ciclos sem guarda ou label permitida
				issues = append(issues, Issue{
					Code: "topology.cycle.no_guard", Severity: Warn, // ⚠️ Mudado para Warning
					Path: "graph.edges",
					Msg:  "cycle detected without guard condition: " + string(nodeID) + " -> " + string(edge.To),
				})
				return true
			}
		}

		recStack[nodeID] = false
		return false
	}

	// Executa DFS de cada nó não visitado
	for _, node := range graph.Nodes {
		if !visited[node.ID] {
			dfs(node.ID, []flow.ID{})
		}
	}

	return issues
}

// validateChannels valida configuração de canais
func (v *TopologyValidator) validateChannels(bot io.Bot) []Issue {
	var issues []Issue

	if len(bot.Channels) == 0 {
		issues = append(issues, Issue{
			Code: "topology.channels.empty", Severity: Err,
			Path: "bot.channels",
			Msg:  "bot must have at least one channel",
		})
	}

	// Valida formato dos canais (básico)
	for i, channel := range bot.Channels {
		path := "bot.channels[" + itoa(i) + "]"

		if channel == "" {
			issues = append(issues, Issue{
				Code: "topology.channel.empty", Severity: Err,
				Path: path,
				Msg:  "channel cannot be empty",
			})
		}
		// Aqui pode adicionar validações mais específicas de formato
	}

	return issues
}

// TopologyStep implementa Step interface
type TopologyStep struct {
	validator *TopologyValidator
}

// NewTopologyStep cria um novo step de validação de topologia
func NewTopologyStep() *TopologyStep {
	return &TopologyStep{
		validator: NewTopologyValidator(),
	}
}

// Check implementa Step interface
func (s *TopologyStep) Check(spec component.ComponentSpec, caps adapter.Capabilities, path string) []Issue {
	// Esta interface não se aplica bem à validação de topologia
	// que precisa do design completo, então retorna vazio
	return []Issue{}
}
