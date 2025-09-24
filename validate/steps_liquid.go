package validate

import (
	"strings"

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

	// Chaves de contexto sempre disponíveis (sessão baseada em janelas)
	availableKeys := map[string]bool{
		// Contexto padrão do cliente/sessão - SEMPRE disponível
		"context.name":         true,
		"context.phone_number": true,
		"context.captured_at":  true,
	}

	// Adiciona variáveis do profile se disponível
	if s.designDoc != nil {
		// Variáveis do profile.context (definições) - apenas context. e profile.
		for key, profileVar := range s.designDoc.Profile.Context {
			if profileVar.Persist {
				// Variáveis persistentes vão para profile scope
				availableKeys["profile."+key] = true
			} else {
				// Variáveis temporárias vão para context scope
				availableKeys["context."+key] = true
			}
		}

		// Variáveis do profile.variables (valores atuais)
		if s.designDoc.Profile.Variables != nil {
			for key := range s.designDoc.Profile.Variables {
				// Verificar se a variável é persistente baseado na definição
				if profileVar, hasDefinition := s.designDoc.Profile.Context[key]; hasDefinition {
					if profileVar.Persist {
						availableKeys["profile."+key] = true
					} else {
						availableKeys["context."+key] = true
					}
				} else {
					// Se não há definição, assume context por padrão
					availableKeys["context."+key] = true
				}
			}
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
			// Verificar se usa escopo inválido (user. ou flow.)
			if strings.HasPrefix(v, "user.") || strings.HasPrefix(v, "flow.") {
				issues = append(issues, Issue{
					Code:     "liquid.var.invalid_scope",
					Severity: Err,
					Path:     path,
					Msg:      "Liquid variable '" + v + "' uses invalid scope. Use 'context.' or 'profile.' instead.",
				})
				continue
			}

			// Verificar se a variável está disponível
			if !availableKeys[v] {
				// Verificar se é um acesso válido mas não definido
				if strings.HasPrefix(v, "context.") || strings.HasPrefix(v, "profile.") {
					issues = append(issues, Issue{
						Code:     "liquid.var.missing_persistence",
						Severity: Err,
						Path:     path,
						Msg:      "Liquid variable '" + v + "' is not available in persisted keys/context/profile.",
					})
				} else {
					issues = append(issues, Issue{
						Code:     "liquid.var.invalid_format",
						Severity: Err,
						Path:     path,
						Msg:      "Liquid variable '" + v + "' must use 'context.' or 'profile.' prefix.",
					})
				}
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
