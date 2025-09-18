package validate

import (
	"lib-bot/adapter"
	"lib-bot/component"
)

type Severity string

const (
	Info Severity = "info"
	Warn Severity = "warn"
	Err  Severity = "error"
)

type Issue struct {
	Code     string   `json:"code"`
	Severity Severity `json:"severity"`
	Path     string   `json:"path"`
	Msg      string   `json:"msg"`
}

type Step interface {
	Check(spec component.ComponentSpec, caps adapter.Capabilities, path string) []Issue
}

type Pipeline interface {
	Run(specs []component.ComponentSpec, caps adapter.Capabilities, basePath string) []Issue
}

type DefaultPipeline struct{ steps []Step }

// NewPipeline cria um pipeline de validação completo
func NewPipeline() Pipeline {
	return &DefaultPipeline{
		steps: []Step{
			NewLiquidStep(),
			NewTopologyStep(),
			NewSizeStep(),
			NewAdapterStep(),
		},
	}
}

func (p *DefaultPipeline) Run(specs []component.ComponentSpec, caps adapter.Capabilities, basePath string) []Issue {
	var all []Issue
	for i, s := range specs {
		prefix := basePath + ".routes[" + itoa(i) + "]"
		for _, st := range p.steps {
			all = append(all, st.Check(s, caps, prefix)...)
		}
	}
	return all
}

func itoa(i int) string {
	// evitar dependência de strconv pra manter zero deps (mas você pode usar strconv)
	// simples o suficiente:
	if i == 0 {
		return "0"
	}
	d := []byte{}
	for i > 0 {
		d = append([]byte{byte('0' + i%10)}, d...)
		i /= 10
	}
	return string(d)
}
