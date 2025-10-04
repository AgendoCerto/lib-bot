// Package runtime define contextos de execução para templates e variáveis
package runtime

// Context contém variáveis de contexto disponíveis durante execução
type Context struct {
	Context map[string]any // Variáveis de contexto da sessão (temporárias)
	State   map[string]any // Variáveis do state (persistentes - usuário)
	Global  map[string]any // Variáveis globais (persistentes - bot)
}

// LiquidScope retorna o escopo completo para renderização de templates Liquid
// ✅ Garante que default_client sempre existe (variável VIRTUAL do sistema)
func (c Context) LiquidScope() map[string]any {
	// ✅ Garantir que default_client existe no context (variável virtual do sistema)
	contextScope := c.Context
	if contextScope == nil {
		contextScope = make(map[string]any)
	}

	// Se default_client não existe, criar vazio (será preenchido pelo runtime)
	if _, exists := contextScope["default_client"]; !exists {
		contextScope["default_client"] = map[string]any{
			"phone_number": "",
			"name":         "",
			"captured_at":  "",
		}
	}

	return map[string]any{
		"context": contextScope,
		"state":   c.State,
		"global":  c.Global,
	}
}

// EnsureDefaultClient garante que default_client existe no context
// Esta função deve ser chamada sempre que um Context for criado
func (c *Context) EnsureDefaultClient() {
	if c.Context == nil {
		c.Context = make(map[string]any)
	}

	if _, exists := c.Context["default_client"]; !exists {
		c.Context["default_client"] = map[string]any{
			"phone_number": "",
			"name":         "",
			"captured_at":  "",
		}
	}
}
