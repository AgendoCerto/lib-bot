// Package runtime define contextos de execução para templates e variáveis
package runtime

// Context contém variáveis de contexto disponíveis durante execução
type Context struct {
	Flow map[string]any // Variáveis do fluxo de conversação
	User map[string]any // Dados do usuário (perfil, preferências, etc.)
	Sys  map[string]any // Variáveis do sistema (timestamp, canal, etc.)
}

// LiquidScope retorna o escopo completo para renderização de templates Liquid
func (c Context) LiquidScope() map[string]any {
	return map[string]any{
		"flow": c.Flow,
		"user": c.User,
		"sys":  c.Sys,
	}
}
