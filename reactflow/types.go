// Package reactflow contém tipos compatíveis com React Flow para renderização visual
package reactflow

// Position representa as coordenadas X,Y de um nó no editor visual
type Position struct {
	X float64 `json:"x"` // Coordenada horizontal (pixels)
	Y float64 `json:"y"` // Coordenada vertical (pixels)
}

// Node representa um nó no formato React Flow para o editor visual
type Node struct {
	ID       string         `json:"id"`             // Identificador único do nó
	Type     string         `json:"type,omitempty"` // Tipo do nó (mapeado do flow.Node.Kind)
	Data     map[string]any `json:"data,omitempty"` // Dados do nó (props, refs, título, etc.)
	Position Position       `json:"position"`       // Posição do nó no canvas (obrigatório no React Flow)
	// Campos opcionais disponíveis: className, draggable, selectable, etc.
}

// Edge representa uma conexão/aresta no formato React Flow
type Edge struct {
	ID     string         `json:"id"`             // Identificador único da aresta
	Source string         `json:"source"`         // ID do nó de origem
	Target string         `json:"target"`         // ID do nó de destino
	Data   map[string]any `json:"data,omitempty"` // Metadados da aresta (label, guard, priority)
	// Campos opcionais disponíveis: type, animated, markerEnd, style, etc.
}
