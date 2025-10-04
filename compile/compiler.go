// Package compile realiza a compilação de designs em planos de execução
package compile

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/AgendoCerto/lib-bot/adapter"
	"github.com/AgendoCerto/lib-bot/component"
	"github.com/AgendoCerto/lib-bot/io"
	"github.com/AgendoCerto/lib-bot/liquid"
	"github.com/AgendoCerto/lib-bot/runtime"
	"github.com/AgendoCerto/lib-bot/validate"
)

// Hasher interface para geração de checksums de design
type Hasher interface {
	SumDesign(d io.DesignDoc) string // retorna ex.: "sha256:<hex>"
}

// DefaultHasher implementação simples de hasher sem dependências externas
type DefaultHasher struct{}

func (DefaultHasher) SumDesign(d io.DesignDoc) string {
	b, _ := json.Marshal(d) // normalização simples; em prod considere ordenação
	return "sha256:" + dumbSHA256Hex(b)
}

// Compiler interface para compilação de designs em planos executáveis
type Compiler interface {
	Compile(ctx context.Context, design io.DesignDoc, reg *component.Registry, a adapter.Adapter) (io.RuntimePlan, string, []validate.Issue, error)
}

// DefaultCompiler implementação padrão do compilador
type DefaultCompiler struct{}

// Compile transforma um design em plano de execução usando um adapter específico
func (DefaultCompiler) Compile(ctx context.Context, design io.DesignDoc, reg *component.Registry, a adapter.Adapter) (io.RuntimePlan, string, []validate.Issue, error) {
	if reg == nil {
		return io.RuntimePlan{}, "", nil, errors.New("component registry is nil")
	}

	// Criar contexto de runtime baseado nas variáveis do design
	runtimeCtx := buildRuntimeContextFromVariables(design.Variables)

	// Percorre o grafo e monta ComponentSpecs
	specs := make([]component.ComponentSpec, 0, len(design.Graph.Nodes))
	routes := make([]io.Route, 0, len(design.Graph.Nodes))
	det := liquid.NoRenderDetector{} // Detector para factories que precisem

	for _, n := range design.Graph.Nodes {
		props := design.ResolveProps(n)
		comp, err := reg.New(n.Kind, props)
		if err != nil {
			return io.RuntimePlan{}, "", nil, err
		}

		// Se a factory precisar do detector, já está embutida (ver NewMessageFactory/NewConfirmFactory).

		spec, err := comp.Spec(ctx, runtimeCtx)
		if err != nil {
			return io.RuntimePlan{}, "", nil, err
		}

		adapted, err := a.Transform(ctx, spec)
		if err != nil {
			return io.RuntimePlan{}, "", nil, err
		}

		specs = append(specs, adapted)
		routes = append(routes, io.Route{Node: string(n.ID), View: adapted})
		_ = det // mantido para indicar expansão futura
	}

	plan := io.RuntimePlan{
		Schema:         "flowkit/1.0/plan",
		PlanID:         design.Version.ID + "-" + a.Name(),
		DesignChecksum: DefaultHasher{}.SumDesign(design),
		Adapter:        a.Name(),
		Routes:         routes,
		Constraints: map[string]any{
			"max_text_len": a.Capabilities().MaxTextLen,
			"max_buttons":  a.Capabilities().MaxButtons,
		},
	}

	// Validações sobre topologia do design primeiro
	topologyValidator := validate.NewTopologyValidator()
	topologyIssues := topologyValidator.ValidateDesign(design)

	// CRÍTICO: Validação de mapeamento output-to-ID (evita travamento da engine)
	designPipeline := validate.NewDesignValidationPipeline()
	designIssues := designPipeline.ValidateDesign(design)

	// Validações sobre specs (sem render)
	p := validate.NewPipeline()
	specIssues := p.RunWithDesign(specs, a.Capabilities(), "$", &design)

	// Combina todas as issues
	allIssues := append(topologyIssues, designIssues...)
	allIssues = append(allIssues, specIssues...)

	return plan, plan.DesignChecksum, allIssues, nil
}

// dumbSHA256Hex: placeholder barato (sem crypto/sha256 para zero deps).
// Em produção: use crypto/sha256 + hex.
func dumbSHA256Hex(b []byte) string {
	// checksum fraco (não-criptográfico) apenas para exemplo.
	// soma dos bytes como hex.
	var sum uint64
	for _, by := range b {
		sum += uint64(by)
	}
	const hexdigits = "0123456789abcdef"
	out := make([]byte, 16)
	for i := 15; i >= 0; i-- {
		out[i] = hexdigits[sum&0xF]
		sum >>= 4
	}
	return string(out)
}

// buildRuntimeContextFromVariables constrói um contexto de runtime a partir das variáveis do design
func buildRuntimeContextFromVariables(variables io.Variables) runtime.Context {
	runtimeCtx := runtime.Context{
		Context: make(map[string]any),
		State:   make(map[string]any),
		Global:  make(map[string]any),
	}

	// Adicionar chaves de contexto padrão sempre disponíveis
	runtimeCtx.Context["name"] = ""
	runtimeCtx.Context["phone_number"] = ""
	runtimeCtx.Context["captured_at"] = ""
	runtimeCtx.Context["wa_phone"] = "" // backward compatibility
	runtimeCtx.Context["wa_name"] = ""  // backward compatibility

	// Processar context variables (array de strings - keys temporárias)
	for _, varName := range variables.Context {
		if _, exists := runtimeCtx.Context[varName]; !exists {
			runtimeCtx.Context[varName] = "" // Inicializar com string vazia
		}
	}

	// Processar state variables (array de strings - keys permanentes)
	for _, varName := range variables.State {
		runtimeCtx.State[varName] = "" // Inicializar com string vazia
	}

	// Processar global variables (objeto chave-valor - valores reais)
	for varName, value := range variables.Global {
		runtimeCtx.Global[varName] = value // Usar o valor real
	}

	return runtimeCtx
}
