// CLI para conversão de designs em planos de execução ou formato React Flow
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"lib-bot/adapter"
	"lib-bot/adapter/whatsapp"
	"lib-bot/compile"
	"lib-bot/component"
	"lib-bot/io"
	"lib-bot/liquid"
	rf "lib-bot/reactflow"
)

func main() {
	// Configuração de flags de linha de comando
	in := flag.String("in", "", "Caminho do arquivo Design JSON (opcional; usa exemplo se vazio)")
	out := flag.String("out", "plan", "Tipo de saída: plan | reactflow | reactflow-auto-v | reactflow-auto-h")
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
	reg := component.NewRegistry()
	det := liquid.NoRenderDetector{}
	reg.Register("message", component.NewMessageFactory(det))
	reg.Register("confirm", component.NewConfirmFactory(det))
	reg.Register("buttons", component.NewButtonsFactory(det))
	reg.Register("listpicker", component.NewListPickerFactory(det))
	reg.Register("media", component.NewMediaFactory(det))
	reg.Register("carousel", component.NewCarouselFactory(det))

	// 3) Adapter real (usa a interface adapter.Adapter)
	a := selectAdapter(*adapterName)

	// 4) Processa saída conforme tipo solicitado
	switch *out {
	case "plan":
		doPlan(design, reg, a, *pretty)
	case "reactflow":
		doReactFlow(design, *pretty)
	case "reactflow-auto-v":
		doReactFlowAutoVertical(design, *pretty)
	case "reactflow-auto-h":
		doReactFlowAutoHorizontal(design, *pretty)
	default:
		log.Fatalf("valor inválido para -out: %q (use: plan | reactflow | reactflow-auto-v | reactflow-auto-h)", *out)
	}
}

// doPlan compila o design em um plano de execução usando o adapter especificado
func doPlan(design io.DesignDoc, reg *component.Registry, a adapter.Adapter, pretty bool) {
	comp := compile.DefaultCompiler{}
	plan, checksum, issues, err := comp.Compile(context.Background(), design, reg, a)
	must(err)

	// Informações de debug em stderr
	fmt.Fprintf(os.Stderr, "Design checksum: %s\n", checksum)
	if len(issues) > 0 {
		fmt.Fprintln(os.Stderr, "Validation issues:")
		for _, is := range issues {
			fmt.Fprintf(os.Stderr, " - [%s] %s (%s)\n", is.Severity, is.Msg, is.Path)
		}
	}

	// Plano em stdout
	printJSON(plan, pretty)
}

// doReactFlow converte o design para formato React Flow
func doReactFlow(design io.DesignDoc, pretty bool) {
	nodes, edges := rf.DesignToReactFlow(design)
	payload := map[string]any{
		"nodes": nodes,
		"edges": edges,
	}
	printJSON(payload, pretty)
}

// doReactFlowAutoVertical converte para React Flow com auto-layout vertical
func doReactFlowAutoVertical(design io.DesignDoc, pretty bool) {
	nodes, edges := rf.ApplyAutoLayoutVertical(design)
	payload := map[string]any{
		"nodes": nodes,
		"edges": edges,
		"layout": map[string]any{
			"direction": "vertical",
			"applied":   true,
		},
	}
	printJSON(payload, pretty)
}

// doReactFlowAutoHorizontal converte para React Flow com auto-layout horizontal
func doReactFlowAutoHorizontal(design io.DesignDoc, pretty bool) {
	nodes, edges := rf.ApplyAutoLayoutHorizontal(design)
	payload := map[string]any{
		"nodes": nodes,
		"edges": edges,
		"layout": map[string]any{
			"direction": "horizontal",
			"applied":   true,
		},
	}
	printJSON(payload, pretty)
}

func printJSON(v any, pretty bool) {
	var b []byte
	var err error
	if pretty {
		b, err = json.MarshalIndent(v, "", "  ")
	} else {
		b, err = json.Marshal(v)
	}
	must(err)
	fmt.Println(string(b))
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
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
