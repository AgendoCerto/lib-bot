package validate

import (
	"lib-bot/adapter"
	"lib-bot/component"
	"lib-bot/liquid"
)

type LiquidStep struct {
	Policy liquid.Policy
	Linter liquid.Linter
}

func (s LiquidStep) Check(spec component.ComponentSpec, _ adapter.Capabilities, path string) []Issue {
	var issues []Issue
	if spec.Text != nil && spec.Text.Template {
		for _, is := range s.Linter.Lint(spec.Text.Liquid, s.Policy, path+".view.text") {
			issues = append(issues, Issue{Code: is.Code, Severity: toSev(is.Severity), Path: is.Path, Msg: is.Msg})
		}
	}
	for i := range spec.Buttons {
		if spec.Buttons[i].Label.Template {
			for _, is := range s.Linter.Lint(spec.Buttons[i].Label.Liquid, s.Policy, path+".view.buttons") {
				issues = append(issues, Issue{Code: is.Code, Severity: toSev(is.Severity), Path: is.Path, Msg: is.Msg})
			}
		}
	}
	if spec.HSM != nil {
		for i := range spec.HSM.Params {
			if spec.HSM.Params[i].Template {
				for _, is := range s.Linter.Lint(spec.HSM.Params[i].Liquid, s.Policy, path+".view.hsm.params") {
					issues = append(issues, Issue{Code: is.Code, Severity: toSev(is.Severity), Path: is.Path, Msg: is.Msg})
				}
			}
		}
		for i := range spec.HSM.Buttons {
			if spec.HSM.Buttons[i].Label.Template {
				for _, is := range s.Linter.Lint(spec.HSM.Buttons[i].Label.Liquid, s.Policy, path+".view.hsm.buttons") {
					issues = append(issues, Issue{Code: is.Code, Severity: toSev(is.Severity), Path: is.Path, Msg: is.Msg})
				}
			}
		}
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
