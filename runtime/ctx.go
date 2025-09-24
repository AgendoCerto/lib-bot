// Package runtime define contextos de execução para templates e variáveis
package runtime

// Context contém variáveis de contexto disponíveis durante execução
type Context struct {
	Context map[string]any // Variáveis de contexto da sessão (temporárias)
	Profile map[string]any // Variáveis do perfil (persistentes)
}

// LiquidScope retorna o escopo completo para renderização de templates Liquid
func (c Context) LiquidScope() map[string]any {
	return map[string]any{
		"context": c.Context,
		"profile": c.Profile,
	}
}
