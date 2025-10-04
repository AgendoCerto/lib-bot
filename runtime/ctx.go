// Package runtime define contextos de execução para templates e variáveis
package runtime

// Context contém variáveis de contexto disponíveis durante execução
type Context struct {
	Context map[string]any // Variáveis de contexto da sessão (temporárias)
	State   map[string]any // Variáveis do state (persistentes - usuário)
	Global  map[string]any // Variáveis globais (persistentes - bot)
}

// LiquidScope retorna o escopo completo para renderização de templates Liquid
func (c Context) LiquidScope() map[string]any {
	return map[string]any{
		"context": c.Context,
		"state":   c.State,
		"global":  c.Global,
	}
}
