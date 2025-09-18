// Package layout fornece algoritmos de auto-layout para organização automática de nós
package layout

import (
	"math"
	"sort"

	"lib-bot/flow"
)

// Direction define a direção do layout
type Direction string

const (
	DirectionVertical   Direction = "vertical"   // Layout de cima para baixo
	DirectionHorizontal Direction = "horizontal" // Layout da esquerda para direita
)

// Config configurações para o algoritmo de layout
type Config struct {
	Direction    Direction // Direção principal do layout
	NodeSpacing  float64   // Espaçamento entre nós (pixels)
	LevelSpacing float64   // Espaçamento entre níveis/camadas (pixels)
	StartX       float64   // Posição X inicial
	StartY       float64   // Posição Y inicial
}

// DefaultConfig retorna configuração padrão
func DefaultConfig() Config {
	return Config{
		Direction:    DirectionVertical,
		NodeSpacing:  50,
		LevelSpacing: 100,
		StartX:       100,
		StartY:       100,
	}
}

// NodeDimensions dimensões padrão por tipo de componente
var DefaultNodeDimensions = map[string]struct{ Width, Height float64 }{
	"message":    {Width: 250, Height: 80},
	"confirm":    {Width: 280, Height: 100},
	"buttons":    {Width: 300, Height: 120},
	"menu":       {Width: 320, Height: 140},
	"listpicker": {Width: 350, Height: 160},
	"carousel":   {Width: 400, Height: 200},
	"validate":   {Width: 280, Height: 90},
	"form":       {Width: 350, Height: 180},
	"delay":      {Width: 200, Height: 60},
	"router":     {Width: 220, Height: 80},
	"loop":       {Width: 260, Height: 100},
	"subflow":    {Width: 300, Height: 90},
	"abtest":     {Width: 280, Height: 110},
	"terminal":   {Width: 240, Height: 70},
	"action":     {Width: 250, Height: 80},
	"webhook":    {Width: 280, Height: 90},
	"default":    {Width: 200, Height: 80}, // fallback
}

// GetNodeDimensions retorna as dimensões para um tipo de nó
func GetNodeDimensions(kind string) (width, height float64) {
	if dims, exists := DefaultNodeDimensions[kind]; exists {
		return dims.Width, dims.Height
	}
	// Fallback para dimensões padrão
	return DefaultNodeDimensions["default"].Width, DefaultNodeDimensions["default"].Height
}

// LayoutResult resultado do algoritmo de layout
type LayoutResult struct {
	Nodes  []LayoutNode // Nós com posições calculadas
	Width  float64      // Largura total do layout
	Height float64      // Altura total do layout
}

// LayoutNode nó com posição e dimensões calculadas
type LayoutNode struct {
	ID     flow.ID // ID do nó
	X      float64 // Posição X calculada
	Y      float64 // Posição Y calculada
	Width  float64 // Largura do nó
	Height float64 // Altura do nó
	Level  int     // Nível/camada no layout
}

// AutoLayout aplica algoritmo de auto-layout aos nós
func AutoLayout(nodes []flow.Node, edges []flow.Edge, config Config) LayoutResult {
	if len(nodes) == 0 {
		return LayoutResult{}
	}

	// 1. Constrói grafo de adjacência e calcula níveis
	graph := buildGraph(nodes, edges)
	levels := calculateLevels(graph, findStartNodes(graph))

	// 2. Organiza nós por nível
	nodesByLevel := organizeByLevels(nodes, levels)

	// 3. Calcula posições
	layoutNodes := calculatePositions(nodesByLevel, config)

	// 4. Calcula dimensões totais
	width, height := calculateTotalDimensions(layoutNodes, config)

	return LayoutResult{
		Nodes:  layoutNodes,
		Width:  width,
		Height: height,
	}
}

// buildGraph constrói mapa de adjacência
func buildGraph(nodes []flow.Node, edges []flow.Edge) map[flow.ID][]flow.ID {
	graph := make(map[flow.ID][]flow.ID)

	// Inicializa todos os nós
	for _, node := range nodes {
		graph[node.ID] = []flow.ID{}
	}

	// Adiciona conexões
	for _, edge := range edges {
		graph[edge.From] = append(graph[edge.From], edge.To)
	}

	return graph
}

// findStartNodes encontra nós sem predecessores
func findStartNodes(graph map[flow.ID][]flow.ID) []flow.ID {
	hasIncoming := make(map[flow.ID]bool)

	// Marca todos os nós que têm conexões de entrada
	for _, targets := range graph {
		for _, target := range targets {
			hasIncoming[target] = true
		}
	}

	// Encontra nós sem conexões de entrada
	var startNodes []flow.ID
	for nodeID := range graph {
		if !hasIncoming[nodeID] {
			startNodes = append(startNodes, nodeID)
		}
	}

	return startNodes
}

// calculateLevels calcula o nível de cada nó usando BFS
func calculateLevels(graph map[flow.ID][]flow.ID, startNodes []flow.ID) map[flow.ID]int {
	levels := make(map[flow.ID]int)
	queue := make([]flow.ID, 0)

	// Inicializa nós de início no nível 0
	for _, startID := range startNodes {
		levels[startID] = 0
		queue = append(queue, startID)
	}

	// BFS para calcular níveis
	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]
		currentLevel := levels[currentID]

		for _, nextID := range graph[currentID] {
			nextLevel := currentLevel + 1
			if existingLevel, exists := levels[nextID]; !exists || nextLevel > existingLevel {
				levels[nextID] = nextLevel
				queue = append(queue, nextID)
			}
		}
	}

	return levels
}

// organizeByLevels organiza nós por nível
func organizeByLevels(nodes []flow.Node, levels map[flow.ID]int) [][]flow.Node {
	if len(levels) == 0 {
		return [][]flow.Node{nodes} // Todos no mesmo nível se não há estrutura
	}

	// Encontra o número máximo de níveis
	maxLevel := 0
	for _, level := range levels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	// Cria slices para cada nível
	nodesByLevel := make([][]flow.Node, maxLevel+1)

	for _, node := range nodes {
		level := levels[node.ID]
		nodesByLevel[level] = append(nodesByLevel[level], node)
	}

	// Ordena nós dentro de cada nível por ID para consistência
	for i := range nodesByLevel {
		sort.Slice(nodesByLevel[i], func(a, b int) bool {
			return string(nodesByLevel[i][a].ID) < string(nodesByLevel[i][b].ID)
		})
	}

	return nodesByLevel
}

// calculatePositions calcula posições finais dos nós
func calculatePositions(nodesByLevel [][]flow.Node, config Config) []LayoutNode {
	var layoutNodes []LayoutNode

	for levelIndex, nodesInLevel := range nodesByLevel {
		if len(nodesInLevel) == 0 {
			continue
		}

		// Calcula dimensões dos nós neste nível
		var totalWidth, maxHeight float64
		var nodeDims []struct{ width, height float64 }

		for _, node := range nodesInLevel {
			width, height := getNodeDimensions(node)
			nodeDims = append(nodeDims, struct{ width, height float64 }{width, height})
			totalWidth += width
			if height > maxHeight {
				maxHeight = height
			}
		}

		// Adiciona espaçamento entre nós
		if len(nodesInLevel) > 1 {
			totalWidth += config.NodeSpacing * float64(len(nodesInLevel)-1)
		}

		// Calcula posições baseado na direção
		for i, node := range nodesInLevel {
			var x, y float64

			if config.Direction == DirectionVertical {
				// Layout vertical: nós organizados horizontalmente por nível
				x = config.StartX - totalWidth/2 + getOffsetForNode(i, nodeDims, config.NodeSpacing)
				y = config.StartY + float64(levelIndex)*config.LevelSpacing
			} else {
				// Layout horizontal: nós organizados verticalmente por nível
				x = config.StartX + float64(levelIndex)*config.LevelSpacing
				y = config.StartY - totalWidth/2 + getOffsetForNode(i, nodeDims, config.NodeSpacing)
			}

			layoutNodes = append(layoutNodes, LayoutNode{
				ID:     node.ID,
				X:      x,
				Y:      y,
				Width:  nodeDims[i].width,
				Height: nodeDims[i].height,
				Level:  levelIndex,
			})
		}
	}

	return layoutNodes
}

// getNodeDimensions obtém dimensões de um nó (usa valores persistidos ou padrão)
func getNodeDimensions(node flow.Node) (width, height float64) {
	// Usa dimensões persistidas se disponíveis
	if node.Width != nil && node.Height != nil {
		return *node.Width, *node.Height
	}

	// Senão usa dimensões padrão baseadas no tipo
	return GetNodeDimensions(node.Kind)
}

// getOffsetForNode calcula o offset X (ou Y) para um nó em um nível
func getOffsetForNode(nodeIndex int, nodeDims []struct{ width, height float64 }, spacing float64) float64 {
	var offset float64

	for i := 0; i < nodeIndex; i++ {
		offset += nodeDims[i].width + spacing
	}

	// Adiciona metade da largura do próprio nó para centralizar
	offset += nodeDims[nodeIndex].width / 2

	return offset
}

// calculateTotalDimensions calcula as dimensões totais do layout
func calculateTotalDimensions(layoutNodes []LayoutNode, config Config) (width, height float64) {
	if len(layoutNodes) == 0 {
		return 0, 0
	}

	var minX, maxX, minY, maxY float64
	first := true

	for _, node := range layoutNodes {
		nodeMinX := node.X - node.Width/2
		nodeMaxX := node.X + node.Width/2
		nodeMinY := node.Y - node.Height/2
		nodeMaxY := node.Y + node.Height/2

		if first {
			minX, maxX = nodeMinX, nodeMaxX
			minY, maxY = nodeMinY, nodeMaxY
			first = false
		} else {
			minX = math.Min(minX, nodeMinX)
			maxX = math.Max(maxX, nodeMaxX)
			minY = math.Min(minY, nodeMinY)
			maxY = math.Max(maxY, nodeMaxY)
		}
	}

	return maxX - minX, maxY - minY
}
