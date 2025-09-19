package component

import (
	"context"

	"github.com/AgendoCerto/lib-bot/liquid"
	"github.com/AgendoCerto/lib-bot/runtime"
)

// Carousel componente para carrossel de cards (múltiplas mídias/produtos)
type Carousel struct {
	text  string          // Texto principal opcional
	cards []CardData      // Lista de cards do carrossel
	det   liquid.Detector // Detector para parsing de templates Liquid
}

// CardData representa um card no carrossel
type CardData struct {
	ID          string       `json:"id"`                    // ID único do card
	Title       string       `json:"title"`                 // Título do card
	Description string       `json:"description,omitempty"` // Descrição opcional
	MediaURL    string       `json:"media_url,omitempty"`   // URL da mídia (imagem)
	Buttons     []ButtonData `json:"buttons,omitempty"`     // Botões do card
	Price       string       `json:"price,omitempty"`       // Preço (para produtos)
}

// NewCarousel cria nova instância de componente carousel
func NewCarousel(det liquid.Detector) *Carousel {
	return &Carousel{
		det:   det,
		cards: make([]CardData, 0),
	}
}

// Kind retorna o tipo do componente
func (c *Carousel) Kind() string { return "carousel" }

// WithText define o texto principal
func (c *Carousel) WithText(s string) *Carousel {
	cp := *c
	cp.text = s
	return &cp
}

// AddCard adiciona um card ao carrossel
func (c *Carousel) AddCard(card CardData) *Carousel {
	cp := *c
	cp.cards = append(cp.cards, card)
	return &cp
}

// AddSimpleCard adiciona um card simples com título, descrição e imagem
func (c *Carousel) AddSimpleCard(id, title, description, mediaURL string) *Carousel {
	return c.AddCard(CardData{
		ID:          id,
		Title:       title,
		Description: description,
		MediaURL:    mediaURL,
	})
}

// AddProductCard adiciona um card de produto com preço e botões
func (c *Carousel) AddProductCard(id, title, description, mediaURL, price string, buttons []ButtonData) *Carousel {
	return c.AddCard(CardData{
		ID:          id,
		Title:       title,
		Description: description,
		MediaURL:    mediaURL,
		Price:       price,
		Buttons:     buttons,
	})
}

// Spec gera o ComponentSpec com parsing de templates
func (c *Carousel) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	// Processa texto principal se presente
	var textVal *TextValue
	if c.text != "" {
		meta, err := c.det.Parse(ctx, c.text)
		if err != nil {
			return ComponentSpec{}, err
		}
		textVal = &TextValue{
			Raw:      c.text,
			Template: meta.IsTemplate,
			Liquid:   meta,
		}
	}

	// Processa cards - aplica parsing Liquid nos textos
	processedCards := make([]CardData, 0, len(c.cards))
	for _, card := range c.cards {
		// Para simplificar, não aplicamos Liquid parsing nos cards aqui
		// mas isso poderia ser expandido conforme necessário
		processedCards = append(processedCards, card)
	}

	// Cria metadata específica para carrossel
	meta := map[string]any{
		"cards":         processedCards,
		"carousel_type": "generic", // Pode ser "generic" ou "product"
	}

	// Se todos os cards têm preço, considera como carrossel de produtos
	hasPrice := len(c.cards) > 0
	for _, card := range c.cards {
		if card.Price == "" {
			hasPrice = false
			break
		}
	}
	if hasPrice {
		meta["carousel_type"] = "product"
	}

	return ComponentSpec{
		Kind: "carousel",
		Text: textVal,
		Meta: meta,
	}, nil
}

// Factory

type CarouselFactory struct{ det liquid.Detector }

func NewCarouselFactory(det liquid.Detector) *CarouselFactory {
	return &CarouselFactory{det: det}
}

func (f *CarouselFactory) New(_ string, props map[string]any) (Component, error) {
	c := NewCarousel(f.det)

	// Texto principal
	if text, _ := props["text"].(string); text != "" {
		c = c.WithText(text)
	}

	// Cards
	if cardsRaw, ok := props["cards"].([]any); ok {
		for _, cardRaw := range cardsRaw {
			if cardMap, ok := cardRaw.(map[string]any); ok {
				id, _ := cardMap["id"].(string)
				title, _ := cardMap["title"].(string)
				description, _ := cardMap["description"].(string)
				mediaURL, _ := cardMap["media_url"].(string)
				price, _ := cardMap["price"].(string)

				// Processa botões do card
				var buttons []ButtonData
				if btnList, ok := cardMap["buttons"].([]any); ok {
					for _, btnRaw := range btnList {
						if btnMap, ok := btnRaw.(map[string]any); ok {
							label, _ := btnMap["label"].(string)
							payload, _ := btnMap["payload"].(string)
							kind, _ := btnMap["kind"].(string)
							url, _ := btnMap["url"].(string)

							if kind == "" {
								kind = "reply" // Default
							}

							buttons = append(buttons, ButtonData{
								Label:   label,
								Payload: payload,
								Kind:    kind,
								URL:     url,
							})
						}
					}
				}

				card := CardData{
					ID:          id,
					Title:       title,
					Description: description,
					MediaURL:    mediaURL,
					Price:       price,
					Buttons:     buttons,
				}

				c = c.AddCard(card)
			}
		}
	}

	// Parse behaviors
	behavior, err := ParseBehavior(props, f.det)
	if err != nil {
		return nil, err
	}

	return &CarouselWithBehavior{
		carousel: c,
		behavior: behavior,
	}, nil
}

// CarouselWithBehavior é um wrapper que inclui behaviors
type CarouselWithBehavior struct {
	carousel *Carousel
	behavior *ComponentBehavior
}

func (cwb *CarouselWithBehavior) Kind() string {
	return cwb.carousel.Kind()
}

func (cwb *CarouselWithBehavior) Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) {
	spec, err := cwb.carousel.Spec(ctx, rctx)
	if err != nil {
		return spec, err
	}
	spec.Behavior = cwb.behavior
	return spec, nil
}
