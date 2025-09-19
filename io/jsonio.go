// Package io gerencia a serialização/deserialização de designs e planos de execução
package io

import (
	"encoding/json"

	"github.com/AgendoCerto/lib-bot/flow"
)

// Codec define interface para codificação/decodificação de documentos
type Codec interface {
	DecodeDesign([]byte) (DesignDoc, error) // Deserializa design JSON
	EncodeDesign(DesignDoc) ([]byte, error) // Serializa design para JSON
	DecodePlan([]byte) (RuntimePlan, error) // Deserializa plano de execução
	EncodePlan(RuntimePlan) ([]byte, error) // Serializa plano de execução
}

// JSONCodec implementa Codec usando JSON padrão
type JSONCodec struct{}

func (JSONCodec) DecodeDesign(b []byte) (DesignDoc, error) {
	var d DesignDoc
	return d, json.Unmarshal(b, &d)
}
func (JSONCodec) EncodeDesign(d DesignDoc) ([]byte, error) { return json.Marshal(d) }
func (JSONCodec) DecodePlan(b []byte) (RuntimePlan, error) {
	var p RuntimePlan
	return p, json.Unmarshal(b, &p)
}
func (JSONCodec) EncodePlan(p RuntimePlan) ([]byte, error) { return json.Marshal(p) }

// DesignDoc representa um documento de design editável (formato de entrada)
type DesignDoc struct {
	Schema  string         `json:"schema"`  // Versão do schema (ex: "flowkit/1.0")
	Bot     Bot            `json:"bot"`     // Informações do bot
	Version Version        `json:"version"` // Versão do fluxo
	Entries []flow.Entry   `json:"entries"` // Pontos de entrada do fluxo
	Profile Profile        `json:"profile"` // Configurações de perfil e contexto global
	Graph   Graph          `json:"graph"`   // Grafo de nós e arestas
	Props   map[string]any `json:"props"`   // Propriedades compartilhadas/templates
}

// Profile contém configurações de perfil e contexto global
type Profile struct {
	Context map[string]ProfileVariable `json:"context"` // Variáveis de contexto global
}

// ProfileVariable define uma variável do perfil/contexto
type ProfileVariable struct {
	Type     string `json:"type"`               // Tipo da variável (string, int, bool, etc.)
	Default  string `json:"default,omitempty"`  // Valor padrão
	Persist  bool   `json:"persist,omitempty"`  // Se deve persistir entre sessões
	Required bool   `json:"required,omitempty"` // Se é obrigatória
}

// Bot contém metadados do bot
type Bot struct {
	ID       string   `json:"id"`       // Identificador único do bot
	Channels []string `json:"channels"` // Canais suportados (whatsapp, telegram, etc.)
}

// Version contém informações da versão do fluxo
type Version struct {
	ID     string `json:"id"`     // Identificador da versão
	Status string `json:"status"` // Status: "development" ou "production"
}

// Graph encapsula os nós e arestas do fluxo (redefinido do package flow para JSON)
type Graph struct {
	Nodes []flow.Node `json:"nodes"` // Lista de nós do fluxo
	Edges []flow.Edge `json:"edges"` // Lista de conexões entre nós
}

// ResolveProps resolve referências de propriedades para um nó específico
// Se o nó tem PropsRef, busca em DesignDoc.Props; senão usa Props inline
func (d DesignDoc) ResolveProps(n flow.Node) map[string]any {
	if n.PropsRef != "" {
		if p, ok := d.Props[n.PropsRef]; ok {
			if m, _ := p.(map[string]any); m != nil {
				return m
			}
		}
	}
	return n.Props
}

// RuntimePlan representa um plano compilado pronto para execução
type RuntimePlan struct {
	Schema         string         `json:"schema"`                // Versão do schema
	PlanID         string         `json:"plan_id"`               // ID único do plano
	DesignChecksum string         `json:"design_checksum"`       // Checksum do design original
	Adapter        string         `json:"adapter"`               // Adapter utilizado (whatsapp, etc.)
	Routes         []Route        `json:"routes"`                // Rotas compiladas
	Constraints    map[string]any `json:"constraints,omitempty"` // Restrições do adapter
}

// Route representa uma rota compilada para um nó específico
type Route struct {
	Node string `json:"node"` // ID do nó
	View any    `json:"view"` // ComponentSpec serializado pelo adapter
}
