// Package whatsapp implementa adapter específico para WhatsApp Business API
package whatsapp

import (
	"context"
	"errors"

	"lib-bot/adapter"
	"lib-bot/component"
)

// WhatsApp adapter para WhatsApp Business API
type WhatsApp struct{ caps adapter.Capabilities }

// New cria novo adapter WhatsApp com capabilities específicas
func New() *WhatsApp {
	c := adapter.NewCaps()
	c.SupportsHSM = true                                                      // WhatsApp suporta HSM
	c.SupportsRichText = true                                                 // Rich text será processado pelo sistema final
	c.MaxTextLen = 1024                                                       // Limite de texto do WhatsApp
	c.MaxButtons = 3                                                          // Máximo 3 botões
	c.ButtonKinds = map[string]bool{"reply": true, "url": true, "call": true} // Tipos suportados
	c.SupportsListPicker = true                                               // Suporte a listas

	// Limitações específicas do WhatsApp para validação
	c.MaxListItems = 10      // Máximo 10 itens por lista/seção
	c.MaxListSections = 10   // Máximo 10 seções
	c.MaxButtonTitleLen = 24 // Máximo 24 caracteres no título do botão
	c.MaxDescriptionLen = 72 // Máximo 72 caracteres na descrição
	c.MaxFooterLen = 60      // Máximo 60 caracteres no footer
	c.MaxHeaderLen = 60      // Máximo 60 caracteres no header

	return &WhatsApp{caps: c}
}

func (w *WhatsApp) Name() string                       { return "whatsapp" }
func (w *WhatsApp) Capabilities() adapter.Capabilities { return w.caps }

// Transform aplica transformações específicas do WhatsApp aos specs (sem renderização)
func (w *WhatsApp) Transform(_ context.Context, spec component.ComponentSpec) (component.ComponentSpec, error) {
	// Verifica suporte a HSM
	if spec.HSM != nil && !w.caps.SupportsHSM {
		return component.ComponentSpec{}, errors.New("adapter: HSM not supported")
	}

	// Aplica transformações específicas baseadas no tipo de componente
	switch spec.Kind {
	case "message":
		return w.transformMessage(spec)
	case "buttons":
		return w.transformButtons(spec)
	case "listpicker":
		return w.transformListPicker(spec)
	case "menu":
		return w.transformMenu(spec)
	case "carousel":
		return w.transformCarousel(spec)
	default:
		return w.transformGeneric(spec)
	}
}

// transformMessage aplica transformações específicas para mensagens
func (w *WhatsApp) transformMessage(spec component.ComponentSpec) (component.ComponentSpec, error) {
	// Configura metadata para diferentes tipos de mensagem WhatsApp
	if spec.Meta == nil {
		spec.Meta = make(map[string]any)
	}

	// Define tipo de mensagem baseado no conteúdo
	switch {
	case spec.MediaURL != "":
		spec.Meta["whatsapp_type"] = w.detectMediaType(spec.MediaURL)
		spec.Meta["media_url"] = spec.MediaURL
	case spec.HSM != nil:
		spec.Meta["whatsapp_type"] = "template"
		spec.Meta["template_name"] = spec.HSM.Name
	default:
		spec.Meta["whatsapp_type"] = "text"
		spec.Meta["preview_url"] = true // Habilita preview de links por padrão
	}

	return spec, nil
}

// transformButtons aplica transformações para botões interativos
func (w *WhatsApp) transformButtons(spec component.ComponentSpec) (component.ComponentSpec, error) {
	// Filtra apenas tipos de botão suportados (sem truncar ou limitar quantidade)
	out := make([]component.Button, 0, len(spec.Buttons))
	for _, b := range spec.Buttons {
		if w.caps.ButtonKinds[b.Kind] {
			out = append(out, b)
		}
	}
	spec.Buttons = out

	// Configura metadata do WhatsApp
	if spec.Meta == nil {
		spec.Meta = make(map[string]any)
	}
	spec.Meta["whatsapp_type"] = "interactive"
	spec.Meta["interactive_type"] = "button"

	return spec, nil
}

// transformListPicker aplica transformações para listas
func (w *WhatsApp) transformListPicker(spec component.ComponentSpec) (component.ComponentSpec, error) {
	if spec.Meta == nil {
		spec.Meta = make(map[string]any)
	}

	spec.Meta["whatsapp_type"] = "interactive"
	spec.Meta["interactive_type"] = "list"

	return spec, nil
}

// transformMenu similar a ListPicker mas com configuração específica
func (w *WhatsApp) transformMenu(spec component.ComponentSpec) (component.ComponentSpec, error) {
	return w.transformListPicker(spec) // Menu e ListPicker são tratados igual no WhatsApp
}

// transformCarousel mapeia carrossel para product_list no WhatsApp
func (w *WhatsApp) transformCarousel(spec component.ComponentSpec) (component.ComponentSpec, error) {
	if spec.Meta == nil {
		spec.Meta = make(map[string]any)
	}

	spec.Meta["whatsapp_type"] = "interactive"
	spec.Meta["interactive_type"] = "product_list"

	return spec, nil
}

// transformGeneric aplica transformações básicas para tipos genéricos
func (w *WhatsApp) transformGeneric(spec component.ComponentSpec) (component.ComponentSpec, error) {
	// Apenas configura metadados básicos, sem modificar conteúdo
	if spec.Meta == nil {
		spec.Meta = make(map[string]any)
	}

	// Define tipo básico se ainda não definido
	if spec.Meta["whatsapp_type"] == nil {
		spec.Meta["whatsapp_type"] = "text"
	}

	return spec, nil
}

// detectMediaType detecta o tipo de mídia baseado na URL
func (w *WhatsApp) detectMediaType(url string) string {
	// Lógica simples baseada na extensão do arquivo
	if url == "" {
		return "text"
	}

	// Pode ser expandido com detecção mais sofisticada
	switch {
	case contains(url, ".jpg") || contains(url, ".png") || contains(url, ".jpeg"):
		return "image"
	case contains(url, ".mp4") || contains(url, ".mov"):
		return "video"
	case contains(url, ".mp3") || contains(url, ".ogg") || contains(url, ".wav"):
		return "audio"
	case contains(url, ".pdf") || contains(url, ".doc"):
		return "document"
	case contains(url, ".webp"):
		return "sticker"
	default:
		return "document"
	}
}

// contains função utilitária para verificar substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr
}
