package layout

import (
	"github.com/AgendoCerto/lib-bot/flow"
)

type Direction string

const (
	DirectionVertical   Direction = "vertical"
	DirectionHorizontal Direction = "horizontal"
)

type Config struct {
	Direction    Direction
	NodeSpacing  float64
	LevelSpacing float64
	StartX       float64
	StartY       float64
}

func DefaultConfig() Config {
	return Config{
		Direction:    DirectionVertical,
		NodeSpacing:  150,
		LevelSpacing: 100,
		StartX:       100,
		StartY:       100,
	}
}

func ApplyAutoLayout(graph *flow.Graph, config Config) error {
	if len(graph.Nodes) == 0 {
		return nil
	}

	levels := buildLevels(graph.Nodes, graph.Edges)

	switch config.Direction {
	case DirectionVertical:
		applyVerticalLayout(levels, config, graph)
	case DirectionHorizontal:
		applyHorizontalLayout(levels, config, graph)
	}

	return nil
}

func buildLevels(nodes []flow.Node, edges []flow.Edge) [][]flow.ID {
	nodeExists := make(map[flow.ID]bool)
	for _, node := range nodes {
		nodeExists[node.ID] = true
	}

	predecessors := make(map[flow.ID][]flow.ID)
	successors := make(map[flow.ID][]flow.ID)

	for _, edge := range edges {
		if nodeExists[edge.From] && nodeExists[edge.To] {
			predecessors[edge.To] = append(predecessors[edge.To], edge.From)
			successors[edge.From] = append(successors[edge.From], edge.To)
		}
	}

	var startNodes []flow.ID
	for _, node := range nodes {
		if len(predecessors[node.ID]) == 0 {
			startNodes = append(startNodes, node.ID)
		}
	}

	if len(startNodes) == 0 && len(nodes) > 0 {
		startNodes = append(startNodes, nodes[0].ID)
	}

	levels := [][]flow.ID{}
	visited := make(map[flow.ID]bool)

	if len(startNodes) > 0 {
		levels = append(levels, startNodes)
		for _, nodeID := range startNodes {
			visited[nodeID] = true
		}
	}

	for levelIndex := 0; levelIndex < len(levels); levelIndex++ {
		var nextLevel []flow.ID

		for _, nodeID := range levels[levelIndex] {
			for _, successor := range successors[nodeID] {
				if !visited[successor] {
					allPredsVisited := true
					for _, pred := range predecessors[successor] {
						if !visited[pred] {
							allPredsVisited = false
							break
						}
					}

					if allPredsVisited {
						nextLevel = append(nextLevel, successor)
						visited[successor] = true
					}
				}
			}
		}

		if len(nextLevel) > 0 {
			levels = append(levels, nextLevel)
		}
	}

	var isolatedNodes []flow.ID
	for _, node := range nodes {
		if !visited[node.ID] {
			isolatedNodes = append(isolatedNodes, node.ID)
		}
	}
	if len(isolatedNodes) > 0 {
		levels = append(levels, isolatedNodes)
	}

	return levels
}

func applyVerticalLayout(levels [][]flow.ID, config Config, graph *flow.Graph) {
	nodeMap := make(map[flow.ID]*flow.Node)
	for i := range graph.Nodes {
		nodeMap[graph.Nodes[i].ID] = &graph.Nodes[i]
	}

	for levelIndex, level := range levels {
		y := config.StartY + float64(levelIndex)*config.LevelSpacing

		for nodeIndex, nodeID := range level {
			if node, exists := nodeMap[nodeID]; exists {
				x := config.StartX + float64(nodeIndex)*config.NodeSpacing
				node.X = &x
				node.Y = &y
			}
		}
	}
}

func applyHorizontalLayout(levels [][]flow.ID, config Config, graph *flow.Graph) {
	nodeMap := make(map[flow.ID]*flow.Node)
	for i := range graph.Nodes {
		nodeMap[graph.Nodes[i].ID] = &graph.Nodes[i]
	}

	for levelIndex, level := range levels {
		x := config.StartX + float64(levelIndex)*config.LevelSpacing

		for nodeIndex, nodeID := range level {
			if node, exists := nodeMap[nodeID]; exists {
				y := config.StartY + float64(nodeIndex)*config.NodeSpacing
				node.X = &x
				node.Y = &y
			}
		}
	}
}
