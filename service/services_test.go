// Package service - Testes para verificar se os serviços funcionam corretamente
package service

import (
	"context"
	"fmt"
	"testing"
)

// TestBotManager_CreateBot testa a criação de um bot
func TestBotManager_CreateBot(t *testing.T) {
	ctx := context.Background()
	manager := NewBotManager()

	botID := "test-bot-001"
	adapterName := "whatsapp"

	err := manager.CreateBot(ctx, botID, adapterName)
	if err != nil {
		t.Fatalf("Erro ao criar bot: %v", err)
	}

	// Verificar se bot foi salvo
	exists, err := manager.storeService.Exists(ctx, botID)
	if err != nil {
		t.Fatalf("Erro ao verificar existência do bot: %v", err)
	}

	if !exists {
		t.Fatal("Bot deveria existir após criação")
	}
}

// TestBotManager_AddNode testa adição de nós
func TestBotManager_AddNode(t *testing.T) {
	ctx := context.Background()
	manager := NewBotManager()

	botID := "test-bot-002"
	adapterName := "whatsapp"

	// Criar bot primeiro
	err := manager.CreateBot(ctx, botID, adapterName)
	if err != nil {
		t.Fatalf("Erro ao criar bot: %v", err)
	}

	// Adicionar nó
	err = manager.AddMessageNode(ctx, botID, "test-node", "Mensagem de teste", adapterName)
	if err != nil {
		t.Fatalf("Erro ao adicionar nó: %v", err)
	}

	// Verificar se nó foi adicionado
	design, err := manager.storeService.Load(ctx, botID)
	if err != nil {
		t.Fatalf("Erro ao carregar design: %v", err)
	}

	// Deve ter 2 nós agora (start + test-node)
	if len(design.Graph.Nodes) != 2 {
		t.Fatalf("Esperado 2 nós, encontrado %d", len(design.Graph.Nodes))
	}

	// Verificar se o nó correto foi adicionado
	found := false
	for _, node := range design.Graph.Nodes {
		if string(node.ID) == "test-node" {
			found = true
			if node.Props["text"] != "Mensagem de teste" {
				t.Fatalf("Mensagem do nó incorreta: %v", node.Props["text"])
			}
			break
		}
	}

	if !found {
		t.Fatal("Nó test-node não encontrado")
	}
}

// TestBotManager_ConnectNodes testa conexão entre nós
func TestBotManager_ConnectNodes(t *testing.T) {
	ctx := context.Background()
	manager := NewBotManager()

	botID := "test-bot-003"
	adapterName := "whatsapp"

	// Criar bot
	err := manager.CreateBot(ctx, botID, adapterName)
	if err != nil {
		t.Fatalf("Erro ao criar bot: %v", err)
	}

	// Adicionar nó
	err = manager.AddMessageNode(ctx, botID, "second-node", "Segunda mensagem", adapterName)
	if err != nil {
		t.Fatalf("Erro ao adicionar nó: %v", err)
	}

	// Conectar nós
	err = manager.ConnectNodes(ctx, botID, "start", "second-node", "próximo", adapterName)
	if err != nil {
		t.Fatalf("Erro ao conectar nós: %v", err)
	}

	// Verificar se edge foi criada
	design, err := manager.storeService.Load(ctx, botID)
	if err != nil {
		t.Fatalf("Erro ao carregar design: %v", err)
	}

	if len(design.Graph.Edges) != 1 {
		t.Fatalf("Esperado 1 edge, encontrado %d", len(design.Graph.Edges))
	}

	edge := design.Graph.Edges[0]
	if string(edge.From) != "start" || string(edge.To) != "second-node" {
		t.Fatalf("Edge incorreta: %s -> %s", edge.From, edge.To)
	}
}

// TestBotManager_Validation testa validação de bots
func TestBotManager_Validation(t *testing.T) {
	ctx := context.Background()
	manager := NewBotManager()

	botID := "test-bot-004"
	adapterName := "whatsapp"

	// Criar bot
	err := manager.CreateBot(ctx, botID, adapterName)
	if err != nil {
		t.Fatalf("Erro ao criar bot: %v", err)
	}

	// Validar bot
	result, err := manager.ValidateBot(ctx, botID, adapterName)
	if err != nil {
		t.Fatalf("Erro ao validar bot: %v", err)
	}

	if !result.Valid {
		t.Fatalf("Bot deveria ser válido, mas tem %d issues", len(result.Issues))
	}
}

// TestBotManager_UpdateNode testa atualização de nós
func TestBotManager_UpdateNode(t *testing.T) {
	ctx := context.Background()
	manager := NewBotManager()

	botID := "test-bot-005"
	adapterName := "whatsapp"

	// Criar bot
	err := manager.CreateBot(ctx, botID, adapterName)
	if err != nil {
		t.Fatalf("Erro ao criar bot: %v", err)
	}

	newMessage := "Mensagem atualizada!"

	// Atualizar mensagem do nó start
	err = manager.UpdateNodeMessage(ctx, botID, "start", newMessage, adapterName)
	if err != nil {
		t.Fatalf("Erro ao atualizar nó: %v", err)
	}

	// Verificar se foi atualizado
	design, err := manager.storeService.Load(ctx, botID)
	if err != nil {
		t.Fatalf("Erro ao carregar design: %v", err)
	}

	for _, node := range design.Graph.Nodes {
		if string(node.ID) == "start" {
			if node.Props["text"] != newMessage {
				t.Fatalf("Mensagem não foi atualizada: %v", node.Props["text"])
			}
			return
		}
	}

	t.Fatal("Nó start não encontrado")
}

// TestBotManager_CompleteFlow testa o fluxo completo
func TestBotManager_CompleteFlow(t *testing.T) {
	ctx := context.Background()
	manager := NewBotManager()

	// Executar demonstração completa
	err := manager.DemoCompleteFlow(ctx)
	if err != nil {
		t.Fatalf("Erro na demonstração completa: %v", err)
	}

	// Verificar se bot demo foi criado
	exists, err := manager.storeService.Exists(ctx, "demo-bot-001")
	if err != nil {
		t.Fatalf("Erro ao verificar bot demo: %v", err)
	}

	if !exists {
		t.Fatal("Bot demo deveria existir após demonstração")
	}
}

// BenchmarkBotManager_CreateBot benchmark para criação de bots
func BenchmarkBotManager_CreateBot(b *testing.B) {
	ctx := context.Background()
	manager := NewBotManager()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		botID := fmt.Sprintf("bench-bot-%d", i)
		err := manager.CreateBot(ctx, botID, "whatsapp")
		if err != nil {
			b.Fatalf("Erro no benchmark: %v", err)
		}
	}
}

// TestBotManager_ListBots testa listagem de bots
func TestBotManager_ListBots(t *testing.T) {
	ctx := context.Background()
	manager := NewBotManager()

	// Criar alguns bots
	for i := 0; i < 3; i++ {
		botID := fmt.Sprintf("list-test-bot-%d", i)
		err := manager.CreateBot(ctx, botID, "whatsapp")
		if err != nil {
			t.Fatalf("Erro ao criar bot %d: %v", i, err)
		}
	}

	// Listar bots
	err := manager.ListBots(ctx)
	if err != nil {
		t.Fatalf("Erro ao listar bots: %v", err)
	}

	// Verificar se tem pelo menos 3 bots
	versions, err := manager.storeService.List(ctx)
	if err != nil {
		t.Fatalf("Erro ao obter lista: %v", err)
	}

	if len(versions) < 3 {
		t.Fatalf("Esperado pelo menos 3 bots, encontrado %d", len(versions))
	}
}
