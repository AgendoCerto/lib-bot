// Package reactflow oferece conversores entre o formato interno e React Flow
package reactflow

import (
	"fmt"

	"lib-bot/flow"
	"lib-bot/io"
)

// DesignToReactFlow converte um design interno para formato React Flow
// - Node.Type  = flow.Node.Kind
// - Node.Data  = { "props_ref" | "props", "final": bool, "kind": string, "title": string }
// - Edge.Data  = { "label", "priority", "guard" }
// - Position   = usa coordenadas persistidas (X,Y) se disponíveis, senão (0,0) por padrão
func DesignToReactFlow(d io.DesignDoc) (nodes []Node, edges []Edge) {
	nodes = make([]Node, 0, len(d.Graph.Nodes))
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

		// Usa coordenadas persistidas se disponíveis, senão padrão (0,0)
		pos := Position{X: 0, Y: 0}
		if n.X != nil && n.Y != nil {
			pos.X = *n.X
			pos.Y = *n.Y
		}

		nodes = append(nodes, Node{
			ID:       string(n.ID),
			Type:     n.Kind, // React Flow "type" = seu "kind" (message, confirm, etc.)
			Data:     data,
			Position: pos,
		})
	}

	edges = make([]Edge, 0, len(d.Graph.Edges))
	for i, e := range d.Graph.Edges {
		data := map[string]any{}
		if e.Label != "" {
			data["label"] = e.Label
		}
		if e.Priority != 0 {
			data["priority"] = e.Priority
		}
		if e.Guard != "" {
			data["guard"] = e.Guard
		}
		id := fmt.Sprintf("e%d_%s_%s", i, e.From, e.To)
		edges = append(edges, Edge{
			ID:     id,
			Source: string(e.From),
			Target: string(e.To),
			Data:   data,
		})
	}
	return
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
		} else if p, ok := n.Data["props"].(map[string]any); ok && p != nil {
			fn.Props = p
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
		out.Graph.Edges = append(out.Graph.Edges, flow.Edge{
			From:     flow.ID(e.Source),
			To:       flow.ID(e.Target),
			Label:    label,
			Priority: priority,
			Guard:    guard,
		})
	}

	return out
}
