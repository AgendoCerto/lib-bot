// Package reactflow contém tipos compatíveis com React Flow para renderização visual
package reactflow

// Position representa as coordenadas X,Y de um nó no editor visual
type Position struct {
	X float64 `json:"x"` // Coordenada horizontal (pixels)
	Y float64 `json:"y"` // Coordenada vertical (pixels)
}

// Dimensions representa as dimensões de um nó
type Dimensions struct {
	Width  float64 `json:"width"`  // Largura do nó (pixels)
	Height float64 `json:"height"` // Altura do nó (pixels)
}

// Node representa um nó no formato React Flow para o editor visual
type Node struct {
	ID         string         `json:"id"`                   // Identificador único do nó
	Type       string         `json:"type,omitempty"`       // Tipo do nó (mapeado do flow.Node.Kind)
	Data       map[string]any `json:"data,omitempty"`       // Dados do nó (props, refs, título, etc.)
	Position   Position       `json:"position"`             // Posição do nó no canvas (obrigatório no React Flow)
	Width      *float64       `json:"width,omitempty"`      // Largura do nó para auto-layout
	Height     *float64       `json:"height,omitempty"`     // Altura do nó para auto-layout
	Draggable  *bool          `json:"draggable,omitempty"`  // Se o nó pode ser arrastado
	Selectable *bool          `json:"selectable,omitempty"` // Se o nó pode ser selecionado
	Deletable  *bool          `json:"deletable,omitempty"`  // Se o nó pode ser deletado
	Hidden     *bool          `json:"hidden,omitempty"`     // Se o nó deve ser ocultado
	// Campos para handles de conexão
	SourcePosition string `json:"sourcePosition,omitempty"` // Posição do handle de saída
	TargetPosition string `json:"targetPosition,omitempty"` // Posição do handle de entrada
}

// Edge representa uma conexão/aresta no formato React Flow
type Edge struct {
	ID         string         `json:"id"`                   // Identificador único da aresta
	Source     string         `json:"source"`               // ID do nó de origem
	Target     string         `json:"target"`               // ID do nó de destino
	Type       string         `json:"type,omitempty"`       // Tipo da aresta (default: "default")
	Data       map[string]any `json:"data,omitempty"`       // Metadados da aresta (label, guard, priority)
	Animated   *bool          `json:"animated,omitempty"`   // Se a aresta deve ser animada
	Hidden     *bool          `json:"hidden,omitempty"`     // Se a aresta deve ser ocultada
	Deletable  *bool          `json:"deletable,omitempty"`  // Se a aresta pode ser deletada
	Selectable *bool          `json:"selectable,omitempty"` // Se a aresta pode ser selecionada
	Label      string         `json:"label,omitempty"`      // Label direta da aresta
	MarkerEnd  string         `json:"markerEnd,omitempty"`  // Marcador de fim (seta)
	// Campos opcionais: markerStart, style, className, etc.
}

// ReactFlowDocument representa o documento completo do ReactFlow
type ReactFlowDocument struct {
	Nodes []Node `json:"nodes"` // Lista de nós
	Edges []Edge `json:"edges"` // Lista de arestas
	// Configurações opcionais para o ReactFlow
	DefaultEdgeOptions *DefaultEdgeOptions `json:"defaultEdgeOptions,omitempty"`
	DefaultNodeOptions *DefaultNodeOptions `json:"defaultNodeOptions,omitempty"`
	// Layout information quando aplicado
	Layout *LayoutInfo `json:"layout,omitempty"`
}

// DefaultEdgeOptions opções padrão para edges
type DefaultEdgeOptions struct {
	Type       string `json:"type,omitempty"`      // Tipo padrão
	Animated   bool   `json:"animated,omitempty"`  // Animação padrão
	MarkerEnd  string `json:"markerEnd,omitempty"` // Marcador padrão
	Deletable  bool   `json:"deletable"`           // Deletável por padrão
	Selectable bool   `json:"selectable"`          // Selecionável por padrão
}

// DefaultNodeOptions opções padrão para nodes
type DefaultNodeOptions struct {
	Draggable      bool   `json:"draggable"`                // Arrastável por padrão
	Selectable     bool   `json:"selectable"`               // Selecionável por padrão
	SourcePosition string `json:"sourcePosition,omitempty"` // Posição padrão do handle de saída
	TargetPosition string `json:"targetPosition,omitempty"` // Posição padrão do handle de entrada
}

// LayoutInfo informações sobre layout aplicado
type LayoutInfo struct {
	Direction string `json:"direction"`           // vertical | horizontal
	Applied   bool   `json:"applied"`             // Se layout foi aplicado
	Algorithm string `json:"algorithm,omitempty"` // Algoritmo usado
}
