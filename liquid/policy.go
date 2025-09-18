package liquid

type Policy struct {
	StrictVars     bool
	AllowedFilters map[string]bool
	MaxDepth       int // profundidade de filtros encadeados (heurístico)
}

type Issue struct {
	Severity string // "warn"|"error"
	Msg      string
	Path     string
	Code     string
}

type Linter interface {
	Lint(meta Meta, policy Policy, path string) []Issue
}

type SimpleLinter struct{}

func (SimpleLinter) Lint(meta Meta, policy Policy, path string) []Issue {
	var issues []Issue

	// Filtros não permitidos
	for _, f := range meta.Filters {
		if policy.AllowedFilters != nil && !policy.AllowedFilters[f] {
			issues = append(issues, Issue{
				Severity: "error",
				Path:     path,
				Code:     "liquid.filter.not_allowed",
				Msg:      "filter not allowed: " + f,
			})
		}
	}

	// Validação de profundidade máxima (heurístico simples)
	if policy.MaxDepth > 0 {
		filterCount := len(meta.Filters)
		if filterCount > policy.MaxDepth {
			issues = append(issues, Issue{
				Severity: "warn",
				Path:     path,
				Code:     "liquid.filter.depth_exceeded",
				Msg:      "filter chain depth exceeds maximum: " + string(rune(filterCount+'0')),
			})
		}
	}

	// Validação de variáveis em modo strict
	if policy.StrictVars {
		for _, varRef := range meta.Vars {
			// Em modo strict, pode validar se variáveis existem no contexto
			// Por agora, apenas avisa sobre variáveis não reconhecidas
			if !isKnownNamespace(varRef) {
				issues = append(issues, Issue{
					Severity: "warn",
					Path:     path,
					Code:     "liquid.var.unknown_namespace",
					Msg:      "variable uses unknown namespace: " + varRef,
				})
			}
		}
	}

	return issues
}

// DefaultAllowedFilters retorna o conjunto padrão de filtros permitidos conforme documentação
func DefaultAllowedFilters() map[string]bool {
	return map[string]bool{
		// Filtros de texto básicos
		"upcase":     true,
		"downcase":   true,
		"capitalize": true,
		"strip":      true,
		"truncate":   true,
		"replace":    true,

		// Filtros de formatação
		"date":   true,
		"number": true,

		// Filtros de controle
		"default": true,

		// Filtros de dados
		"json": true,

		// Filtros matemáticos básicos
		"plus":   true,
		"minus":  true,
		"times":  true,
		"divide": true,

		// Filtros de array/objeto
		"size":  true,
		"first": true,
		"last":  true,
		"join":  true,

		// Filtros de escape
		"escape":      true,
		"escape_once": true,
		"url_encode":  true,
	}
}

// DefaultLiquidPolicy retorna a política padrão para validação Liquid
func DefaultLiquidPolicy() Policy {
	return Policy{
		StrictVars:     false, // Modo lax por padrão
		AllowedFilters: DefaultAllowedFilters(),
		MaxDepth:       5, // Máximo 5 filtros encadeados
	}
}

// StrictLiquidPolicy retorna política strict para produção
func StrictLiquidPolicy() Policy {
	return Policy{
		StrictVars:     true, // Modo strict
		AllowedFilters: DefaultAllowedFilters(),
		MaxDepth:       3, // Máximo 3 filtros encadeados
	}
}

// isKnownNamespace verifica se a variável usa um namespace conhecido
func isKnownNamespace(varRef string) bool {
	knownNamespaces := []string{"flow.", "user.", "sys."}

	for _, ns := range knownNamespaces {
		if len(varRef) >= len(ns) && varRef[:len(ns)] == ns {
			return true
		}
	}

	// Variáveis sem namespace também são válidas (ex: "name" ao invés de "user.name")
	return true
}
