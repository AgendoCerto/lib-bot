package validate

import (
	"github.com/AgendoCerto/lib-bot/adapter"
	"github.com/AgendoCerto/lib-bot/component"
)

type SizeStep struct{}

// NewSizeStep cria um novo validador de tamanho
func NewSizeStep() *SizeStep {
	return &SizeStep{}
}

func (SizeStep) Check(spec component.ComponentSpec, caps adapter.Capabilities, path string) []Issue {
	var issues []Issue
	// Se NÃO é template, podemos checar tamanho estático.
	if spec.Text != nil && !spec.Text.Template {
		if caps.MaxTextLen > 0 && spec.Text.Liquid.EstimatedStaticLen > caps.MaxTextLen {
			issues = append(issues, Issue{
				Code: "text.length.exceeded", Severity: Err,
				Path: path + ".view.text", Msg: "static length exceeds adapter limit",
			})
		}
	}
	// Se é template, deferimos a checagem ao sender (pós-render).
	if spec.Text != nil && spec.Text.Template {
		issues = append(issues, Issue{
			Code: "text.length.deferred", Severity: Warn,
			Path: path + ".view.text", Msg: "length check deferred to runtime (post-render)",
		})
	}
	return issues
}
