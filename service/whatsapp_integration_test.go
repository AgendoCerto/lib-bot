package service

import (
	"context"
	"testing"
)

// TestWhatsAppValidation testa validações WhatsApp integradas
func TestWhatsAppValidation(t *testing.T) {
	ctx := context.Background()
	svc := NewBotService()

	// 1. Criar bot
	err := svc.CreateBot(ctx, "test-whatsapp", "Test Bot", "whatsapp")
	if err != nil {
		t.Fatalf("Erro ao criar bot: %v", err)
	}

	// 2. Testar buttons com mais de 3 botões (deve falhar)
	t.Run("Buttons - Máximo 3 botões", func(t *testing.T) {
		buttons := []map[string]string{
			{"label": "Opção 1", "payload": "opt1"},
			{"label": "Opção 2", "payload": "opt2"},
			{"label": "Opção 3", "payload": "opt3"},
			{"label": "Opção 4", "payload": "opt4"}, // 4º botão - deve falhar
		}

		err := svc.AddButtonsNode(ctx, "test-whatsapp", "btn_invalid", "Escolha:", buttons, "", "")
		if err == nil {
			t.Error("Esperava erro com 4 botões, mas não houve erro")
		} else {
			t.Logf("✅ Validação correta: %v", err)
		}
	})

	// 3. Testar buttons válido (máx 3)
	t.Run("Buttons - Válido com 3 botões", func(t *testing.T) {
		buttons := []map[string]string{
			{"label": "Sim", "payload": "yes"},
			{"label": "Não", "payload": "no"},
			{"label": "Talvez", "payload": "maybe"},
		}

		err := svc.AddButtonsNode(ctx, "test-whatsapp", "btn_valid", "Aceita?", buttons, "Confirmação", "Responda por favor")
		if err != nil {
			t.Errorf("Erro inesperado: %v", err)
		} else {
			t.Log("✅ Buttons válido criado com sucesso")
		}
	})

	// 4. Testar label muito longo (>20 chars)
	t.Run("Buttons - Label muito longo", func(t *testing.T) {
		buttons := []map[string]string{
			{"label": "Este label tem mais de vinte caracteres", "payload": "long"},
		}

		err := svc.AddButtonsNode(ctx, "test-whatsapp", "btn_long_label", "Teste:", buttons, "", "")
		if err == nil {
			t.Error("Esperava erro com label longo, mas não houve erro")
		} else {
			t.Logf("✅ Validação correta: %v", err)
		}
	})

	// 5. Testar header muito longo (>60 chars)
	t.Run("Buttons - Header muito longo", func(t *testing.T) {
		buttons := []map[string]string{
			{"label": "OK", "payload": "ok"},
		}

		longHeader := "Este header tem mais de sessenta caracteres e deve falhar na validação"

		err := svc.AddButtonsNode(ctx, "test-whatsapp", "btn_long_header", "Teste:", buttons, longHeader, "")
		if err == nil {
			t.Error("Esperava erro com header longo, mas não houve erro")
		} else {
			t.Logf("✅ Validação correta: %v", err)
		}
	})

	// 6. Testar listpicker com mais de 10 seções
	t.Run("ListPicker - Máximo 10 seções", func(t *testing.T) {
		sections := make([]map[string]any, 11) // 11 seções - deve falhar
		for i := 0; i < 11; i++ {
			sections[i] = map[string]any{
				"title": "Seção",
				"items": []map[string]any{
					{"id": "item1", "title": "Item"},
				},
			}
		}

		err := svc.AddListPickerNode(ctx, "test-whatsapp", "list_invalid", "Escolha:", "Ver opções", sections, "", "")
		if err == nil {
			t.Error("Esperava erro com 11 seções, mas não houve erro")
		} else {
			t.Logf("✅ Validação correta: %v", err)
		}
	})

	// 7. Testar listpicker válido
	t.Run("ListPicker - Válido", func(t *testing.T) {
		sections := []map[string]any{
			{
				"title": "Serviços",
				"items": []map[string]any{
					{"id": "srv1", "title": "Consulta", "description": "Agendamento de consulta"},
					{"id": "srv2", "title": "Exame", "description": "Agendamento de exame"},
				},
			},
		}

		err := svc.AddListPickerNode(ctx, "test-whatsapp", "list_valid", "Escolha um serviço:", "Ver Serviços", sections, "Menu Principal", "Atendimento 24h")
		if err != nil {
			t.Errorf("Erro inesperado: %v", err)
		} else {
			t.Log("✅ ListPicker válido criado com sucesso")
		}
	})

	// 8. Testar mídia com audio e ptt
	t.Run("Media - Audio com PTT", func(t *testing.T) {
		err := svc.AddMediaNode(ctx, "test-whatsapp", "audio_voice", "audio", "https://example.com/voice.mp3", "", "", true)
		if err != nil {
			t.Errorf("Erro inesperado: %v", err)
		} else {
			t.Log("✅ Áudio com PTT criado com sucesso")
		}
	})

	// 9. Testar documento sem filename (deve falhar)
	t.Run("Media - Document sem filename", func(t *testing.T) {
		err := svc.AddMediaNode(ctx, "test-whatsapp", "doc_invalid", "document", "https://example.com/doc.pdf", "Documento", "", false)
		if err == nil {
			t.Error("Esperava erro com documento sem filename, mas não houve erro")
		} else {
			t.Logf("✅ Validação correta: %v", err)
		}
	})

	// 10. Validação final do bot
	t.Run("Validação Final", func(t *testing.T) {
		result, err := svc.ValidateBot(ctx, "test-whatsapp", "whatsapp")
		if err != nil {
			t.Errorf("Erro ao validar: %v", err)
		} else {
			t.Logf("Valid: %v, Issues: %d", result.Valid, len(result.Issues))

			if len(result.Issues) > 0 {
				t.Log("Issues encontrados:")
				for _, issue := range result.Issues {
					t.Logf("  - [%s] %s: %s", issue.Severity, issue.Code, issue.Msg)
				}
			}
		}
	})
}
