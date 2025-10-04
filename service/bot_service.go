// Package service - BotService unificado para manipulação de bots
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/AgendoCerto/lib-bot/flow"
	"github.com/AgendoCerto/lib-bot/io"
)

// BotService é o serviço unificado que manipula designs diretamente (sem patches)
// Suporta WhatsApp completo (header/footer/ptt) + funções auxiliares (GetBot, Clone, Import)
type BotService struct {
	validation *ValidationService
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

// ValidationIssue representa um problema de validação
type ValidationIssue struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Path     string `json:"path"`
}

// NewBotService cria instância simplificada
func NewBotService() *BotService {
	return &BotService{
		validation: NewValidationService(),
		store:      NewStoreService(),
	}
}

// CreateBot cria um novo bot
func (sbs *BotService) CreateBot(ctx context.Context, botID, name, adapterName string) error {
	if adapterName == "" {
		adapterName = "whatsapp"
	}

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
		Variables: io.Variables{
			Context: []string{},            // Keys temporárias (sessão)
			State:   []string{"user_name"}, // Keys permanentes (usuário)
			Global:  map[string]any{},      // Valores compartilhados (bot)
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

	// Validar
	result, err := sbs.validation.ValidateDesign(ctx, design, adapterName)
	if err != nil {
		return fmt.Errorf("erro na validação: %w", err)
	}

	if !result.Valid {
		return fmt.Errorf("design básico inválido: %d issues", len(result.Issues))
	}

	// Salvar
	_, err = sbs.store.Save(ctx, botID, design)
	return err
}

// AddMessageNode adiciona nó de mensagem
func (sbs *BotService) AddMessageNode(ctx context.Context, botID, nodeID, message string) error {
	design, err := sbs.store.Load(ctx, botID)
	if err != nil {
		return err
	}

	// Verificar se nó já existe
	for _, node := range design.Graph.Nodes {
		if string(node.ID) == nodeID {
			return fmt.Errorf("nó %s já existe", nodeID)
		}
	}

	// Adicionar nó
	newNode := flow.Node{
		ID:      flow.ID(nodeID),
		Kind:    "message",
		Title:   "Mensagem",
		Props:   map[string]any{"text": message},
		Outputs: []string{"complete"},
		X:       &[]float64{200 + float64(len(design.Graph.Nodes)*150)}[0],
		Y:       &[]float64{200}[0],
	}

	design.Graph.Nodes = append(design.Graph.Nodes, newNode)

	// Validar e salvar
	result, err := sbs.validation.ValidateDesign(ctx, design, "whatsapp")
	if err != nil {
		return err
	}

	if !result.Valid {
		return fmt.Errorf("design inválido: %d issues", len(result.Issues))
	}

	_, err = sbs.store.Save(ctx, botID, design)
	return err
}

// AddButtonsNode adiciona nó com botões (máx 3 botões WhatsApp)
func (sbs *BotService) AddButtonsNode(ctx context.Context, botID, nodeID, text string, buttons []map[string]string, header, footer string) error {
	design, err := sbs.store.Load(ctx, botID)
	if err != nil {
		return err
	}

	// Verificar se nó já existe
	for _, node := range design.Graph.Nodes {
		if string(node.ID) == nodeID {
			return fmt.Errorf("nó %s já existe", nodeID)
		}
	}

	// Preparar props do componente buttons
	props := map[string]any{
		"text":    text,
		"buttons": buttons,
	}

	// Adicionar header e footer se fornecidos (WhatsApp)
	if header != "" {
		props["header"] = header
	}
	if footer != "" {
		props["footer"] = footer
	}

	// Preparar outputs baseado nos botões
	outputs := []string{"selected"} // Spec v2.2: output único

	// Adicionar nó
	newNode := flow.Node{
		ID:      flow.ID(nodeID),
		Kind:    "buttons",
		Title:   "Botões",
		Props:   props,
		Outputs: outputs,
		X:       &[]float64{200 + float64(len(design.Graph.Nodes)*150)}[0],
		Y:       &[]float64{200}[0],
	}

	design.Graph.Nodes = append(design.Graph.Nodes, newNode)

	// Validar e salvar (validações WhatsApp serão aplicadas)
	result, err := sbs.validation.ValidateDesign(ctx, design, "whatsapp")
	if err != nil {
		return err
	}

	if !result.Valid {
		fmt.Printf("Issues encontrados:\n")
		for i, issue := range result.Issues {
			fmt.Printf("%d. [%s] %s: %s (Path: %s)\n",
				i+1, issue.Severity, issue.Code, issue.Msg, issue.Path)
		}
		return fmt.Errorf("design inválido: %d issues", len(result.Issues))
	}

	_, err = sbs.store.Save(ctx, botID, design)
	return err
}

// AddListPickerNode adiciona nó com lista de opções (máx 10 seções WhatsApp)
func (sbs *BotService) AddListPickerNode(ctx context.Context, botID, nodeID, text, buttonText string, sections []map[string]any, header, footer string) error {
	design, err := sbs.store.Load(ctx, botID)
	if err != nil {
		return err
	}

	// Verificar se nó já existe
	for _, node := range design.Graph.Nodes {
		if string(node.ID) == nodeID {
			return fmt.Errorf("nó %s já existe", nodeID)
		}
	}

	// Preparar props do componente listpicker
	props := map[string]any{
		"text":        text,
		"button_text": buttonText,
		"sections":    sections,
	}

	// Adicionar header e footer se fornecidos (WhatsApp)
	if header != "" {
		props["header"] = header
	}
	if footer != "" {
		props["footer"] = footer
	}

	// Preparar outputs
	outputs := []string{"selected"} // Spec v2.2: output único

	// Adicionar nó
	newNode := flow.Node{
		ID:      flow.ID(nodeID),
		Kind:    "listpicker",
		Title:   "Lista de Seleção",
		Props:   props,
		Outputs: outputs,
		X:       &[]float64{200 + float64(len(design.Graph.Nodes)*150)}[0],
		Y:       &[]float64{200}[0],
	}

	design.Graph.Nodes = append(design.Graph.Nodes, newNode)

	// Validar e salvar (validações WhatsApp serão aplicadas)
	result, err := sbs.validation.ValidateDesign(ctx, design, "whatsapp")
	if err != nil {
		return err
	}

	if !result.Valid {
		fmt.Printf("Issues encontrados:\n")
		for i, issue := range result.Issues {
			fmt.Printf("%d. [%s] %s: %s (Path: %s)\n",
				i+1, issue.Severity, issue.Code, issue.Msg, issue.Path)
		}
		return fmt.Errorf("design inválido: %d issues", len(result.Issues))
	}

	_, err = sbs.store.Save(ctx, botID, design)
	return err
}

// AddMediaNode adiciona nó de mídia (image, video, audio, document)
func (sbs *BotService) AddMediaNode(ctx context.Context, botID, nodeID, mediaType, url, caption, filename string, ptt bool) error {
	design, err := sbs.store.Load(ctx, botID)
	if err != nil {
		return err
	}

	// Verificar se nó já existe
	for _, node := range design.Graph.Nodes {
		if string(node.ID) == nodeID {
			return fmt.Errorf("nó %s já existe", nodeID)
		}
	}

	// Preparar props do componente media
	props := map[string]any{
		"type": mediaType,
		"url":  url,
	}

	if caption != "" {
		props["caption"] = caption
	}

	if filename != "" {
		props["filename"] = filename
	}

	// Push-to-talk para áudio (WhatsApp)
	if mediaType == "audio" {
		props["ptt"] = ptt
	}

	// Adicionar nó
	newNode := flow.Node{
		ID:      flow.ID(nodeID),
		Kind:    "media",
		Title:   fmt.Sprintf("Mídia (%s)", mediaType),
		Props:   props,
		Outputs: []string{"sent"}, // Media usa "sent" como output
		X:       &[]float64{200 + float64(len(design.Graph.Nodes)*150)}[0],
		Y:       &[]float64{200}[0],
	}

	design.Graph.Nodes = append(design.Graph.Nodes, newNode)

	// Validar e salvar (validações WhatsApp serão aplicadas)
	result, err := sbs.validation.ValidateDesign(ctx, design, "whatsapp")
	if err != nil {
		return err
	}

	if !result.Valid {
		fmt.Printf("Issues encontrados:\n")
		for i, issue := range result.Issues {
			fmt.Printf("%d. [%s] %s: %s (Path: %s)\n",
				i+1, issue.Severity, issue.Code, issue.Msg, issue.Path)
		}
		return fmt.Errorf("design inválido: %d issues", len(result.Issues))
	}

	_, err = sbs.store.Save(ctx, botID, design)
	return err
}

// AddConfirmNode adiciona nó de confirmação
func (sbs *BotService) AddConfirmNode(ctx context.Context, botID, nodeID, question, yesLabel, noLabel string) error {
	design, err := sbs.store.Load(ctx, botID)
	if err != nil {
		return err
	}

	// Verificar se nó já existe
	for _, node := range design.Graph.Nodes {
		if string(node.ID) == nodeID {
			return fmt.Errorf("nó %s já existe", nodeID)
		}
	}

	// Adicionar nó de confirmação
	newNode := flow.Node{
		ID:    flow.ID(nodeID),
		Kind:  "confirm",
		Title: "Confirmação",
		Props: map[string]any{
			"title": question,
			"yes":   yesLabel,
			"no":    noLabel,
		},
		Outputs: []string{"confirmed", "cancelled", "timeout"},
		X:       &[]float64{200 + float64(len(design.Graph.Nodes)*150)}[0],
		Y:       &[]float64{200}[0],
	}

	design.Graph.Nodes = append(design.Graph.Nodes, newNode)

	// Validar e salvar
	result, err := sbs.validation.ValidateDesign(ctx, design, "whatsapp")
	if err != nil {
		return err
	}

	if !result.Valid {
		fmt.Printf("Issues encontrados:\n")
		for i, issue := range result.Issues {
			fmt.Printf("%d. [%s] %s: %s (Path: %s)\n",
				i+1, issue.Severity, issue.Code, issue.Msg, issue.Path)
		}
		return fmt.Errorf("design inválido: %d issues", len(result.Issues))
	}

	_, err = sbs.store.Save(ctx, botID, design)
	return err
}

// ConnectNodes conecta dois nós
func (sbs *BotService) ConnectNodes(ctx context.Context, botID, fromNodeID, toNodeID, label string) error {
	design, err := sbs.store.Load(ctx, botID)
	if err != nil {
		return err
	}

	// Verificar se nós existem
	fromExists := false
	toExists := false
	for _, node := range design.Graph.Nodes {
		if string(node.ID) == fromNodeID {
			fromExists = true
		}
		if string(node.ID) == toNodeID {
			toExists = true
		}
	}

	if !fromExists {
		return fmt.Errorf("nó origem %s não encontrado", fromNodeID)
	}
	if !toExists {
		return fmt.Errorf("nó destino %s não encontrado", toNodeID)
	}

	// Adicionar edge
	newEdge := flow.Edge{
		From:  flow.ID(fromNodeID),
		To:    flow.ID(toNodeID),
		Label: label,
	}

	design.Graph.Edges = append(design.Graph.Edges, newEdge)

	// Validar e salvar
	result, err := sbs.validation.ValidateDesign(ctx, design, "whatsapp")
	if err != nil {
		return err
	}

	if !result.Valid {
		return fmt.Errorf("design inválido: %d issues", len(result.Issues))
	}

	_, err = sbs.store.Save(ctx, botID, design)
	return err
}

// ValidateBot valida um bot
func (sbs *BotService) ValidateBot(ctx context.Context, botID, adapterName string) (*ValidationResult, error) {
	design, err := sbs.store.Load(ctx, botID)
	if err != nil {
		return nil, err
	}

	return sbs.validation.ValidateDesign(ctx, design, adapterName)
}

// GetBotInfo obtém informações básicas
func (sbs *BotService) GetBotInfo(ctx context.Context, botID string) (map[string]interface{}, error) {
	design, err := sbs.store.Load(ctx, botID)
	if err != nil {
		return nil, err
	}

	result, err := sbs.validation.ValidateDesign(ctx, design, "whatsapp")
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":       design.Bot.ID,
		"nodes":    len(design.Graph.Nodes),
		"edges":    len(design.Graph.Edges),
		"valid":    result.Valid,
		"issues":   len(result.Issues),
		"channels": design.Bot.Channels,
		"version":  design.Version.ID,
	}, nil
}

// ListBots lista todos os bots
func (sbs *BotService) ListBots(ctx context.Context) error {
	versions, err := sbs.store.List(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("📋 Total de bots: %d\n", len(versions))
	for _, version := range versions {
		info, err := sbs.GetBotInfo(ctx, version.BotID)
		if err != nil {
			fmt.Printf("   ❌ %s (erro ao carregar)\n", version.BotID)
			continue
		}

		status := "❌"
		if info["valid"].(bool) {
			status = "✅"
		}

		fmt.Printf("   %s %s - %d nós, %d edges\n",
			status, version.BotID, info["nodes"], info["edges"])
	}

	return nil
}

// DemoSimpleBot demonstração completa
func (sbs *BotService) DemoSimpleBot(ctx context.Context) error {
	fmt.Println("🚀 DEMONSTRAÇÃO SIMPLEBOT SERVICE")
	fmt.Println("=================================")

	botID := "simple-demo-bot"

	// 1. Criar bot
	fmt.Println("📋 1. Criando bot...")
	err := sbs.CreateBot(ctx, botID, "Bot Demo Simples", "whatsapp")
	if err != nil {
		return fmt.Errorf("erro ao criar bot: %w", err)
	}
	fmt.Printf("✅ Bot criado: %s\n", botID)

	// 2. Adicionar nós
	fmt.Println("📋 2. Adicionando nós...")

	err = sbs.AddMessageNode(ctx, botID, "menu", "Escolha uma opção:\n1️⃣ Suporte\n2️⃣ Vendas")
	if err != nil {
		return fmt.Errorf("erro ao adicionar menu: %w", err)
	}
	fmt.Println("✅ Menu adicionado")

	err = sbs.AddConfirmNode(ctx, botID, "confirm_support", "Deseja falar com suporte?", "Sim", "Não")
	if err != nil {
		return fmt.Errorf("erro ao adicionar confirmação: %w", err)
	}
	fmt.Println("✅ Confirmação adicionada")

	err = sbs.AddMessageNode(ctx, botID, "thank_you", "Obrigado pelo contato! 😊")
	if err != nil {
		return fmt.Errorf("erro ao adicionar agradecimento: %w", err)
	}
	fmt.Println("✅ Agradecimento adicionado")

	// 3. Conectar nós
	fmt.Println("📋 3. Conectando nós...")

	err = sbs.ConnectNodes(ctx, botID, "start", "menu", "início")
	if err != nil {
		return fmt.Errorf("erro ao conectar start->menu: %w", err)
	}
	fmt.Println("✅ start -> menu")

	err = sbs.ConnectNodes(ctx, botID, "menu", "confirm_support", "1")
	if err != nil {
		return fmt.Errorf("erro ao conectar menu->confirm: %w", err)
	}
	fmt.Println("✅ menu -> confirm_support")

	err = sbs.ConnectNodes(ctx, botID, "confirm_support", "thank_you", "confirmed")
	if err != nil {
		return fmt.Errorf("erro ao conectar confirm->thank_you: %w", err)
	}
	fmt.Println("✅ confirm_support -> thank_you")

	// 4. Validar
	fmt.Println("📋 4. Validando bot...")
	result, err := sbs.ValidateBot(ctx, botID, "whatsapp")
	if err != nil {
		return fmt.Errorf("erro na validação: %w", err)
	}

	if result.Valid {
		fmt.Printf("✅ Bot válido! (%v)\n", result.Duration)
		if result.Plan != nil {
			fmt.Printf("📋 Plano de execução gerado (checksum: %s)\n", result.Plan.DesignChecksum)
		}
	} else {
		fmt.Printf("⚠️ Bot tem %d issues:\n", len(result.Issues))
		for _, issue := range result.Issues {
			fmt.Printf("   - [%s] %s\n", issue.Severity, issue.Msg)
		}
	}

	// 5. Informações finais
	fmt.Println("📋 5. Informações finais...")
	info, err := sbs.GetBotInfo(ctx, botID)
	if err != nil {
		return fmt.Errorf("erro ao obter info: %w", err)
	}

	fmt.Printf("📊 Estatísticas:\n")
	fmt.Printf("   - ID: %s\n", info["id"])
	fmt.Printf("   - Nós: %d\n", info["nodes"])
	fmt.Printf("   - Conexões: %d\n", info["edges"])
	fmt.Printf("   - Válido: %t\n", info["valid"])

	fmt.Println("\n🎉 DEMONSTRAÇÃO CONCLUÍDA COM SUCESSO!")
	return nil
}
