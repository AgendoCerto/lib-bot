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
	c.SupportsRichText = false                                                // Sem formatação rica
	c.MaxTextLen = 1024                                                       // Limite de texto do WhatsApp
	c.MaxButtons = 3                                                          // Máximo 3 botões
	c.ButtonKinds = map[string]bool{"reply": true, "url": true, "call": true} // Tipos suportados
	c.SupportsListPicker = true                                               // Suporte a listas

	// Limitações específicas do WhatsApp
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
	// Trunca texto se exceder limite
	if spec.Text != nil && len(spec.Text.Raw) > w.caps.MaxTextLen {
		spec.Text.Raw = spec.Text.Raw[:w.caps.MaxTextLen]
	}

	// Configura metadata para diferentes tipos de mensagem WhatsApp
	if spec.Meta == nil {
		spec.Meta = make(map[string]any)
	}

	// Se tem mídia, configura tipo específico
	if spec.MediaURL != "" {
		spec.Meta["whatsapp_type"] = w.detectMediaType(spec.MediaURL)
		spec.Meta["media_url"] = spec.MediaURL
	} else if spec.HSM != nil {
		spec.Meta["whatsapp_type"] = "template"
		spec.Meta["template_name"] = spec.HSM.Name
	} else {
		spec.Meta["whatsapp_type"] = "text"
		spec.Meta["preview_url"] = true // Habilita preview de links por padrão
	}

	return spec, nil
}

// transformButtons aplica transformações para botões interativos
func (w *WhatsApp) transformButtons(spec component.ComponentSpec) (component.ComponentSpec, error) {
	// Limita número de botões conforme capabilities
	if len(spec.Buttons) > w.caps.MaxButtons {
		spec.Buttons = spec.Buttons[:w.caps.MaxButtons]
	}

	// Filtra tipos de botão não suportados e aplica limitações de tamanho
	out := make([]component.Button, 0, len(spec.Buttons))
	for _, b := range spec.Buttons {
		if w.caps.ButtonKinds[b.Kind] {
			// Trunca título do botão se necessário
			if len(b.Label.Raw) > w.caps.MaxButtonTitleLen {
				b.Label.Raw = b.Label.Raw[:w.caps.MaxButtonTitleLen]
			}
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

	// Aplica limitações específicas do WhatsApp para listas
	w.applyMetaLimitations(spec.Meta)

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
	// Trunca texto se exceder limite
	if spec.Text != nil && len(spec.Text.Raw) > w.caps.MaxTextLen {
		spec.Text.Raw = spec.Text.Raw[:w.caps.MaxTextLen]
	}

	// Limita botões se presentes
	if len(spec.Buttons) > w.caps.MaxButtons {
		spec.Buttons = spec.Buttons[:w.caps.MaxButtons]
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

// applyMetaLimitations aplica limitações do WhatsApp em campos de metadata
func (w *WhatsApp) applyMetaLimitations(meta map[string]any) {
	// Limita items em listas
	if items, ok := meta["items"].([]any); ok && len(items) > w.caps.MaxListItems {
		meta["items"] = items[:w.caps.MaxListItems]
	}

	// Limita seções em listas
	if sections, ok := meta["sections"].([]any); ok && len(sections) > w.caps.MaxListSections {
		meta["sections"] = sections[:w.caps.MaxListSections]
	}

	// Trunca footer se muito longo
	if footer, ok := meta["footer"].(string); ok && len(footer) > w.caps.MaxFooterLen {
		meta["footer"] = footer[:w.caps.MaxFooterLen]
	}

	// Trunca header se muito longo
	if header, ok := meta["header"].(string); ok && len(header) > w.caps.MaxHeaderLen {
		meta["header"] = header[:w.caps.MaxHeaderLen]
	}

	// Trunca descrições se muito longas
	if desc, ok := meta["description"].(string); ok && len(desc) > w.caps.MaxDescriptionLen {
		meta["description"] = desc[:w.caps.MaxDescriptionLen]
	}
}
