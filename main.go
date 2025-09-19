package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"lib-bot/adapter"
	"lib-bot/adapter/whatsapp"
	"lib-bot/compile"
	"lib-bot/component"
	"lib-bot/io"
	rf "lib-bot/reactflow"
	"lib-bot/validate"
)

func main() {
	// Configuração de flags de linha de comando
	in := flag.String("in", "", "Caminho do arquivo Design JSON (opcional; usa exemplo se vazio)")
	out := flag.String("out", "plan", "Tipo de saída: plan | plan-full | reactflow | reactflow-auto-v | reactflow-auto-h")
	outFile := flag.String("outfile", "", "Arquivo de saída (opcional; se vazio, imprime no stdout)")
	adapterName := flag.String("adapter", "whatsapp", "Adapter: whatsapp (por enquanto)")
	pretty := flag.Bool("pretty", true, "Imprimir JSON com identação")
	flag.Parse()

	// 1) Carrega Design JSON (arquivo ou exemplo embutido)
	var designJSON []byte
	var err error
	if *in != "" {
		designJSON, err = os.ReadFile(*in) // <- substitui ioutil.ReadFile
		must(err)
	} else {
		designJSON = []byte(sampleDesignJSON)
	}
	codec := io.JSONCodec{}
	design, err := codec.DecodeDesign(designJSON)
	must(err)

	// 2) Registry de componentes (parse-only do Liquid; sem render)
	reg := component.DefaultRegistry()

	// 3) Adapter real (usa a interface adapter.Adapter)
	a := selectAdapter(*adapterName)

	// 4) Gera nome do arquivo de saída se não especificado
	finalOutFile := *outFile
	if finalOutFile == "" && *in != "" && *out != "plan" {
		finalOutFile = generateOutputFileName(*in, *out)
	}

	// 5) Processa saída conforme tipo solicitado
	switch *out {
	case "plan":
		doPlan(design, reg, a, *pretty, finalOutFile, false)
	case "plan-full":
		doPlan(design, reg, a, *pretty, finalOutFile, true)
	case "reactflow":
		doReactFlow(design, *pretty, finalOutFile)
	case "reactflow-auto-v":
		doReactFlowAutoVertical(design, *pretty, finalOutFile)
	case "reactflow-auto-h":
		doReactFlowAutoHorizontal(design, *pretty, finalOutFile)
	default:
		log.Fatalf("valor inválido para -out: %q (use: plan | plan-full | reactflow | reactflow-auto-v | reactflow-auto-h)", *out)
	}
}

// ValidationResult estrutura o resultado completo da validação
type ValidationResult struct {
	Plan        io.RuntimePlan    `json:"plan"`
	Checksum    string            `json:"checksum"`
	Issues      []validate.Issue  `json:"issues"`
	Summary     ValidationSummary `json:"summary"`
	Persistence PersistenceInfo   `json:"persistence,omitempty"`
}

// ValidationSummary resume estatísticas da validação
type ValidationSummary struct {
	TotalIssues int `json:"total_issues"`
	Errors      int `json:"errors"`
	Warnings    int `json:"warnings"`
	Infos       int `json:"infos"`
}

// PersistenceInfo contém informações sobre persistência no design
type PersistenceInfo struct {
	ContextKeys    []string `json:"context_keys,omitempty"`
	ProfileKeys    []string `json:"profile_keys,omitempty"`
	HasPersistence bool     `json:"has_persistence"`
}

// doPlan compila o design em um plano de execução usando o adapter especificado
func doPlan(design io.DesignDoc, reg *component.Registry, a adapter.Adapter, pretty bool, outFile string, fullOutput bool) {
	comp := compile.DefaultCompiler{}

	// O compiler já inclui todas as validações: design + component + design pipeline
	plan, checksum, allIssues, err := comp.Compile(context.Background(), design, reg, a)
	must(err)

	// Cria resultado estruturado
	result := ValidationResult{
		Plan:     plan,
		Checksum: checksum,
		Issues:   allIssues,
		Summary: ValidationSummary{
			TotalIssues: len(allIssues),
			Errors:      countBySeverity(allIssues, "error"),
			Warnings:    countBySeverity(allIssues, "warn"),
			Infos:       countBySeverity(allIssues, "info"),
		},
		Persistence: extractPersistenceInfo(design),
	}

	// Debug info em stderr
	fmt.Fprintf(os.Stderr, "Design checksum: %s\n", checksum)
	fmt.Fprintf(os.Stderr, "Validation: %d issues (%d errors, %d warnings)\n",
		result.Summary.TotalIssues, result.Summary.Errors, result.Summary.Warnings)

	if len(allIssues) > 0 {
		fmt.Fprintln(os.Stderr, "Validation issues:")
		for _, is := range allIssues {
			fmt.Fprintf(os.Stderr, " - [%s] %s (%s)\n", is.Severity, is.Msg, is.Path)
		}
	}

	// Output: plan ou resultado completo
	if fullOutput {
		// Saída completa com validação
		writeJSON(result, pretty, outFile)
	} else {
		// Apenas o plano (compatibilidade)
		writeJSON(plan, pretty, outFile)
	}
}

// doReactFlow converte o design para formato React Flow
func doReactFlow(design io.DesignDoc, pretty bool, outFile string) {
	nodes, edges := rf.DesignToReactFlow(design)
	payload := map[string]any{
		"nodes": nodes,
		"edges": edges,
	}
	writeJSON(payload, pretty, outFile)
}

// doReactFlowAutoVertical converte para React Flow com auto-layout vertical
func doReactFlowAutoVertical(design io.DesignDoc, pretty bool, outFile string) {
	nodes, edges := rf.ApplyAutoLayoutVertical(design)
	payload := map[string]any{
		"nodes": nodes,
		"edges": edges,
		"layout": map[string]any{
			"direction": "vertical",
			"applied":   true,
		},
	}
	writeJSON(payload, pretty, outFile)
}

// doReactFlowAutoHorizontal converte para React Flow com auto-layout horizontal
func doReactFlowAutoHorizontal(design io.DesignDoc, pretty bool, outFile string) {
	nodes, edges := rf.ApplyAutoLayoutHorizontal(design)
	payload := map[string]any{
		"nodes": nodes,
		"edges": edges,
		"layout": map[string]any{
			"direction": "horizontal",
			"applied":   true,
		},
	}
	writeJSON(payload, pretty, outFile)
}

// generateOutputFileName gera o nome do arquivo de saída baseado no arquivo de entrada
func generateOutputFileName(inputFile, outputType string) string {
	// Remove extensão do arquivo de entrada
	base := strings.TrimSuffix(inputFile, filepath.Ext(inputFile))

	switch outputType {
	case "plan-full":
		return base + "-plan-full.json"
	case "reactflow":
		return base + "-reactflow.json"
	case "reactflow-auto-v":
		return base + "-reactflow-vertical.json"
	case "reactflow-auto-h":
		return base + "-reactflow-horizontal.json"
	default:
		return base + "-" + outputType + ".json"
	}
}

// writeJSON escreve JSON para arquivo ou stdout
func writeJSON(v any, pretty bool, outFile string) {
	var b []byte
	var err error
	if pretty {
		b, err = json.MarshalIndent(v, "", "  ")
	} else {
		b, err = json.Marshal(v)
	}
	must(err)

	if outFile != "" {
		// Escreve para arquivo
		err = os.WriteFile(outFile, b, 0o644)
		must(err)
		fmt.Fprintf(os.Stderr, "Arquivo gerado: %s\n", outFile)
	} else {
		// Escreve para stdout
		fmt.Println(string(b))
	}
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// countBySeverity conta issues por severidade
func countBySeverity(issues []validate.Issue, severity string) int {
	count := 0
	for _, issue := range issues {
		if string(issue.Severity) == severity {
			count++
		}
	}
	return count
}

// extractPersistenceInfo extrai informações de persistência do design
func extractPersistenceInfo(design io.DesignDoc) PersistenceInfo {
	info := PersistenceInfo{
		ContextKeys: make([]string, 0),
		ProfileKeys: make([]string, 0),
	}

	// Coleta keys do profile.context
	for key := range design.Profile.Context {
		info.ContextKeys = append(info.ContextKeys, key)
	}

	// Analisa props para encontrar configurações de persistência
	for _, props := range design.Props {
		if propsMap, ok := props.(map[string]any); ok {
			if persistenceConfig, err := component.ParsePersistence(propsMap); err == nil && persistenceConfig != nil && persistenceConfig.Enabled {
				info.HasPersistence = true
				key := persistenceConfig.Key
				if key != "" {
					if persistenceConfig.Scope == "context" {
						info.ContextKeys = append(info.ContextKeys, key)
					} else if persistenceConfig.Scope == "profile" {
						info.ProfileKeys = append(info.ProfileKeys, key)
					}
				}
			}
		}
	}

	return info
}

// --- seleção de adapter usando a interface real ---

func selectAdapter(name string) adapter.Adapter {
	switch name {
	case "whatsapp":
		return whatsapp.New()
	default:
		log.Fatalf("adapter desconhecido: %s (suportado: whatsapp)", name)
		return nil
	}
}

// --- exemplo de Design JSON ---

const sampleDesignJSON = `{
  "schema": "flowkit/1.0",
  "bot": { "id": "bot_01H", "channels": ["whatsapp:55119..."] },
  "version": { "id": "01JPLAN", "status": "development" },
  "entries": [{ "kind": "global_start", "target": "n_msg" }],
  "graph": {
    "nodes": [
      { "id": "n_msg", "kind": "message", "props_ref": "p_welcome" },
      { "id": "n_conf", "kind": "confirm", "props_ref": "p_confirm" }
    ],
    "edges": [
      { "from": "n_msg", "to": "n_conf", "label": "next", "priority": 1 }
    ]
  },
  "props": {
    "p_welcome": { "text": "Olá {{ user.name | default: \"cliente\" }}!" },
    "p_confirm": {
      "title": "Confirmar agendamento?",
      "positive": "Sim",
      "negative": "Não",
      "hsm_ref": {
        "id": "agc_confirm_v1",
        "locale": "pt_BR",
        "params": ["{{ user.name }}","{{ flow.slot_time }}"],
        "buttons": [
          {"label":"Confirmar","kind":"reply","data":"yes"},
          {"label":"Alterar","kind":"reply","data":"no"}
        ],
        "policy": "fallback_to_text"
      }
    }
  }
}`
