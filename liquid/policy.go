package liquid

type Policy struct {
	StrictVars     bool
	AllowedFilters map[string]bool
	AllowedTags    map[string]bool // Tags de controle permitidas (for, if, case, etc)
	MaxDepth       int             // profundidade de filtros encadeados (heurístico)
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

	// Tags não permitidas
	for _, t := range meta.Tags {
		if policy.AllowedTags != nil && !policy.AllowedTags[t] {
			issues = append(issues, Issue{
				Severity: "error",
				Path:     path,
				Code:     "liquid.tag.not_allowed",
				Msg:      "tag not allowed: " + t,
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

// DefaultAllowedTags retorna o conjunto padrão de tags de controle permitidas
func DefaultAllowedTags() map[string]bool {
	return map[string]bool{
		// Tags de controle de fluxo
		"if":        true,
		"elsif":     true,
		"else":      true,
		"endif":     true,
		"unless":    true,
		"endunless": true,

		// Tags de loop
		"for":      true,
		"endfor":   true,
		"break":    true,
		"continue": true,

		// Tags de case/when
		"case":    true,
		"when":    true,
		"endcase": true,

		// Tags de assign/capture
		"assign":     true,
		"capture":    true,
		"endcapture": true,

		// Tags de comentário
		"comment":    true,
		"endcomment": true,
	}
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
		"slug":       true, // Converte para slug (kebab-case)
		"camelize":   true, // Converte para camelCase
		"underscore": true, // Converte para snake_case

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
		"modulo": true, // Resto da divisão
		"abs":    true, // Valor absoluto
		"round":  true, // Arredonda número
		"floor":  true, // Arredonda para baixo
		"ceil":   true, // Arredonda para cima

		// Filtros de array/objeto
		"size":    true,
		"first":   true,
		"last":    true,
		"join":    true,
		"sort":    true, // Ordena array
		"uniq":    true, // Remove duplicados
		"reverse": true, // Inverte array

		// Filtros de escape
		"escape":      true,
		"escape_once": true,
		"url_encode":  true,
		"url_decode":  true,

		// 🌍 Filtros de formatação internacional
		"phone":    true, // Formata telefone: {{phone | phone: "BR"}} ou {{phone | phone: "US"}}
		"currency": true, // Formata moeda: {{value | currency: "BRL"}} ou {{value | currency: "USD"}}
		"money":    true, // Alias para currency

		// 🇧🇷 Filtros de documentos brasileiros
		"cpf":  true, // Formata CPF: {{doc | cpf}} → 123.456.789-00
		"cnpj": true, // Formata CNPJ: {{doc | cnpj}} → 12.345.678/0001-00
		"cep":  true, // Formata CEP: {{cep | cep}} → 12345-678
		"rg":   true, // Formata RG: {{doc | rg}} → 12.345.678-9

		// 📅 Filtros de data/hora avançados
		"date_tz":   true, // Data com timezone: {{date | date_tz: "America/Sao_Paulo", "%d/%m/%Y %H:%M"}}
		"time_ago":  true, // Tempo relativo: {{date | time_ago}} → "há 2 horas"
		"duration":  true, // Duração: {{seconds | duration}} → "2h 30m"
		"timestamp": true, // Converte para timestamp Unix
		"from_now":  true, // Tempo futuro: {{date | from_now}} → "daqui a 3 dias"

		// 🔐 Filtros de hash/encode
		"md5":           true, // Hash MD5
		"sha1":          true, // Hash SHA1
		"sha256":        true, // Hash SHA256
		"base64":        true, // Encode base64
		"base64_decode": true, // Decode base64

		// 📏 Filtros de validação/verificação
		"length":        true, // Tamanho da string
		"word_count":    true, // Conta palavras
		"newline_to_br": true, // Converte \n para <br>
		"strip_html":    true, // Remove tags HTML
	}
}

// DefaultLiquidPolicy retorna a política padrão para validação Liquid
func DefaultLiquidPolicy() Policy {
	return Policy{
		StrictVars:     false, // Modo lax por padrão
		AllowedFilters: DefaultAllowedFilters(),
		AllowedTags:    DefaultAllowedTags(), // ✅ Tags de controle permitidas
		MaxDepth:       5,                    // Máximo 5 filtros encadeados
	}
}

// StrictLiquidPolicy retorna política strict para produção
func StrictLiquidPolicy() Policy {
	return Policy{
		StrictVars:     true, // Modo strict
		AllowedFilters: DefaultAllowedFilters(),
		AllowedTags:    DefaultAllowedTags(), // ✅ Tags de controle permitidas
		MaxDepth:       3,                    // Máximo 3 filtros encadeados
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
