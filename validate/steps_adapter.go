package validate

import (
	"lib-bot/adapter"
	"lib-bot/component"
)

type AdapterSupportStep struct{}

func (AdapterSupportStep) Check(spec component.ComponentSpec, caps adapter.Capabilities, path string) []Issue {
	var issues []Issue
	// Exemplo: HSM sem suporte
	if spec.HSM != nil && !caps.SupportsHSM {
		issues = append(issues, Issue{
			Code: "adapter.hsm.unsupported", Severity: Err,
			Path: path + ".view.hsm", Msg: "HSM not supported by adapter",
		})
	}
	// Botões excedendo limite (mesmo que adapter vá truncar, avisamos)
	if caps.MaxButtons > 0 && len(spec.Buttons) > caps.MaxButtons {
		issues = append(issues, Issue{
			Code: "adapter.buttons.exceeded", Severity: Warn,
			Path: path + ".view.buttons", Msg: "buttons exceed adapter limit; will be truncated",
		})
	}
	return issues
}
