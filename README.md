# Lib-Bot - Serviços Integrados para Criação de Bots

## Visão Geral

A Lib-Bot oferece dois serviços principais para criação e manipulação de bots conversacionais com validação automática e integração completa.

## Serviços Disponíveis

### 1. SimpleBotService (Recomendado)

Manipulação direta e rápida para a maioria dos casos de uso.

```go
import "github.com/AgendoCerto/lib-bot/service"

// Criar serviço
sbs := service.NewSimpleBotService()
ctx := context.Background()

// Criar bot
err := sbs.CreateBot(ctx, "meu-bot", "Bot de Atendimento", "whatsapp")

// Adicionar nós
err = sbs.AddMessageNode(ctx, "meu-bot", "welcome", "Olá! Como posso ajudar?")
err = sbs.AddConfirmNode(ctx, "meu-bot", "confirm", "Continuar?", "Sim", "Não")
err = sbs.AddMessageNode(ctx, "meu-bot", "thanks", "Obrigado!")

// Conectar nós
err = sbs.ConnectNodes(ctx, "meu-bot", "start", "welcome", "início")
err = sbs.ConnectNodes(ctx, "meu-bot", "welcome", "confirm", "next")
err = sbs.ConnectNodes(ctx, "meu-bot", "confirm", "thanks", "confirmed")

// Validar
result, err := sbs.ValidateBot(ctx, "meu-bot", "whatsapp")
if result.Valid {
    fmt.Println("Bot válido!")
}
```

### 2. BotService (Avançado)

Recursos avançados com posicionamento visual e patches RFC 6902.

```go
// Criar serviço
botService := service.NewBotService()

// Criar bot
bot, err := botService.CreateBot(ctx, service.CreateBotRequest{
    ID:          "bot-avancado",
    Name:        "Bot Avançado",
    AdapterName: "whatsapp",
})

// Adicionar nós com posicionamento
pos1 := &service.Position{X: 100, Y: 100}
menu, err := botService.AddMessageNode(ctx, "bot-avancado", "menu", 
    "Escolha uma opção", pos1)

pos2 := &service.Position{X: 300, Y: 100}
confirm, err := botService.AddConfirmNode(ctx, "bot-avancado", "confirm", 
    "Confirma?", "Sim", "Não", pos2)

// Conectar nós
edge, err := botService.ConnectNodes(ctx, "bot-avancado", service.EdgeInfo{
    From: "menu", To: "confirm", Label: "1",
})
```

## Componentes Suportados

### Mensagem
```go
// SimpleBotService
err = sbs.AddMessageNode(ctx, botID, "msg1", "Texto da mensagem")

// BotService
node, err := botService.AddMessageNode(ctx, botID, "msg1", "Texto", position)
```

### Confirmação
```go
// SimpleBotService
err = sbs.AddConfirmNode(ctx, botID, "confirm1", "Pergunta?", "Sim", "Não")

// BotService  
node, err := botService.AddConfirmNode(ctx, botID, "confirm1", "Pergunta?", "Sim", "Não", position)
```

### Lista de Opções
```go
// BotService
options := []string{"Opção 1", "Opção 2", "Opção 3"}
node, err := botService.AddListPickerNode(ctx, botID, "list1", "Escolha:", options, position)
```

### Entrada de Texto
```go
// BotService
node, err := botService.AddInputNode(ctx, botID, "input1", "Digite seu nome:", "placeholder", position)
```

### Delay
```go
// BotService
node, err := botService.AddDelayNode(ctx, botID, "delay1", 3000, position) // 3 segundos
```

## Validação

Ambos os serviços validam automaticamente:

```go
result, err := service.ValidateBot(ctx, botID, "whatsapp")
if result.Valid {
    fmt.Println("Bot válido!")
    if result.Plan != nil {
        fmt.Printf("Plano compilado: %s\n", result.Plan.DesignChecksum)
    }
} else {
    fmt.Printf("Problemas encontrados: %d issues\n", len(result.Issues))
    for _, issue := range result.Issues {
        fmt.Printf("- [%s] %s\n", issue.Severity, issue.Msg)
    }
}
```

## Exemplo Completo

```go
package main

import (
    "context"
    "fmt"
    "github.com/AgendoCerto/lib-bot/service"
)

func main() {
    ctx := context.Background()
    sbs := service.NewSimpleBotService()
    
    // Criar bot de atendimento
    botID := "atendimento-bot"
    err := sbs.CreateBot(ctx, botID, "Bot Atendimento", "whatsapp")
    if err != nil {
        panic(err)
    }
    
    // Construir fluxo
    sbs.AddMessageNode(ctx, botID, "menu", "1 - Suporte\n2 - Vendas\n3 - Informações")
    sbs.AddConfirmNode(ctx, botID, "confirm_support", "Falar com suporte?", "Sim", "Não")
    sbs.AddMessageNode(ctx, botID, "thanks", "Obrigado pelo contato!")
    
    // Conectar fluxo
    sbs.ConnectNodes(ctx, botID, "start", "menu", "início")
    sbs.ConnectNodes(ctx, botID, "menu", "confirm_support", "1")
    sbs.ConnectNodes(ctx, botID, "confirm_support", "thanks", "confirmed")
    
    // Validar
    result, _ := sbs.ValidateBot(ctx, botID, "whatsapp")
    if result.Valid {
        fmt.Println("Bot criado com sucesso!")
    }
}
```

## Comparação dos Serviços

| Característica | SimpleBotService | BotService |
|----------------|------------------|------------|
| Performance | Alta | Média |
| Simplicidade | Máxima | Média |
| Posicionamento Visual | Não | Sim |
| Patches RFC 6902 | Não | Sim |
| Histórico | Não | Sim |
| Casos de Uso | Bots simples e rápidos | Editores visuais |

## Demonstração

Execute a demonstração completa:

```bash
go test ./service/ -run TestIntegracaoCompleta -v
```

## Testes

```bash
# Testar SimpleBotService
go test ./service/ -run TestSimpleBotService -v

# Testar BotService
go test ./service/ -run TestBotService -v

# Todos os testes
go test ./service/ -v
```

## Outputs dos Componentes

- **message**: `["complete"]`
- **confirm**: `["confirmed", "cancelled", "timeout"]`
- **listpicker**: `["complete"]`
- **text**: `["complete"]`
- **delay**: `["complete"]`

## Status da Integração

- ValidationService: Integrado
- DesignService: Integrado  
- StoreService: Integrado
- Componentes: Todos suportados
- Validação: Automática
- Testes: Completos

A Lib-Bot fornece interface unificada e simplificada para criação de bots conversacionais com validação automática e integração completa.