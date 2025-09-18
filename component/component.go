// Package component define tipos e interfaces para componentes de conversação
package component

import (
	"context"

	"lib-bot/liquid"
	"lib-bot/runtime"
)

// TextValue armazena texto com suporte a templates Liquid (sem renderização)
type TextValue struct {
	Raw      string      `json:"raw"`      // Texto original com possíveis templates
	Template bool        `json:"template"` // Indica se contém templates Liquid
	Liquid   liquid.Meta `json:"liquid"`   // Metadados de parsing do Liquid
}

// HSMView representa uma HSM (Highly Structured Message) com parâmetros templated
type HSMView struct {
	ID        string      `json:"id"`                  // Identificador da HSM
	Locale    string      `json:"locale"`              // Localização (ex: pt_BR)
	Namespace string      `json:"namespace,omitempty"` // Namespace da HSM
	Params    []TextValue `json:"params"`              // Parâmetros (podem conter Liquid)
	Buttons   []Button    `json:"buttons,omitempty"`   // Botões interativos
	Policy    string      `json:"policy,omitempty"`    // Política de fallback: error_on_missing|fallback_to_text|fallback_to_menu
}

// Button representa um botão interativo
type Button struct {
	Label   TextValue `json:"label"`   // Texto do botão (pode ter templates)
	Payload string    `json:"payload"` // Dados enviados ao clicar
	Kind    string    `json:"kind"`    // Tipo: reply|url|call
}

// ComponentSpec é o modelo canônico de um componente (sem renderização final)
type ComponentSpec struct {
	Kind     string         `json:"kind"`                // Tipo do componente (message, confirm, etc.)
	Text     *TextValue     `json:"text,omitempty"`      // Texto principal
	MediaURL string         `json:"media_url,omitempty"` // URL de mídia (imagem, vídeo, etc.)
	Buttons  []Button       `json:"buttons,omitempty"`   // Botões interativos
	HSM      *HSMView       `json:"hsm,omitempty"`       // Configuração de HSM
	Meta     map[string]any `json:"meta,omitempty"`      // Metadados adicionais
}

// Component interface para geração de specs canônicos (apenas parsing, sem render)
type Component interface {
	Kind() string                                                          // Retorna o tipo do componente
	Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) // Gera spec canônico
}

// Factory interface para criação de componentes a partir de propriedades
type Factory interface {
	New(kind string, props map[string]any) (Component, error) // Cria componente das props do design
}

// Registry gerencia fábricas de componentes por tipo
type Registry struct{ factories map[string]Factory }

// NewRegistry cria um novo registry vazio
func NewRegistry() *Registry { return &Registry{factories: map[string]Factory{}} }

// Register registra uma factory para um tipo específico
func (r *Registry) Register(kind string, f Factory) { r.factories[kind] = f }

// New cria um componente do tipo especificado com as propriedades fornecidas
func (r *Registry) New(kind string, props map[string]any) (Component, error) {
	f, ok := r.factories[kind]
	if !ok {
		return nil, ErrUnknownKind{Kind: kind}
	}
	return f.New(kind, props)
}

// ErrUnknownKind erro retornado quando um tipo de componente não é encontrado
type ErrUnknownKind struct{ Kind string }

func (e ErrUnknownKind) Error() string { return "unknown component kind: " + e.Kind }
