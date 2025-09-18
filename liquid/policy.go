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
	// (Opcional) profundidade/encadeamento pode ser checado por contagem de '|' – heurístico simples
	return issues
}
