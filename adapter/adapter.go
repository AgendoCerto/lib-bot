// Package adapter define interfaces para adapters de canal (WhatsApp, Telegram, etc.)
package adapter

import (
	"context"

	"lib-bot/component"
)

// Adapter interface para transformação de specs por canal específico (sem renderização)
type Adapter interface {
	Name() string                                                                                 // Nome do adapter (ex: "whatsapp")
	Capabilities() Capabilities                                                                   // Capacidades suportadas pelo canal
	Transform(ctx context.Context, spec component.ComponentSpec) (component.ComponentSpec, error) // Transforma spec para formato do canal
}

// Provider interface opcional para registro/busca de adapters
type Provider interface {
	Get(name string) (Adapter, bool) // Busca adapter por nome
	Register(a Adapter)              // Registra novo adapter
}

// DefaultProvider implementação padrão de provider
type DefaultProvider struct{ items map[string]Adapter }

// NewProvider cria um novo provider vazio
func NewProvider() *DefaultProvider { return &DefaultProvider{items: map[string]Adapter{}} }

// Get busca um adapter por nome
func (p *DefaultProvider) Get(name string) (Adapter, bool) { a, ok := p.items[name]; return a, ok }

// Register registra um adapter
func (p *DefaultProvider) Register(a Adapter) { p.items[a.Name()] = a }
