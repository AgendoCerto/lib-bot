// Package service - Exemplo completo de uso do BotService unificado
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// ExampleBotService demonstra todas as funcionalidades do BotService
type ExampleBotService struct {
	botService *BotService
}

// NewExampleBotService cria uma nova instância do exemplo
func NewExampleBotService() *ExampleBotService {
	return &ExampleBotService{
		botService: NewBotService(),
	}
}

// DemoCompleteWorkflow demonstra um fluxo completo de criação de bot
func (ex *ExampleBotService) DemoCompleteWorkflow(ctx context.Context) error {
	fmt.Println("🚀 DEMONSTRAÇÃO COMPLETA DO BOTSERVICE UNIFICADO")
	fmt.Println("===============================================")

	botID := "atendimento-unified"

	// 1. CRIAR BOT
	fmt.Println("\n📋 1. Criando bot...")
	botInfo, err := ex.botService.CreateBot(ctx, CreateBotRequest{
		ID:          botID,
		Name:        "Bot de Atendimento",
		Description: "Bot para atendimento ao cliente",
		Channels:    []string{"whatsapp", "telegram"},
		AdapterName: "whatsapp",
	})
	if err != nil {
		return fmt.Errorf("erro ao criar bot: %w", err)
	}
	fmt.Printf("✅ Bot criado: %s (versão: %s)\n", botInfo.ID, botInfo.Version)

	// 2. ADICIONAR MENU PRINCIPAL
	fmt.Println("\n📋 2. Adicionando menu principal...")
	menuNode, err := ex.botService.AddMessageNode(ctx, botID, "main_menu",
		"🤖 Olá! Como posso ajudar?\n\n1️⃣ Suporte Técnico\n2️⃣ Vendas\n3️⃣ Informações\n4️⃣ Falar com Humano",
		&Position{X: 300, Y: 200})
	if err != nil {
		return fmt.Errorf("erro ao adicionar menu: %w", err)
	}
	fmt.Printf("✅ Menu adicionado: %s\n", menuNode.ID)

	// 3. ADICIONAR NÓIS DE SUPORTE
	fmt.Println("\n📋 3. Adicionando seção de suporte...")

	// Nó de confirmação para suporte
	supportConfirm, err := ex.botService.AddConfirmNode(ctx, botID, "support_confirm",
		"Você precisa de suporte técnico?", "Sim, preciso", "Não, obrigado",
		&Position{X: 500, Y: 300})
	if err != nil {
		return fmt.Errorf("erro ao adicionar confirmação suporte: %w", err)
	}
	fmt.Printf("✅ Confirmação de suporte: %s\n", supportConfirm.ID)

	// Nó de coleta de informações
	supportInput, err := ex.botService.AddInputNode(ctx, botID, "support_details",
		"Descreva brevemente o problema:", "Ex: Não consigo fazer login",
		&Position{X: 700, Y: 250})
	if err != nil {
		return fmt.Errorf("erro ao adicionar input suporte: %w", err)
	}
	fmt.Printf("✅ Input de suporte: %s\n", supportInput.ID)

	// 4. ADICIONAR SEÇÃO DE VENDAS
	fmt.Println("\n📋 4. Adicionando seção de vendas...")

	salesPicker, err := ex.botService.AddListPickerNode(ctx, botID, "sales_products",
		"Qual produto te interessa?",
		[]string{"Plano Básico", "Plano Premium", "Plano Enterprise", "Consultoria"},
		&Position{X: 500, Y: 400})
	if err != nil {
		return fmt.Errorf("erro ao adicionar lista vendas: %w", err)
	}
	fmt.Printf("✅ Seleção de produtos: %s\n", salesPicker.ID)

	// 5. ADICIONAR DELAYS E FINALIZAÇÃO
	fmt.Println("\n📋 5. Adicionando delays e mensagens finais...")

	delay, err := ex.botService.AddDelayNode(ctx, botID, "processing_delay", 2,
		&Position{X: 700, Y: 450})
	if err != nil {
		return fmt.Errorf("erro ao adicionar delay: %w", err)
	}
	fmt.Printf("✅ Delay: %s\n", delay.ID)

	finalMessage, err := ex.botService.AddMessageNode(ctx, botID, "final_message",
		"Obrigado pelo contato! Em breve entraremos em contato. 😊",
		&Position{X: 900, Y: 350})
	if err != nil {
		return fmt.Errorf("erro ao adicionar mensagem final: %w", err)
	}
	fmt.Printf("✅ Mensagem final: %s\n", finalMessage.ID)

	// 6. CONECTAR NÓIS
	fmt.Println("\n📋 6. Conectando fluxo...")

	connections := []EdgeInfo{
		{From: "start", To: "main_menu", Label: "início"},
		{From: "main_menu", To: "support_confirm", Label: "1"},
		{From: "main_menu", To: "sales_products", Label: "2"},
		{From: "support_confirm", To: "support_details", Label: "confirmed"},
		{From: "support_confirm", To: "final_message", Label: "cancelled"},
		{From: "support_details", To: "processing_delay", Label: "complete"},
		{From: "sales_products", To: "processing_delay", Label: "complete"},
		{From: "processing_delay", To: "final_message", Label: "complete"},
	}

	for _, conn := range connections {
		_, err := ex.botService.ConnectNodes(ctx, botID, conn)
		if err != nil {
			return fmt.Errorf("erro ao conectar %s -> %s: %w", conn.From, conn.To, err)
		}
		fmt.Printf("✅ Conectado: %s -> %s\n", conn.From, conn.To)
	}

	// 7. VALIDAR BOT
	fmt.Println("\n📋 7. Validando bot...")
	validationResult, err := ex.botService.ValidateBot(ctx, botID, "whatsapp")
	if err != nil {
		return fmt.Errorf("erro na validação: %w", err)
	}

	if validationResult.Valid {
		fmt.Printf("✅ Bot válido! (%v)\n", validationResult.Duration)
		if validationResult.Plan != nil {
			fmt.Printf("   📋 Plano de execução gerado (checksum: %s)\n", validationResult.Plan.DesignChecksum)
		}
	} else {
		fmt.Printf("⚠️ Bot tem %d issues:\n", len(validationResult.Issues))
		for _, issue := range validationResult.Issues {
			fmt.Printf("   - [%s] %s\n", issue.Severity, issue.Msg)
		}
	}

	// 8. OBTER INFORMAÇÕES FINAIS
	fmt.Println("\n📋 8. Informações finais do bot...")
	finalBotInfo, err := ex.botService.GetBot(ctx, botID)
	if err != nil {
		return fmt.Errorf("erro ao obter info final: %w", err)
	}

	fmt.Printf("📊 Estatísticas finais:\n")
	fmt.Printf("   - ID: %s\n", finalBotInfo.ID)
	fmt.Printf("   - Nós: %d\n", finalBotInfo.NodesCount)
	fmt.Printf("   - Conexões: %d\n", finalBotInfo.EdgesCount)
	fmt.Printf("   - Válido: %t\n", finalBotInfo.Valid)
	fmt.Printf("   - Checksum: %s\n", finalBotInfo.Checksum)

	// 9. LISTAR TODOS OS NÓIS E CONEXÕES
	fmt.Println("\n📋 9. Estrutura completa do bot...")
	nodes, err := ex.botService.GetNodes(ctx, botID)
	if err != nil {
		return fmt.Errorf("erro ao obter nós: %w", err)
	}

	fmt.Printf("🔵 Nós (%d):\n", len(nodes))
	for _, node := range nodes {
		pos := "sem posição"
		if node.Position != nil {
			pos = fmt.Sprintf("(%.0f, %.0f)", node.Position.X, node.Position.Y)
		}
		fmt.Printf("   - %s [%s]: %s %s\n", node.ID, node.Kind, node.Title, pos)
	}

	edges, err := ex.botService.GetEdges(ctx, botID)
	if err != nil {
		return fmt.Errorf("erro ao obter edges: %w", err)
	}

	fmt.Printf("🔗 Conexões (%d):\n", len(edges))
	for _, edge := range edges {
		fmt.Printf("   - %s -> %s (%s)\n", edge.From, edge.To, edge.Label)
	}

	fmt.Println("\n🎉 DEMONSTRAÇÃO CONCLUÍDA COM SUCESSO!")
	return nil
}

// DemoAdvancedFeatures demonstra funcionalidades avançadas
func (ex *ExampleBotService) DemoAdvancedFeatures(ctx context.Context) error {
	fmt.Println("\n🔬 DEMONSTRAÇÃO DE FUNCIONALIDADES AVANÇADAS")
	fmt.Println("===========================================")

	// 1. CRIAR BOT DE TESTE
	testBotID := "advanced-test-bot"
	_, err := ex.botService.CreateBot(ctx, CreateBotRequest{
		ID:          testBotID,
		Name:        "Bot de Teste Avançado",
		AdapterName: "whatsapp",
	})
	if err != nil {
		return fmt.Errorf("erro ao criar bot teste: %w", err)
	}

	// 2. ADICIONAR ALGUNS NÓIS
	_, err = ex.botService.AddMessageNode(ctx, testBotID, "test_msg", "Mensagem teste", nil)
	if err != nil {
		return fmt.Errorf("erro ao adicionar nó teste: %w", err)
	}

	// 3. ATUALIZAR NÓ
	fmt.Println("📝 Atualizando nó...")
	_, err = ex.botService.UpdateNode(ctx, testBotID, NodeInfo{
		ID:    "test_msg",
		Kind:  "message",
		Title: "Mensagem Atualizada",
		Props: map[string]interface{}{
			"text": "Esta mensagem foi atualizada! 🚀",
		},
		Outputs:  []string{"complete"},
		Position: &Position{X: 400, Y: 300},
	})
	if err != nil {
		return fmt.Errorf("erro ao atualizar nó: %w", err)
	}
	fmt.Println("✅ Nó atualizado com sucesso")

	// 4. CLONAR BOT
	fmt.Println("📋 Clonando bot...")
	clonedBot, err := ex.botService.CloneBot(ctx, testBotID, "cloned-bot")
	if err != nil {
		return fmt.Errorf("erro ao clonar bot: %w", err)
	}
	fmt.Printf("✅ Bot clonado: %s\n", clonedBot.ID)

	// 5. EXPORT/IMPORT
	fmt.Println("💾 Testando export/import...")
	design, err := ex.botService.GetBotDesign(ctx, testBotID)
	if err != nil {
		return fmt.Errorf("erro ao exportar design: %w", err)
	}

	// Salvar como JSON (simulação)
	designJSON, err := json.MarshalIndent(design, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar design: %w", err)
	}
	fmt.Printf("📄 Design exportado (%d bytes)\n", len(designJSON))

	// Importar para novo bot
	importedBot, err := ex.botService.ImportBotDesign(ctx, "imported-bot", *design, "whatsapp")
	if err != nil {
		return fmt.Errorf("erro ao importar design: %w", err)
	}
	fmt.Printf("✅ Design importado para bot: %s\n", importedBot.ID)

	// 6. LISTAR TODOS OS BOTS
	fmt.Println("📋 Listando todos os bots...")
	allBots, err := ex.botService.ListBots(ctx)
	if err != nil {
		return fmt.Errorf("erro ao listar bots: %w", err)
	}

	fmt.Printf("🤖 Total de bots: %d\n", len(allBots))
	for _, bot := range allBots {
		status := "❌"
		if bot.Valid {
			status = "✅"
		}
		fmt.Printf("   %s %s - %d nós, %d conexões\n", status, bot.ID, bot.NodesCount, bot.EdgesCount)
	}

	// 7. LIMPEZA (REMOVER BOTS DE TESTE)
	fmt.Println("🧹 Limpando bots de teste...")
	testBots := []string{testBotID, "cloned-bot", "imported-bot"}
	for _, botID := range testBots {
		err := ex.botService.DeleteBot(ctx, botID)
		if err != nil {
			fmt.Printf("⚠️ Erro ao remover %s: %v\n", botID, err)
		} else {
			fmt.Printf("🗑️ Bot %s removido\n", botID)
		}
	}

	fmt.Println("\n🎯 DEMONSTRAÇÃO AVANÇADA CONCLUÍDA!")
	return nil
}

// RunAllDemos executa todas as demonstrações
func (ex *ExampleBotService) RunAllDemos(ctx context.Context) error {
	fmt.Println("🌟 INICIANDO DEMONSTRAÇÃO COMPLETA DO BOTSERVICE")
	fmt.Println("===============================================")

	// Demo básico
	if err := ex.DemoCompleteWorkflow(ctx); err != nil {
		return fmt.Errorf("erro no demo básico: %w", err)
	}

	// Demo avançado
	if err := ex.DemoAdvancedFeatures(ctx); err != nil {
		return fmt.Errorf("erro no demo avançado: %w", err)
	}

	// Resultado final
	fmt.Println("\n🏆 TODAS AS DEMONSTRAÇÕES CONCLUÍDAS COM SUCESSO!")

	// Mostrar bots restantes
	finalBots, err := ex.botService.ListBots(ctx)
	if err != nil {
		return fmt.Errorf("erro ao listar bots finais: %w", err)
	}

	fmt.Printf("📊 Bots criados na demonstração: %d\n", len(finalBots))
	for _, bot := range finalBots {
		fmt.Printf("   🤖 %s (%d nós, válido: %t)\n", bot.ID, bot.NodesCount, bot.Valid)
	}

	return nil
}

// RunDemo função de conveniência para executar a demonstração
func RunDemo() error {
	ctx := context.Background()
	example := NewExampleBotService()
	return example.RunAllDemos(ctx)
}

// ExampleUsage exemplo rápido de uso
func ExampleUsage() {
	ctx := context.Background()
	botService := NewBotService()

	// Criar bot
	bot, err := botService.CreateBot(ctx, CreateBotRequest{
		ID:          "exemplo-rapido",
		Name:        "Exemplo Rápido",
		AdapterName: "whatsapp",
	})
	if err != nil {
		log.Fatalf("Erro: %v", err)
	}

	// Adicionar nó
	_, err = botService.AddMessageNode(ctx, bot.ID, "saudacao",
		"Olá! Bem-vindo!", &Position{X: 200, Y: 200})
	if err != nil {
		log.Fatalf("Erro: %v", err)
	}

	// Conectar
	_, err = botService.ConnectNodes(ctx, bot.ID, EdgeInfo{
		From:  "start",
		To:    "saudacao",
		Label: "início",
	})
	if err != nil {
		log.Fatalf("Erro: %v", err)
	}

	// Validar
	result, err := botService.ValidateBot(ctx, bot.ID, "whatsapp")
	if err != nil {
		log.Fatalf("Erro: %v", err)
	}

	fmt.Printf("Bot %s criado: válido = %t\n", bot.ID, result.Valid)
}
