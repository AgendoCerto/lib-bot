package component

import (
	"context"

	"lib-bot/liquid"
	"lib-bot/runtime"
)

// Media componente para envio de mídias (imagem, vídeo, áudio, documento)
type Media struct {
	mediaURL  string          // URL da mídia
	caption   string          // Legenda opcional
	filename  string          // Nome do arquivo (para documentos)
	mediaType string          // Tipo: image, video, audio, document, sticker
	det       liquid.Detector // Detector para parsing de templates Liquid
}

// NewMedia cria nova instância de componente media
func NewMedia(det liquid.Detector) *Media {
	return &Media{det: det}
}

// Kind retorna o tipo do componente
func (m *Media) Kind() string { return "media" }

// WithImage define uma imagem
func (m *Media) WithImage(url, caption string) *Media {
	cp := *m
	cp.mediaURL = url
	cp.caption = caption
	cp.mediaType = "image"
	return &cp
}

// WithVideo define um vídeo
func (m *Media) WithVideo(url, caption string) *Media {
	cp := *m
	cp.mediaURL = url
	cp.caption = caption
	cp.mediaType = "video"
	return &cp
}

// WithAudio define um áudio
func (m *Media) WithAudio(url string) *Media {
	cp := *m
	cp.mediaURL = url
	cp.mediaType = "audio"
	return &cp
}

// WithDocument define um documento
func (m *Media) WithDocument(url, caption, filename string) *Media {
	cp := *m
	cp.mediaURL = url
	cp.caption = caption
	cp.filename = filename
	cp.mediaType = "document"
	return &cp
}

// WithSticker define um sticker
func (m *Media) WithSticker(url string) *Media {
	cp := *m
	cp.mediaURL = url
	cp.mediaType = "sticker"
	return &cp
}

// Spec gera o ComponentSpec com parsing de templates
func (m *Media) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	// Processa caption se presente
	var textVal *TextValue
	if m.caption != "" {
		meta, err := m.det.Parse(ctx, m.caption)
		if err != nil {
			return ComponentSpec{}, err
		}
		textVal = &TextValue{
			Raw:      m.caption,
			Template: meta.IsTemplate,
			Liquid:   meta,
		}
	}

	// Cria metadata específica para mídia
	meta := map[string]any{
		"media_type": m.mediaType,
	}
	if m.filename != "" {
		meta["filename"] = m.filename
	}

	return ComponentSpec{
		Kind:     "media",
		Text:     textVal,
		MediaURL: m.mediaURL,
		Meta:     meta,
	}, nil
}

// Factory

type MediaFactory struct{ det liquid.Detector }

func NewMediaFactory(det liquid.Detector) *MediaFactory {
	return &MediaFactory{det: det}
}

func (f *MediaFactory) New(_ string, props map[string]any) (Component, error) {
	m := NewMedia(f.det)

	url, _ := props["url"].(string)
	caption, _ := props["caption"].(string)
	filename, _ := props["filename"].(string)
	mediaType, _ := props["type"].(string)

	// Define o tipo baseado na propriedade ou detecta pela URL
	switch mediaType {
	case "image":
		m = m.WithImage(url, caption)
	case "video":
		m = m.WithVideo(url, caption)
	case "audio":
		m = m.WithAudio(url)
	case "document":
		m = m.WithDocument(url, caption, filename)
	case "sticker":
		m = m.WithSticker(url)
	default:
		// Auto-detecta baseado na URL se não especificado
		if url != "" {
			m = m.WithImage(url, caption) // Default para imagem
		}
	}

	return m, nil
}
