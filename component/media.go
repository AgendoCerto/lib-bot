package component

import (
	"context"

	"github.com/AgendoCerto/lib-bot/liquid"
	"github.com/AgendoCerto/lib-bot/runtime"
)

// Media componente para envio de mídias (imagem, vídeo, áudio, documento)
type Media struct {
	mediaURL  string          // URL da mídia
	caption   string          // Legenda opcional
	filename  string          // Nome do arquivo (para documentos)
	mediaType string          // Tipo: image, video, audio, document, sticker
	ptt       bool            // Push-to-talk (para áudio como mensagem de voz)
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

// WithAudio define um áudio (com opção de push-to-talk para mensagem de voz)
func (m *Media) WithAudio(url string, ptt bool) *Media {
	cp := *m
	cp.mediaURL = url
	cp.mediaType = "audio"
	cp.ptt = ptt
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
	if m.mediaType == "audio" {
		meta["ptt"] = m.ptt // Push-to-talk (mensagem de voz)
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
	ptt, _ := props["ptt"].(bool) // Push-to-talk para áudio

	// Define o tipo baseado na propriedade ou detecta pela URL
	switch mediaType {
	case "image":
		m = m.WithImage(url, caption)
	case "video":
		m = m.WithVideo(url, caption)
	case "audio":
		m = m.WithAudio(url, ptt)
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

	// Parse behaviors
	behavior, err := ParseBehavior(props, f.det)
	if err != nil {
		return nil, err
	}

	return &MediaWithBehavior{
		media:    m,
		behavior: behavior,
	}, nil
}

// MediaWithBehavior é um wrapper que inclui behaviors
type MediaWithBehavior struct {
	media    *Media
	behavior *ComponentBehavior
}

func (mwb *MediaWithBehavior) Kind() string {
	return mwb.media.Kind()
}

func (mwb *MediaWithBehavior) Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) {
	spec, err := mwb.media.Spec(ctx, rctx)
	if err != nil {
		return spec, err
	}
	spec.Behavior = mwb.behavior
	return spec, nil
}
