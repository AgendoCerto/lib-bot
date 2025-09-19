// Package layout fornece algoritmos de auto-layout para organização automática de nós
package layout

import (
	"math"
	"sort"
	"strings"

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
	"text":       {Width: 250, Height: 80},
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

// AutoLayout aplica algoritmo de posicionamento automático para nós no React Flow
func AutoLayout(nodes []flow.Node, edges []flow.Edge, config Config) LayoutResult {
	if len(nodes) == 0 {
		return LayoutResult{}
	}

	// 1. Separa nós verdadeiramente isolados dos conectados ao fluxo
	connectedNodes, isolatedNodes := separateConnectedAndTrulyIsolated(nodes, edges)

	// 2. Constrói grafo apenas com nós conectados e calcula níveis
	graph := buildGraph(connectedNodes, edges)
	levels := calculateLevels(graph, findStartNodes(graph))

	// 3. Organiza nós conectados por nível
	nodesByLevel := organizeByLevels(connectedNodes, levels, edges)

	// 4. Calcula posições dos nós conectados
	layoutNodes := calculatePositions(nodesByLevel, config)

	// 5. Adiciona nós isolados próximos ao fluxo principal
	if len(isolatedNodes) > 0 {
		isolatedLayoutNodes := positionIsolatedNodes(isolatedNodes, layoutNodes, config)
		layoutNodes = append(layoutNodes, isolatedLayoutNodes...)
	}

	// 6. Calcula dimensões totais
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

// calculateLevels calcula o nível de cada nó usando análise de fluxo inteligente
func calculateLevels(graph map[flow.ID][]flow.ID, startNodes []flow.ID) map[flow.ID]int {
	levels := make(map[flow.ID]int)
	incomingCount := make(map[flow.ID]int)

	// Conta conexões de entrada para cada nó
	for from, targets := range graph {
		if _, exists := incomingCount[from]; !exists {
			incomingCount[from] = 0
		}
		for _, to := range targets {
			incomingCount[to]++
		}
	}

	// Processamento topológico modificado para considerar fluxo
	queue := make([]flow.ID, 0)

	// Inicializa nós de início no nível 0
	for _, startID := range startNodes {
		levels[startID] = 0
		queue = append(queue, startID)
	}

	// Processamento por ondas para criar níveis mais lógicos
	for len(queue) > 0 {
		currentWave := make([]flow.ID, len(queue))
		copy(currentWave, queue)
		queue = queue[:0] // limpa para próxima onda

		for _, currentID := range currentWave {
			for _, nextID := range graph[currentID] {
				// Verifica se todos os predecessores já foram processados
				allPredecessorsProcessed := true
				maxPredecessorLevel := -1

				for from, targets := range graph {
					for _, to := range targets {
						if to == nextID {
							if predLevel, exists := levels[from]; exists {
								if predLevel > maxPredecessorLevel {
									maxPredecessorLevel = predLevel
								}
							} else {
								allPredecessorsProcessed = false
								break
							}
						}
					}
					if !allPredecessorsProcessed {
						break
					}
				}

				if allPredecessorsProcessed {
					finalLevel := maxPredecessorLevel + 1
					if existingLevel, exists := levels[nextID]; !exists || finalLevel > existingLevel {
						levels[nextID] = finalLevel
						// Adiciona à próxima onda se ainda não foi processado
						shouldAdd := true
						for _, id := range queue {
							if id == nextID {
								shouldAdd = false
								break
							}
						}
						if shouldAdd {
							queue = append(queue, nextID)
						}
					}
				}
			}
		}
	}

	return levels
}

// organizeByLevels organiza nós por nível
func organizeByLevels(nodes []flow.Node, levels map[flow.ID]int, edges []flow.Edge) [][]flow.Node {
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
		if level, exists := levels[node.ID]; exists {
			nodesByLevel[level] = append(nodesByLevel[level], node)
		}
	}

	// Ordena nós dentro de cada nível baseado em conexões, não apenas ID
	for i := range nodesByLevel {
		organizeNodesInLevel(nodesByLevel[i], edges)
	}

	return nodesByLevel
}

// organizeNodesInLevel ordena nós dentro de um nível baseado em suas conexões
func organizeNodesInLevel(levelNodes []flow.Node, edges []flow.Edge) {
	if len(levelNodes) <= 1 {
		return
	}

	// Constrói mapa de edges para análise de conexões
	edgeMap := make(map[flow.ID][]flow.ID)
	incomingEdges := make(map[flow.ID][]flow.ID)

	for _, edge := range edges {
		edgeMap[edge.From] = append(edgeMap[edge.From], edge.To)
		incomingEdges[edge.To] = append(incomingEdges[edge.To], edge.From)
	}

	// Identifica nós que fazem parte de caminhos paralelos
	// (têm o mesmo predecessor comum)
	parallelGroups := make(map[flow.ID][]flow.Node)

	for _, node := range levelNodes {
		// Para cada nó, encontra seus predecessores diretos
		predecessors := incomingEdges[node.ID]

		// Se tem exatamente um predecessor, pode ser parte de um grupo paralelo
		if len(predecessors) == 1 {
			pred := predecessors[0]
			parallelGroups[pred] = append(parallelGroups[pred], node)
		}
	}

	// Ordena baseado na lógica de fluxo melhorada:
	sort.Slice(levelNodes, func(a, b int) bool {
		nodeA, nodeB := levelNodes[a], levelNodes[b]

		// Critério 1: Prioridade para nós sequenciais (1 entrada, 1 saída)
		incomingA, outgoingA := len(incomingEdges[nodeA.ID]), len(edgeMap[nodeA.ID])
		incomingB, outgoingB := len(incomingEdges[nodeB.ID]), len(edgeMap[nodeB.ID])

		isSequentialA := incomingA == 1 && outgoingA == 1
		isSequentialB := incomingB == 1 && outgoingB == 1

		if isSequentialA != isSequentialB {
			return isSequentialA // Sequenciais primeiro
		}

		// Critério 2: Nós com mais conexões totais (mais importantes no fluxo)
		totalConnectionsA := incomingA + outgoingA
		totalConnectionsB := incomingB + outgoingB

		if totalConnectionsA != totalConnectionsB {
			return totalConnectionsA > totalConnectionsB
		}

		// Critério 3: Para nós paralelos, ordena por predecessor comum
		predA := getPrimaryPredecessor(nodeA.ID, incomingEdges)
		predB := getPrimaryPredecessor(nodeB.ID, incomingEdges)

		if predA != predB {
			return string(predA) < string(predB)
		}

		// Critério 4: ID para consistência
		return string(nodeA.ID) < string(nodeB.ID)
	})
}

// getPrimaryPredecessor retorna o predecessor principal de um nó
func getPrimaryPredecessor(nodeID flow.ID, incomingEdges map[flow.ID][]flow.ID) flow.ID {
	predecessors := incomingEdges[nodeID]
	if len(predecessors) == 0 {
		return ""
	}
	// Para simplificar, retorna o primeiro predecessor
	// Em uma implementação mais sofisticada, poderia analisar qual é o "principal"
	return predecessors[0]
}

// separateConnectedAndTrulyIsolated separa nós que têm conexões dos verdadeiramente isolados
func separateConnectedAndTrulyIsolated(nodes []flow.Node, edges []flow.Edge) (connected, isolated []flow.Node) {
	// Cria mapa de nós que aparecem em edges
	connectedNodeIDs := make(map[flow.ID]bool)

	for _, edge := range edges {
		connectedNodeIDs[edge.From] = true
		connectedNodeIDs[edge.To] = true
	}

	// Separa nós baseado na presença em edges
	for _, node := range nodes {
		if connectedNodeIDs[node.ID] {
			connected = append(connected, node)
		} else {
			isolated = append(isolated, node)
		}
	}

	return connected, isolated
}

// positionIsolatedNodes posiciona nós isolados próximos ao fluxo principal
func positionIsolatedNodes(isolatedNodes []flow.Node, connectedNodes []LayoutNode, config Config) []LayoutNode {
	if len(isolatedNodes) == 0 {
		return []LayoutNode{}
	}

	// Encontra a extremidade direita/inferior do fluxo principal
	var maxX, maxY float64
	var minX, minY float64
	first := true

	for _, node := range connectedNodes {
		nodeMaxX := node.X + node.Width/2
		nodeMaxY := node.Y + node.Height/2
		nodeMinX := node.X - node.Width/2
		nodeMinY := node.Y - node.Height/2

		if first {
			maxX, maxY = nodeMaxX, nodeMaxY
			minX, minY = nodeMinX, nodeMinY
			first = false
		} else {
			if nodeMaxX > maxX {
				maxX = nodeMaxX
			}
			if nodeMaxY > maxY {
				maxY = nodeMaxY
			}
			if nodeMinX < minX {
				minX = nodeMinX
			}
			if nodeMinY < minY {
				minY = nodeMinY
			}
		}
	}

	var isolatedLayoutNodes []LayoutNode

	// Configuração para nós isolados
	isolatedSpacing := config.NodeSpacing * 1.5 // Espaçamento maior entre nós isolados
	var isolatedGap float64 = 150               // Distância do fluxo principal

	if config.Direction == DirectionVertical {
		// Layout vertical: coloca nós isolados à direita do fluxo principal
		startX := maxX + isolatedGap
		startY := minY

		for i, node := range isolatedNodes {
			width, height := getNodeDimensions(node)

			isolatedLayoutNodes = append(isolatedLayoutNodes, LayoutNode{
				ID:     node.ID,
				X:      startX,
				Y:      startY + float64(i)*(height+isolatedSpacing),
				Width:  width,
				Height: height,
				Level:  -1, // Marca como nó isolado
			})
		}
	} else {
		// Layout horizontal: coloca nós isolados abaixo do fluxo principal
		startX := minX
		startY := maxY + isolatedGap

		for i, node := range isolatedNodes {
			width, height := getNodeDimensions(node)

			isolatedLayoutNodes = append(isolatedLayoutNodes, LayoutNode{
				ID:     node.ID,
				X:      startX + float64(i)*(width+isolatedSpacing),
				Y:      startY,
				Width:  width,
				Height: height,
				Level:  -1, // Marca como nó isolado
			})
		}
	}

	return isolatedLayoutNodes
}

// calculatePositions calcula posições finais dos nós
func calculatePositions(nodesByLevel [][]flow.Node, config Config) []LayoutNode {
	var layoutNodes []LayoutNode

	for levelIndex, nodesInLevel := range nodesByLevel {
		if len(nodesInLevel) == 0 {
			continue
		}

		// Separa nós principais de nós de fallback
		mainNodes, fallbackNodes := separateFallbackNodes(nodesInLevel)

		// Posiciona nós principais
		layoutNodes = append(layoutNodes, positionMainNodes(mainNodes, levelIndex, config)...)

		// Posiciona nós de fallback lateralmente
		if len(fallbackNodes) > 0 {
			layoutNodes = append(layoutNodes, positionFallbackNodes(fallbackNodes, levelIndex, config)...)
		}
	}

	return layoutNodes
}

// separateFallbackNodes separa nós principais de nós de fallback baseado no ID
func separateFallbackNodes(nodes []flow.Node) (main, fallback []flow.Node) {
	for _, node := range nodes {
		nodeID := strings.ToLower(string(node.ID))
		if isFallbackNode(nodeID) {
			fallback = append(fallback, node)
		} else {
			main = append(main, node)
		}
	}
	return main, fallback
}

// isFallbackNode verifica se um nó é de fallback baseado no ID
func isFallbackNode(nodeID string) bool {
	fallbackKeywords := []string{"erro", "error", "timeout", "invalid", "fallback", "retry"}
	for _, keyword := range fallbackKeywords {
		if strings.Contains(nodeID, keyword) {
			return true
		}
	}
	return false
}

// positionMainNodes posiciona nós principais normalmente
func positionMainNodes(nodes []flow.Node, levelIndex int, config Config) []LayoutNode {
	if len(nodes) == 0 {
		return []LayoutNode{}
	}

	var layoutNodes []LayoutNode
	var totalWidth, maxHeight float64
	var nodeDims []struct{ width, height float64 }

	for _, node := range nodes {
		width, height := getNodeDimensions(node)
		nodeDims = append(nodeDims, struct{ width, height float64 }{width, height})
		totalWidth += width
		if height > maxHeight {
			maxHeight = height
		}
	}

	// Adiciona espaçamento entre nós
	if len(nodes) > 1 {
		totalWidth += config.NodeSpacing * float64(len(nodes)-1)
	}

	// Calcula posições baseado na direção
	for i, node := range nodes {
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

	return layoutNodes
}

// positionFallbackNodes posiciona nós de fallback lateralmente
func positionFallbackNodes(nodes []flow.Node, levelIndex int, config Config) []LayoutNode {
	if len(nodes) == 0 {
		return []LayoutNode{}
	}

	var layoutNodes []LayoutNode
	fallbackSpacing := config.NodeSpacing * 1.5 // Espaçamento maior para fallbacks

	for i, node := range nodes {
		width, height := getNodeDimensions(node)
		var x, y float64

		if config.Direction == DirectionVertical {
			// Nós de fallback ficam à direita do fluxo principal
			x = config.StartX + 400 + float64(i)*fallbackSpacing
			y = config.StartY + float64(levelIndex)*config.LevelSpacing
		} else {
			// Nós de fallback ficam abaixo do fluxo principal
			x = config.StartX + float64(levelIndex)*config.LevelSpacing
			y = config.StartY + 300 + float64(i)*fallbackSpacing
		}

		layoutNodes = append(layoutNodes, LayoutNode{
			ID:     node.ID,
			X:      x,
			Y:      y,
			Width:  width,
			Height: height,
			Level:  levelIndex,
		})
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
