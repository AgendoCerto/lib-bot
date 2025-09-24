// Package service - BotService unificado para manipulação simplificada de bots
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/AgendoCerto/lib-bot/flow"
	"github.com/AgendoCerto/lib-bot/io"
)

// BotService é o serviço principal que unifica todas as operações de bot
// Fornece interface simplificada e integrada com a lib
type BotService struct {
	validation *ValidationService
	design     *DesignService
	store      *StoreService
}

// BotInfo representa informações básicas de um bot
type BotInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Channels    []string          `json:"channels"`
	Version     string            `json:"version"`
	Status      string            `json:"status"`
	NodesCount  int               `json:"nodes_count"`
	EdgesCount  int               `json:"edges_count"`
	Valid       bool              `json:"valid"`
	Issues      []ValidationIssue `json:"issues,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Checksum    string            `json:"checksum"`
}

// NodeInfo representa informações de um nó
type NodeInfo struct {
	ID       string                 `json:"id"`
	Kind     string                 `json:"kind"`
	Title    string                 `json:"title"`
	Props    map[string]interface{} `json:"props"`
	Inputs   []string               `json:"inputs"`
	Outputs  []string               `json:"outputs"`
	Position *Position              `json:"position,omitempty"`
}

// Position representa posição visual do nó
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// EdgeInfo representa informações de uma conexão
type EdgeInfo struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Label    string `json:"label"`
	Output   string `json:"output,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

// ValidationIssue representa um problema de validação
type ValidationIssue struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Path     string `json:"path"`
}

// CreateBotRequest request para criação de bot
type CreateBotRequest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Channels    []string `json:"channels,omitempty"`
	AdapterName string   `json:"adapter_name"`
}

// NewBotService cria uma nova instância do serviço unificado
func NewBotService() *BotService {
	validation := NewValidationService()
	design := NewDesignService(validation)
	store := NewStoreService()

	return &BotService{
		validation: validation,
		design:     design,
		store:      store,
	}
}

// CreateBot cria um novo bot com configuração básica
func (bs *BotService) CreateBot(ctx context.Context, req CreateBotRequest) (*BotInfo, error) {
	// Valores padrão
	if len(req.Channels) == 0 {
		req.Channels = []string{"whatsapp"}
	}
	if req.AdapterName == "" {
		req.AdapterName = "whatsapp"
	}

	// Criar design básico
	design := io.DesignDoc{
		Schema: "flowkit/1.0",
		Bot: io.Bot{
			ID:       req.ID,
			Channels: req.Channels,
		},
		Version: io.Version{
			ID:     "v1.0.0",
			Status: "development",
		},
		Entries: []flow.Entry{
			{
				Kind:   flow.EntryGlobalStart,
				Target: flow.ID("start"),
			},
		},
		Profile: io.Profile{
			Variables: io.ProfileVariables{
				Context: map[string]io.ProfileVariable{
					"user_name": {
						Type:     "string",
						Default:  "",
						Required: false,
					},
				},
				Profile: map[string]any{
					"user_name": "",
				},
			},
		},
		Graph: io.Graph{
			Nodes: []flow.Node{
				{
					ID:      flow.ID("start"),
					Kind:    "message",
					Title:   "Início",
					Props:   map[string]any{"text": "Olá! Como posso ajudar?"},
					Outputs: []string{"complete"},
					X:       &[]float64{100}[0],
					Y:       &[]float64{100}[0],
				},
			},
			Edges: []flow.Edge{},
		},
		Props: make(map[string]any),
	}

	// Validar design
	result, err := bs.validation.ValidateDesign(ctx, design, req.AdapterName)
	if err != nil {
		return nil, fmt.Errorf("erro na validação: %w", err)
	}

	if !result.Valid {
		return nil, fmt.Errorf("design básico inválido: %d issues", len(result.Issues))
	}

	// Salvar no store
	versionInfo, err := bs.store.Save(ctx, req.ID, design)
	if err != nil {
		return nil, fmt.Errorf("erro ao salvar: %w", err)
	}

	// Retornar informações do bot
	return &BotInfo{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Channels:    req.Channels,
		Version:     versionInfo.ID,
		Status:      "development",
		NodesCount:  1,
		EdgesCount:  0,
		Valid:       true,
		CreatedAt:   versionInfo.CreatedAt,
		UpdatedAt:   versionInfo.UpdatedAt,
		Checksum:    versionInfo.Checksum,
	}, nil
}

// GetBot obtém informações completas de um bot
func (bs *BotService) GetBot(ctx context.Context, botID string) (*BotInfo, error) {
	// Verificar se existe
	exists, err := bs.store.Exists(ctx, botID)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar existência: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("bot %s não encontrado", botID)
	}

	// Carregar design
	design, err := bs.store.Load(ctx, botID)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar design: %w", err)
	}

	// Obter versão
	versionInfo, err := bs.store.GetVersion(ctx, botID)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter versão: %w", err)
	}

	// Validar
	result, err := bs.validation.ValidateDesign(ctx, design, "whatsapp")
	if err != nil {
		return nil, fmt.Errorf("erro na validação: %w", err)
	}

	// Converter issues
	var issues []ValidationIssue
	for _, issue := range result.Issues {
		issues = append(issues, ValidationIssue{
			Code:     issue.Code,
			Severity: string(issue.Severity),
			Message:  issue.Msg,
			Path:     issue.Path,
		})
	}

	return &BotInfo{
		ID:         design.Bot.ID,
		Channels:   design.Bot.Channels,
		Version:    design.Version.ID,
		Status:     design.Version.Status,
		NodesCount: len(design.Graph.Nodes),
		EdgesCount: len(design.Graph.Edges),
		Valid:      result.Valid,
		Issues:     issues,
		CreatedAt:  versionInfo.CreatedAt,
		UpdatedAt:  versionInfo.UpdatedAt,
		Checksum:   versionInfo.Checksum,
	}, nil
}

// ListBots lista todos os bots
func (bs *BotService) ListBots(ctx context.Context) ([]*BotInfo, error) {
	versions, err := bs.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar: %w", err)
	}

	var bots []*BotInfo
	for _, version := range versions {
		botInfo, err := bs.GetBot(ctx, version.BotID)
		if err != nil {
			// Log error but continue
			continue
		}
		bots = append(bots, botInfo)
	}

	return bots, nil
}

// DeleteBot remove um bot
func (bs *BotService) DeleteBot(ctx context.Context, botID string) error {
	return bs.store.Delete(ctx, botID)
}

// AddMessageNode adiciona um nó de mensagem
func (bs *BotService) AddMessageNode(ctx context.Context, botID, nodeID, message string, position *Position) (*NodeInfo, error) {
	return bs.addNode(ctx, botID, NodeInfo{
		ID:       nodeID,
		Kind:     "message",
		Title:    "Mensagem",
		Props:    map[string]interface{}{"text": message},
		Outputs:  []string{"complete"},
		Position: position,
	})
}

// AddConfirmNode adiciona um nó de confirmação
func (bs *BotService) AddConfirmNode(ctx context.Context, botID, nodeID, question, yesLabel, noLabel string, position *Position) (*NodeInfo, error) {
	return bs.addNode(ctx, botID, NodeInfo{
		ID:    nodeID,
		Kind:  "confirm",
		Title: "Confirmação",
		Props: map[string]interface{}{
			"title": question,
			"yes":   yesLabel,
			"no":    noLabel,
		},
		Outputs:  []string{"confirmed", "cancelled", "timeout"},
		Position: position,
	})
}

// AddInputNode adiciona um nó de entrada de dados (usando text como input)
func (bs *BotService) AddInputNode(ctx context.Context, botID, nodeID, prompt, placeholder string, position *Position) (*NodeInfo, error) {
	return bs.addNode(ctx, botID, NodeInfo{
		ID:    nodeID,
		Kind:  "text",
		Title: "Entrada de Texto",
		Props: map[string]interface{}{
			"body": prompt,
			"text": prompt, // fallback
		},
		Outputs:  []string{"complete"},
		Position: position,
	})
}

// AddListPickerNode adiciona um nó de seleção de lista
func (bs *BotService) AddListPickerNode(ctx context.Context, botID, nodeID, prompt string, options []string, position *Position) (*NodeInfo, error) {
	// Converter opções em formato do listpicker
	var sections []map[string]interface{}
	var items []map[string]interface{}

	for i, option := range options {
		items = append(items, map[string]interface{}{
			"id":    fmt.Sprintf("option_%d", i),
			"title": option,
		})
	}

	sections = append(sections, map[string]interface{}{
		"title": "Opções",
		"items": items,
	})

	return bs.addNode(ctx, botID, NodeInfo{
		ID:    nodeID,
		Kind:  "listpicker",
		Title: "Seleção de Lista",
		Props: map[string]interface{}{
			"text":     prompt,
			"sections": sections,
		},
		Outputs:  []string{"complete"},
		Position: position,
	})
}

// AddDelayNode adiciona um nó de delay
func (bs *BotService) AddDelayNode(ctx context.Context, botID, nodeID string, seconds int, position *Position) (*NodeInfo, error) {
	return bs.addNode(ctx, botID, NodeInfo{
		ID:    nodeID,
		Kind:  "delay",
		Title: "Delay",
		Props: map[string]interface{}{
			"duration": seconds * 1000, // converter para millisegundos
			"unit":     "milliseconds",
			"reason":   "processing",
		},
		Outputs:  []string{"complete"},
		Position: position,
	})
}

// UpdateNode atualiza propriedades de um nó
func (bs *BotService) UpdateNode(ctx context.Context, botID string, nodeInfo NodeInfo) (*NodeInfo, error) {
	// Carregar design atual
	design, err := bs.store.Load(ctx, botID)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar design: %w", err)
	}

	// Encontrar e atualizar nó
	found := false
	for i, node := range design.Graph.Nodes {
		if string(node.ID) == nodeInfo.ID {
			// Atualizar nó
			updatedNode := flow.Node{
				ID:      flow.ID(nodeInfo.ID),
				Kind:    nodeInfo.Kind,
				Title:   nodeInfo.Title,
				Props:   nodeInfo.Props,
				Outputs: nodeInfo.Outputs,
				Inputs:  nodeInfo.Inputs,
			}

			if nodeInfo.Position != nil {
				updatedNode.X = &nodeInfo.Position.X
				updatedNode.Y = &nodeInfo.Position.Y
			}

			design.Graph.Nodes[i] = updatedNode
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("nó %s não encontrado", nodeInfo.ID)
	}

	// Validar e salvar
	result, err := bs.validation.ValidateDesign(ctx, design, "whatsapp")
	if err != nil {
		return nil, fmt.Errorf("erro na validação: %w", err)
	}

	if !result.Valid {
		return nil, fmt.Errorf("design inválido após atualização: %d issues", len(result.Issues))
	}

	_, err = bs.store.Save(ctx, botID, design)
	if err != nil {
		return nil, fmt.Errorf("erro ao salvar: %w", err)
	}

	return &nodeInfo, nil
}

// RemoveNode remove um nó
func (bs *BotService) RemoveNode(ctx context.Context, botID, nodeID string) error {
	// Carregar design atual
	design, err := bs.store.Load(ctx, botID)
	if err != nil {
		return fmt.Errorf("erro ao carregar design: %w", err)
	}

	// Remover nó
	nodeIndex := -1
	for i, node := range design.Graph.Nodes {
		if string(node.ID) == nodeID {
			nodeIndex = i
			break
		}
	}

	if nodeIndex == -1 {
		return fmt.Errorf("nó %s não encontrado", nodeID)
	}

	// Remove node
	design.Graph.Nodes = append(design.Graph.Nodes[:nodeIndex], design.Graph.Nodes[nodeIndex+1:]...)

	// Remove edges relacionadas
	var filteredEdges []flow.Edge
	for _, edge := range design.Graph.Edges {
		if string(edge.From) != nodeID && string(edge.To) != nodeID {
			filteredEdges = append(filteredEdges, edge)
		}
	}
	design.Graph.Edges = filteredEdges

	// Validar e salvar
	result, err := bs.validation.ValidateDesign(ctx, design, "whatsapp")
	if err != nil {
		return fmt.Errorf("erro na validação: %w", err)
	}

	if !result.Valid {
		return fmt.Errorf("design inválido após remoção: %d issues", len(result.Issues))
	}

	_, err = bs.store.Save(ctx, botID, design)
	if err != nil {
		return fmt.Errorf("erro ao salvar: %w", err)
	}

	return nil
}

// ConnectNodes conecta dois nós
func (bs *BotService) ConnectNodes(ctx context.Context, botID string, edge EdgeInfo) (*EdgeInfo, error) {
	// Carregar design atual
	design, err := bs.store.Load(ctx, botID)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar design: %w", err)
	}

	// Verificar se nós existem
	if !bs.nodeExists(design, edge.From) {
		return nil, fmt.Errorf("nó origem %s não encontrado", edge.From)
	}
	if !bs.nodeExists(design, edge.To) {
		return nil, fmt.Errorf("nó destino %s não encontrado", edge.To)
	}

	// Criar edge
	newEdge := flow.Edge{
		From:     flow.ID(edge.From),
		To:       flow.ID(edge.To),
		Label:    edge.Label,
		Priority: edge.Priority,
	}

	design.Graph.Edges = append(design.Graph.Edges, newEdge)

	// Validar e salvar
	result, err := bs.validation.ValidateDesign(ctx, design, "whatsapp")
	if err != nil {
		return nil, fmt.Errorf("erro na validação: %w", err)
	}

	if !result.Valid {
		return nil, fmt.Errorf("design inválido após conexão: %d issues", len(result.Issues))
	}

	_, err = bs.store.Save(ctx, botID, design)
	if err != nil {
		return nil, fmt.Errorf("erro ao salvar: %w", err)
	}

	return &edge, nil
}

// DisconnectNodes desconecta dois nós
func (bs *BotService) DisconnectNodes(ctx context.Context, botID, fromNodeID, toNodeID string) error {
	// Carregar design atual
	design, err := bs.store.Load(ctx, botID)
	if err != nil {
		return fmt.Errorf("erro ao carregar design: %w", err)
	}

	// Encontrar e remover edge
	edgeIndex := -1
	for i, edge := range design.Graph.Edges {
		if string(edge.From) == fromNodeID && string(edge.To) == toNodeID {
			edgeIndex = i
			break
		}
	}

	if edgeIndex == -1 {
		return fmt.Errorf("conexão %s -> %s não encontrada", fromNodeID, toNodeID)
	}

	// Remove edge
	design.Graph.Edges = append(design.Graph.Edges[:edgeIndex], design.Graph.Edges[edgeIndex+1:]...)

	// Salvar
	_, err = bs.store.Save(ctx, botID, design)
	if err != nil {
		return fmt.Errorf("erro ao salvar: %w", err)
	}

	return nil
}

// GetNodes obtém todos os nós de um bot
func (bs *BotService) GetNodes(ctx context.Context, botID string) ([]*NodeInfo, error) {
	design, err := bs.store.Load(ctx, botID)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar design: %w", err)
	}

	var nodes []*NodeInfo
	for _, node := range design.Graph.Nodes {
		nodeInfo := &NodeInfo{
			ID:      string(node.ID),
			Kind:    node.Kind,
			Title:   node.Title,
			Props:   node.Props,
			Inputs:  node.Inputs,
			Outputs: node.Outputs,
		}

		if node.X != nil && node.Y != nil {
			nodeInfo.Position = &Position{X: *node.X, Y: *node.Y}
		}

		nodes = append(nodes, nodeInfo)
	}

	return nodes, nil
}

// GetEdges obtém todas as conexões de um bot
func (bs *BotService) GetEdges(ctx context.Context, botID string) ([]*EdgeInfo, error) {
	design, err := bs.store.Load(ctx, botID)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar design: %w", err)
	}

	var edges []*EdgeInfo
	for _, edge := range design.Graph.Edges {
		edges = append(edges, &EdgeInfo{
			From:     string(edge.From),
			To:       string(edge.To),
			Label:    edge.Label,
			Priority: edge.Priority,
		})
	}

	return edges, nil
}

// ValidateBot valida um bot
func (bs *BotService) ValidateBot(ctx context.Context, botID, adapterName string) (*ValidationResult, error) {
	design, err := bs.store.Load(ctx, botID)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar design: %w", err)
	}

	return bs.validation.ValidateDesign(ctx, design, adapterName)
}

// CloneBot cria uma cópia de um bot
func (bs *BotService) CloneBot(ctx context.Context, sourceBotID, targetBotID string) (*BotInfo, error) {
	_, err := bs.store.Clone(ctx, sourceBotID, targetBotID)
	if err != nil {
		return nil, fmt.Errorf("erro ao clonar: %w", err)
	}

	return bs.GetBot(ctx, targetBotID)
}

// GetBotDesign obtém o design completo de um bot (para exportação/importação)
func (bs *BotService) GetBotDesign(ctx context.Context, botID string) (*io.DesignDoc, error) {
	design, err := bs.store.Load(ctx, botID)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar design: %w", err)
	}
	return &design, nil
}

// ImportBotDesign importa um design completo
func (bs *BotService) ImportBotDesign(ctx context.Context, botID string, design io.DesignDoc, adapterName string) (*BotInfo, error) {
	// Atualizar ID do bot no design
	design.Bot.ID = botID

	// Validar
	result, err := bs.validation.ValidateDesign(ctx, design, adapterName)
	if err != nil {
		return nil, fmt.Errorf("erro na validação: %w", err)
	}

	if !result.Valid {
		return nil, fmt.Errorf("design importado inválido: %d issues", len(result.Issues))
	}

	// Salvar
	_, err = bs.store.Save(ctx, botID, design)
	if err != nil {
		return nil, fmt.Errorf("erro ao salvar: %w", err)
	}

	return bs.GetBot(ctx, botID)
}

// Métodos auxiliares privados

func (bs *BotService) addNode(ctx context.Context, botID string, nodeInfo NodeInfo) (*NodeInfo, error) {
	// Carregar design atual
	design, err := bs.store.Load(ctx, botID)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar design: %w", err)
	}

	// Verificar se nó já existe
	for _, node := range design.Graph.Nodes {
		if string(node.ID) == nodeInfo.ID {
			return nil, fmt.Errorf("nó %s já existe", nodeInfo.ID)
		}
	}

	// Criar nó
	newNode := flow.Node{
		ID:      flow.ID(nodeInfo.ID),
		Kind:    nodeInfo.Kind,
		Title:   nodeInfo.Title,
		Props:   nodeInfo.Props,
		Outputs: nodeInfo.Outputs,
		Inputs:  nodeInfo.Inputs,
	}

	// Definir posição
	if nodeInfo.Position != nil {
		newNode.X = &nodeInfo.Position.X
		newNode.Y = &nodeInfo.Position.Y
	} else {
		// Posição padrão baseada no número de nós existentes
		x := float64(200 + len(design.Graph.Nodes)*150)
		y := float64(200)
		newNode.X = &x
		newNode.Y = &y
	}

	// Adicionar ao design
	design.Graph.Nodes = append(design.Graph.Nodes, newNode)

	// Validar
	result, err := bs.validation.ValidateDesign(ctx, design, "whatsapp")
	if err != nil {
		return nil, fmt.Errorf("erro na validação: %w", err)
	}

	if !result.Valid {
		return nil, fmt.Errorf("design inválido após adicionar nó: %d issues", len(result.Issues))
	}

	// Salvar
	_, err = bs.store.Save(ctx, botID, design)
	if err != nil {
		return nil, fmt.Errorf("erro ao salvar: %w", err)
	}

	// Retornar informações do nó criado
	if nodeInfo.Position == nil && newNode.X != nil && newNode.Y != nil {
		nodeInfo.Position = &Position{X: *newNode.X, Y: *newNode.Y}
	}

	return &nodeInfo, nil
}

func (bs *BotService) nodeExists(design io.DesignDoc, nodeID string) bool {
	for _, node := range design.Graph.Nodes {
		if string(node.ID) == nodeID {
			return true
		}
	}
	return false
}
