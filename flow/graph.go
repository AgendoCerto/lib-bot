// Package flow define as estruturas básicas do fluxo de conversação
package flow

// ID é um identificador único para nós e referências
type ID string

// Node representa um nó no fluxo de conversação
type Node struct {
	ID       ID             `json:"id"`                  // Identificador único do nó
	Kind     string         `json:"kind"`                // Tipo do nó (message, confirm, etc.)
	Title    string         `json:"title,omitempty"`     // Título opcional para exibição no editor
	Props    map[string]any `json:"props"`               // Propriedades específicas do componente
	PropsRef string         `json:"props_ref,omitempty"` // Referência para propriedades compartilhadas
	Final    bool           `json:"final,omitempty"`     // Indica se é um nó terminal
	// Coordenadas para persistência de posição no React Flow
	X *float64 `json:"x,omitempty"` // Coordenada X da posição no editor visual
	Y *float64 `json:"y,omitempty"` // Coordenada Y da posição no editor visual
}

// Edge representa uma conexão entre dois nós
type Edge struct {
	From     ID     `json:"from"`            // ID do nó de origem
	To       ID     `json:"to"`              // ID do nó de destino
	Label    string `json:"label"`           // Rótulo da transição
	Priority int    `json:"priority"`        // Prioridade para ordenação de transições
	Guard    string `json:"guard,omitempty"` // Condição para ativação da transição
}

// Graph representa o grafo completo do fluxo
type Graph struct {
	Nodes []Node `json:"nodes"` // Lista de todos os nós
	Edges []Edge `json:"edges"` // Lista de todas as arestas/conexões
}
