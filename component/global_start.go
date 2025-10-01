package component

import (
	"context"

	"github.com/AgendoCerto/lib-bot/liquid"
	"github.com/AgendoCerto/lib-bot/runtime"
)

// GlobalStart representa o componente de início global do bot
// Este é um componente especial que serve como ponto de entrada do fluxo
type GlobalStart struct {
	infoMessage string          // Mensagem informativa (somente leitura)
	det         liquid.Detector // Detector para parsing de templates Liquid
}

// NewGlobalStart cria nova instância de componente global_start
func NewGlobalStart(det liquid.Detector) *GlobalStart {
	return &GlobalStart{det: det}
}

func (gs *GlobalStart) Kind() string { return "global_start" }

// WithInfoMessage define a mensagem informativa
func (gs *GlobalStart) WithInfoMessage(s string) *GlobalStart {
	cp := *gs
	cp.infoMessage = s
	return &cp
}

// Spec gera o ComponentSpec para global_start
func (gs *GlobalStart) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	// global_start é um componente especial que não renderiza conteúdo
	// Serve apenas como ponto de entrada visual no editor
	spec := ComponentSpec{
		Kind: "global_start",
		Meta: map[string]any{
			"info_message": gs.infoMessage,
			"system_node":  true, // Marca como nó do sistema
		},
	}

	return spec, nil
}

// GlobalStartFactory factory para criar componentes global_start
type GlobalStartFactory struct{ det liquid.Detector }

func NewGlobalStartFactory(det liquid.Detector) *GlobalStartFactory {
	return &GlobalStartFactory{det: det}
}

func (f *GlobalStartFactory) New(_ string, props map[string]any) (Component, error) {
	gs := NewGlobalStart(f.det)

	// Extrair info_message das props se existir
	if infoMsg, ok := props["info_message"].(string); ok {
		gs = gs.WithInfoMessage(infoMsg)
	} else {
		// Valor padrão
		gs = gs.WithInfoMessage("Este é o ponto de entrada do seu bot. Conecte-o ao primeiro componente.")
	}

	return gs, nil
}
