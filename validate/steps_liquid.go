package validate

import (
	"github.com/AgendoCerto/lib-bot/adapter"
	"github.com/AgendoCerto/lib-bot/component"
	"github.com/AgendoCerto/lib-bot/io"
	"github.com/AgendoCerto/lib-bot/liquid"
)

type LiquidStep struct {
	Policy liquid.Policy
	Linter liquid.Linter
	// Adicionado para suportar validação com contexto global
	designDoc *io.DesignDoc
}

// NewLiquidStep cria um novo validador de templates Liquid
func NewLiquidStep() *LiquidStep {
	return &LiquidStep{
		Policy: liquid.Policy{
			StrictVars:     true,
			AllowedFilters: map[string]bool{"upcase": true, "downcase": true, "capitalize": true},
			MaxDepth:       5,
		},
		Linter: liquid.SimpleLinter{},
	}
}

// SetDesignContext define o contexto do design para validação
func (s *LiquidStep) SetDesignContext(doc *io.DesignDoc) {
	s.designDoc = doc
}

func (s LiquidStep) Check(spec component.ComponentSpec, _ adapter.Capabilities, path string) []Issue {
	var issues []Issue

	// Collect available keys: context/profile keys from persistence, plus WhatsApp defaults
	availableKeys := map[string]bool{
		"context.wa_phone": true,
		"context.wa_name":  true,
	}

	// Adiciona variáveis globais do profile.context se disponível
	if s.designDoc != nil {
		for key := range s.designDoc.Profile.Context {
			availableKeys["context."+key] = true
		}
	}

	if spec.Persistence != nil && spec.Persistence.Enabled {
		key := spec.Persistence.Key
		scope := spec.Persistence.Scope
		if key != "" {
			if scope == "context" {
				availableKeys["context."+key] = true
			} else if scope == "profile" {
				availableKeys["profile."+key] = true
			}
		}
	}

	checkVars := func(meta liquid.Meta, path string) {
		for _, v := range meta.Vars {
			// Accept dot notation (context.key, profile.key)
			if !availableKeys[v] {
				issues = append(issues, Issue{
					Code:     "liquid.var.missing_persistence",
					Severity: Err,
					Path:     path,
					Msg:      "Liquid variable '" + v + "' is not available in persisted keys/context/profile.",
				})
			}
		}
	}

	if spec.Text != nil && spec.Text.Template {
		for _, is := range s.Linter.Lint(spec.Text.Liquid, s.Policy, path+".view.text") {
			issues = append(issues, Issue{Code: is.Code, Severity: toSev(is.Severity), Path: is.Path, Msg: is.Msg})
		}
		checkVars(spec.Text.Liquid, path+".view.text")
	}
	for i := range spec.Buttons {
		if spec.Buttons[i].Label.Template {
			for _, is := range s.Linter.Lint(spec.Buttons[i].Label.Liquid, s.Policy, path+".view.buttons") {
				issues = append(issues, Issue{Code: is.Code, Severity: toSev(is.Severity), Path: is.Path, Msg: is.Msg})
			}
			checkVars(spec.Buttons[i].Label.Liquid, path+".view.buttons")
		}
	}
	if spec.HSM != nil {
		// HSM simplificado não tem parâmetros ou botões complexos para validar
		// A validação do nome já é feita no HSMValidationStep
	}
	return issues
}

func toSev(s string) Severity {
	switch s {
	case "error":
		return Err
	case "warn":
		return Warn
	default:
		return Info
	}
}
