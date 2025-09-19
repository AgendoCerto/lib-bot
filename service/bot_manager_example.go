// Package service - Exemplo completo de uso dos serviços para criação e gerenciamento de bots
package service

import (
	"context"
	"fmt"
	"log"

	"lib-bot/flow"
	"lib-bot/io"
)

// BotManager demonstra como usar os serviços juntos
type BotManager struct {
	validationService *ValidationService
	designService     *DesignService
	storeService      *StoreService
}

// NewBotManager cria um novo gerenciador de bots
func NewBotManager() *BotManager {
	validationService := NewValidationService()
	designService := NewDesignService(validationService)
	storeService := NewStoreService()

	return &BotManager{
		validationService: validationService,
		designService:     designService,
		storeService:      storeService,
	}
}

// CreateBot demonstra como criar um novo bot do zero
func (bm *BotManager) CreateBot(ctx context.Context, botID string, adapterName string) error {
	fmt.Printf("=== Criando novo bot: %s ===\n", botID)

	// 1. Criar design básico
	design := io.DesignDoc{
		Schema: "flowkit/1.0",
		Bot: io.Bot{
			ID:       botID,
			Channels: []string{"whatsapp"},
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
			Context: map[string]io.ProfileVariable{
				"user_name": {
					Type:     "string",
					Default:  "",
					Required: false,
				},
			},
		},
		Graph: io.Graph{
			Nodes: []flow.Node{
				{
					ID:    flow.ID("start"),
					Kind:  "message",
					Title: "Mensagem de Boas-vindas",
					Props: map[string]any{
						"text": "Olá! Bem-vindo ao nosso atendimento. Como posso ajudar?",
					},
					Outputs: []string{"complete"},
					X:       float64Ptr(100),
					Y:       float64Ptr(100),
				},
			},
			Edges: []flow.Edge{},
		},
		Props: make(map[string]any),
	}

	// 2. Criar o design usando DesignService
	codec := io.JSONCodec{}
	designJSON, err := codec.EncodeDesign(design)
	if err != nil {
		return fmt.Errorf("erro ao codificar design: %w", err)
	}

	createdDesign, err := bm.designService.Create(ctx, designJSON, adapterName)
	if err != nil {
		return fmt.Errorf("erro ao criar design: %w", err)
	}

	// 3. Salvar no store
	versionInfo, err := bm.storeService.Save(ctx, botID, createdDesign)
	if err != nil {
		return fmt.Errorf("erro ao salvar no store: %w", err)
	}

	fmt.Printf("✅ Bot criado com sucesso: %s (versão: %s)\n", botID, versionInfo.ID)
	return nil
}

// AddMessageNode demonstra como adicionar um novo nó de mensagem
func (bm *BotManager) AddMessageNode(ctx context.Context, botID string, nodeID string, message string, adapterName string) error {
	fmt.Printf("=== Adicionando nó de mensagem: %s ===\n", nodeID)

	// 1. Carregar design atual do store
	design, err := bm.storeService.Load(ctx, botID)
	if err != nil {
		return fmt.Errorf("erro ao carregar design: %w", err)
	}

	// 2. Criar novo nó
	newNode := flow.Node{
		ID:    flow.ID(nodeID),
		Kind:  "message",
		Title: "Nova Mensagem",
		Props: map[string]any{
			"text": message,
		},
		Outputs: []string{"complete"},
		X:       float64Ptr(300),
		Y:       float64Ptr(200),
	}

	// 3. Adicionar nó usando DesignService
	updatedDesign, err := bm.designService.AddNode(ctx, design, newNode, adapterName)
	if err != nil {
		return fmt.Errorf("erro ao adicionar nó: %w", err)
	}

	// 4. Salvar design atualizado
	_, err = bm.storeService.Save(ctx, botID, updatedDesign)
	if err != nil {
		return fmt.Errorf("erro ao salvar design atualizado: %w", err)
	}

	fmt.Printf("✅ Nó %s adicionado com sucesso\n", nodeID)
	return nil
}

// ConnectNodes demonstra como conectar dois nós
func (bm *BotManager) ConnectNodes(ctx context.Context, botID string, fromNodeID, toNodeID string, label string, adapterName string) error {
	fmt.Printf("=== Conectando nós: %s -> %s ===\n", fromNodeID, toNodeID)

	// 1. Carregar design atual
	design, err := bm.storeService.Load(ctx, botID)
	if err != nil {
		return fmt.Errorf("erro ao carregar design: %w", err)
	}

	// 2. Verificar se os nós existem
	if !bm.nodeExists(design, fromNodeID) {
		return fmt.Errorf("nó origem %s não encontrado", fromNodeID)
	}
	if !bm.nodeExists(design, toNodeID) {
		return fmt.Errorf("nó destino %s não encontrado", toNodeID)
	}

	// 3. Criar nova edge
	newEdge := flow.Edge{
		From:  flow.ID(fromNodeID),
		To:    flow.ID(toNodeID),
		Label: label,
	}

	// 4. Adicionar edge usando DesignService
	updatedDesign, err := bm.designService.AddEdge(ctx, design, newEdge, adapterName)
	if err != nil {
		return fmt.Errorf("erro ao adicionar edge: %w", err)
	}

	// 5. Salvar design atualizado
	_, err = bm.storeService.Save(ctx, botID, updatedDesign)
	if err != nil {
		return fmt.Errorf("erro ao salvar design atualizado: %w", err)
	}

	fmt.Printf("✅ Conexão criada com sucesso: %s -> %s\n", fromNodeID, toNodeID)
	return nil
}

// AddConfirmNode demonstra como adicionar um nó de confirmação
func (bm *BotManager) AddConfirmNode(ctx context.Context, botID string, nodeID string, question string, yesLabel string, noLabel string, adapterName string) error {
	fmt.Printf("=== Adicionando nó de confirmação: %s ===\n", nodeID)

	// 1. Carregar design atual
	design, err := bm.storeService.Load(ctx, botID)
	if err != nil {
		return fmt.Errorf("erro ao carregar design: %w", err)
	}

	// 2. Criar nó de confirmação
	confirmNode := flow.Node{
		ID:    flow.ID(nodeID),
		Kind:  "confirm",
		Title: "Confirmação",
		Props: map[string]any{
			"text":      question,
			"yes_label": yesLabel,
			"no_label":  noLabel,
		},
		Outputs: []string{"yes", "no"},
		X:       float64Ptr(500),
		Y:       float64Ptr(300),
	}

	// 3. Adicionar nó
	updatedDesign, err := bm.designService.AddNode(ctx, design, confirmNode, adapterName)
	if err != nil {
		return fmt.Errorf("erro ao adicionar nó de confirmação: %w", err)
	}

	// 4. Salvar design atualizado
	_, err = bm.storeService.Save(ctx, botID, updatedDesign)
	if err != nil {
		return fmt.Errorf("erro ao salvar design atualizado: %w", err)
	}

	fmt.Printf("✅ Nó de confirmação %s adicionado com sucesso\n", nodeID)
	return nil
}

// ValidateBot valida um bot completo
func (bm *BotManager) ValidateBot(ctx context.Context, botID string, adapterName string) (*ValidationResult, error) {
	fmt.Printf("=== Validando bot: %s ===\n", botID)

	// 1. Carregar design
	design, err := bm.storeService.Load(ctx, botID)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar design: %w", err)
	}

	// 2. Validar usando ValidationService
	result, err := bm.designService.Validate(ctx, design, adapterName)
	if err != nil {
		return nil, fmt.Errorf("erro na validação: %w", err)
	}

	// 3. Exibir resultado
	if result.Valid {
		fmt.Printf("✅ Bot %s é válido! (%v)\n", botID, result.Duration)
	} else {
		fmt.Printf("❌ Bot %s tem problemas:\n", botID)
		for _, issue := range result.Issues {
			fmt.Printf("  - [%s] %s: %s\n", issue.Severity, issue.Code, issue.Msg)
		}
	}

	return result, nil
}

// ListBots lista todos os bots no store
func (bm *BotManager) ListBots(ctx context.Context) error {
	fmt.Printf("=== Listando bots ===\n")

	versions, err := bm.storeService.List(ctx)
	if err != nil {
		return fmt.Errorf("erro ao listar bots: %w", err)
	}

	if len(versions) == 0 {
		fmt.Println("Nenhum bot encontrado")
		return nil
	}

	for _, version := range versions {
		checksum := version.Checksum
		if len(checksum) > 8 {
			checksum = checksum[:8]
		}
		fmt.Printf("- Bot: %s (versão: %s, checksum: %s)\n",
			version.BotID, version.ID, checksum)
	}

	return nil
}

// GetBotInfo obtém informações detalhadas de um bot
func (bm *BotManager) GetBotInfo(ctx context.Context, botID string) error {
	fmt.Printf("=== Informações do bot: %s ===\n", botID)

	// 1. Verificar se existe
	exists, err := bm.storeService.Exists(ctx, botID)
	if err != nil {
		return fmt.Errorf("erro ao verificar existência: %w", err)
	}

	if !exists {
		fmt.Printf("❌ Bot %s não encontrado\n", botID)
		return nil
	}

	// 2. Carregar design
	design, err := bm.storeService.Load(ctx, botID)
	if err != nil {
		return fmt.Errorf("erro ao carregar design: %w", err)
	}

	// 3. Obter informações de versão
	versionInfo, err := bm.storeService.GetVersion(ctx, botID)
	if err != nil {
		return fmt.Errorf("erro ao obter versão: %w", err)
	}

	// 4. Exibir informações
	fmt.Printf("Bot ID: %s\n", design.Bot.ID)
	fmt.Printf("Canais: %v\n", design.Bot.Channels)
	fmt.Printf("Versão: %s (%s)\n", design.Version.ID, design.Version.Status)
	fmt.Printf("Nós: %d\n", len(design.Graph.Nodes))
	fmt.Printf("Conexões: %d\n", len(design.Graph.Edges))
	fmt.Printf("Entradas: %d\n", len(design.Entries))
	fmt.Printf("Checksum: %s\n", versionInfo.Checksum)
	fmt.Printf("Criado em: %s\n", versionInfo.CreatedAt.Format("2006-01-02 15:04:05"))

	return nil
}

// UpdateNodeMessage atualiza a mensagem de um nó
func (bm *BotManager) UpdateNodeMessage(ctx context.Context, botID string, nodeID string, newMessage string, adapterName string) error {
	fmt.Printf("=== Atualizando mensagem do nó: %s ===\n", nodeID)

	// 1. Carregar design atual
	design, err := bm.storeService.Load(ctx, botID)
	if err != nil {
		return fmt.Errorf("erro ao carregar design: %w", err)
	}

	// 2. Encontrar o nó
	var targetNode *flow.Node
	for i, node := range design.Graph.Nodes {
		if string(node.ID) == nodeID {
			targetNode = &design.Graph.Nodes[i]
			break
		}
	}

	if targetNode == nil {
		return fmt.Errorf("nó %s não encontrado", nodeID)
	}

	// 3. Atualizar propriedades do nó
	updatedNode := *targetNode
	if updatedNode.Props == nil {
		updatedNode.Props = make(map[string]any)
	}
	updatedNode.Props["text"] = newMessage

	// 4. Usar DesignService para atualizar
	updatedDesign, err := bm.designService.UpdateNode(ctx, design, nodeID, updatedNode, adapterName)
	if err != nil {
		return fmt.Errorf("erro ao atualizar nó: %w", err)
	}

	// 5. Salvar no store
	_, err = bm.storeService.Save(ctx, botID, updatedDesign)
	if err != nil {
		return fmt.Errorf("erro ao salvar design atualizado: %w", err)
	}

	fmt.Printf("✅ Mensagem do nó %s atualizada com sucesso\n", nodeID)
	return nil
}

// DeleteBot remove um bot completamente
func (bm *BotManager) DeleteBot(ctx context.Context, botID string) error {
	fmt.Printf("=== Removendo bot: %s ===\n", botID)

	err := bm.storeService.Delete(ctx, botID)
	if err != nil {
		return fmt.Errorf("erro ao remover bot: %w", err)
	}

	fmt.Printf("✅ Bot %s removido com sucesso\n", botID)
	return nil
}

// Exemplo de uso completo
func (bm *BotManager) DemoCompleteFlow(ctx context.Context) error {
	fmt.Println("🚀 DEMONSTRAÇÃO COMPLETA DOS SERVIÇOS")
	fmt.Println("=====================================")

	botID := "demo-bot-001"
	adapterName := "whatsapp"

	// 1. Criar bot
	if err := bm.CreateBot(ctx, botID, adapterName); err != nil {
		return err
	}

	// 2. Adicionar nós
	if err := bm.AddMessageNode(ctx, botID, "menu", "Escolha uma opção:\n1. Suporte\n2. Vendas\n3. Informações", adapterName); err != nil {
		return err
	}

	if err := bm.AddConfirmNode(ctx, botID, "confirm_support", "Deseja ser transferido para o suporte?", "Sim", "Não", adapterName); err != nil {
		return err
	}

	if err := bm.AddMessageNode(ctx, botID, "support_msg", "Você será transferido para nosso suporte em instantes...", adapterName); err != nil {
		return err
	}

	// 3. Conectar nós
	if err := bm.ConnectNodes(ctx, botID, "start", "menu", "início", adapterName); err != nil {
		return err
	}

	if err := bm.ConnectNodes(ctx, botID, "menu", "confirm_support", "suporte", adapterName); err != nil {
		return err
	}

	if err := bm.ConnectNodes(ctx, botID, "confirm_support", "support_msg", "sim", adapterName); err != nil {
		return err
	}

	// 4. Validar bot
	result, err := bm.ValidateBot(ctx, botID, adapterName)
	if err != nil {
		return err
	}

	// 5. Mostrar informações
	if err := bm.GetBotInfo(ctx, botID); err != nil {
		return err
	}

	// 6. Atualizar uma mensagem
	if err := bm.UpdateNodeMessage(ctx, botID, "start", "Olá! Bem-vindo ao nosso novo sistema de atendimento! Como posso ajudar?", adapterName); err != nil {
		return err
	}

	// 7. Validar novamente para confirmar
	_, err = bm.ValidateBot(ctx, botID, adapterName)
	if err != nil {
		return err
	}

	// 8. Listar todos os bots
	if err := bm.ListBots(ctx); err != nil {
		return err
	}

	fmt.Println("\n✅ DEMONSTRAÇÃO CONCLUÍDA COM SUCESSO!")

	// Mostrar plano se válido
	if result != nil && result.Valid && result.Plan != nil {
		fmt.Printf("\n📋 Plano de execução gerado (checksum: %s)\n", result.Plan.DesignChecksum)
		fmt.Printf("   - Adapter: %s\n", result.Plan.Adapter)
		fmt.Printf("   - Rotas: %d\n", len(result.Plan.Routes))
	}

	return nil
}

// Funções auxiliares
func (bm *BotManager) nodeExists(design io.DesignDoc, nodeID string) bool {
	for _, node := range design.Graph.Nodes {
		if string(node.ID) == nodeID {
			return true
		}
	}
	return false
}

func float64Ptr(f float64) *float64 {
	return &f
}

// ExampleMain demonstra como usar o BotManager
func ExampleMain() {
	ctx := context.Background()

	// Criar gerenciador de bots
	botManager := NewBotManager()

	// Executar demonstração completa
	if err := botManager.DemoCompleteFlow(ctx); err != nil {
		log.Fatalf("Erro na demonstração: %v", err)
	}
}
