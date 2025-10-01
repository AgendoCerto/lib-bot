// Package liquid fornece funcionalidades para detecção e parsing de templates Liquid
package liquid

import (
	"context"
	"regexp"
	"strings"
)

// Detector interface para parsing de templates Liquid (sem renderização)
type Detector interface {
	Parse(ctx context.Context, input string) (Meta, error) // Analisa template e extrai metadados
}

// Meta contém metadados de um template Liquid após parsing
type Meta struct {
	IsTemplate         bool     `json:"isTemplate"`         // Indica se contém templates
	Vars               []string `json:"vars"`               // Variáveis encontradas
	Filters            []string `json:"filters"`            // Filtros utilizados
	Tags               []string `json:"tags"`               // Tags de controle encontradas (for, if, case, etc)
	EstimatedStaticLen int      `json:"estimatedStaticLen"` // Tamanho estimado após renderização
}

// NoRenderDetector implementação que apenas detecta patterns sem renderizar
type NoRenderDetector struct{}

// Parse analisa string e extrai metadados de templates Liquid
func (NoRenderDetector) Parse(_ context.Context, s string) (Meta, error) {
	m := Meta{EstimatedStaticLen: len(stripDelims(s))}

	// Detectar variáveis {{}}
	if strings.Contains(s, "{{") && strings.Contains(s, "}}") {
		m.IsTemplate = true
		m.Vars = extractVars(s)
		m.Filters = extractFilters(s)
	}

	// Detectar tags de controle {% %}
	if strings.Contains(s, "{%") && strings.Contains(s, "%}") {
		m.IsTemplate = true
		m.Tags = extractTags(s)
	}

	return m, nil
}

var (
	reVar    = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.]+)`)
	reFilter = regexp.MustCompile(`\|\s*([a-zA-Z0-9_]+)`)
	reTag    = regexp.MustCompile(`\{%\s*([a-zA-Z0-9_]+)`) // Extrai nome da tag (for, if, case, etc)
)

func extractVars(s string) []string {
	m := reVar.FindAllStringSubmatch(s, -1)
	out := make([]string, 0, len(m))
	seen := map[string]bool{}
	for _, g := range m {
		if len(g) > 1 && !seen[g[1]] {
			out = append(out, g[1])
			seen[g[1]] = true
		}
	}
	return out
}

func extractFilters(s string) []string {
	m := reFilter.FindAllStringSubmatch(s, -1)
	out := make([]string, 0, len(m))
	seen := map[string]bool{}
	for _, g := range m {
		if len(g) > 1 && !seen[g[1]] {
			out = append(out, g[1])
			seen[g[1]] = true
		}
	}
	return out
}

func extractTags(s string) []string {
	m := reTag.FindAllStringSubmatch(s, -1)
	out := make([]string, 0, len(m))
	seen := map[string]bool{}
	for _, g := range m {
		if len(g) > 1 && !seen[g[1]] {
			out = append(out, g[1])
			seen[g[1]] = true
		}
	}
	return out
}

func stripDelims(s string) string {
	s = strings.ReplaceAll(s, "{{", "")
	s = strings.ReplaceAll(s, "}}", "")
	s = strings.ReplaceAll(s, "{%", "")
	s = strings.ReplaceAll(s, "%}", "")
	return s
}
