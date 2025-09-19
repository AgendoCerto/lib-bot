// Package service - DesignService para operações CRUD de designs
package service

import (
	"context"
	"encoding/json"
	"fmt"

	"lib-bot/flow"
	"lib-bot/io"
)

// DesignService é responsável por operações CRUD em designs de bots
type DesignService struct {
	validationService *ValidationService
}

// PatchOperation representa uma operação de patch RFC 6902
type PatchOperation struct {
	Op    string      `json:"op"`              // add, remove, replace, move, copy, test
	Path  string      `json:"path"`            // JSON pointer path
	Value interface{} `json:"value,omitempty"` // valor para add/replace
	From  string      `json:"from,omitempty"`  // path de origem para move/copy
}

// NewDesignService cria um novo serviço de design
func NewDesignService(validationService *ValidationService) *DesignService {
	return &DesignService{
		validationService: validationService,
	}
}

// Create cria um novo design e retorna design validado
func (s *DesignService) Create(ctx context.Context, designJSON []byte, adapterName string) (io.DesignDoc, error) {
	// Parse do design
	codec := io.JSONCodec{}
	design, err := codec.DecodeDesign(designJSON)
	if err != nil {
		return io.DesignDoc{}, fmt.Errorf("erro ao decodificar design: %w", err)
	}

	// Valida o design
	result, err := s.validationService.Validate(ctx, designJSON, adapterName)
	if err != nil {
		return io.DesignDoc{}, fmt.Errorf("erro na validação: %w", err)
	}

	if !result.Valid {
		return io.DesignDoc{}, fmt.Errorf("design inválido: %d issues encontrados", len(result.Issues))
	}

	return design, nil
}

// Update aplica patches a um design existente
func (s *DesignService) Update(ctx context.Context, currentDesign io.DesignDoc, patches []PatchOperation, adapterName string) (io.DesignDoc, error) {
	if len(patches) == 0 {
		return currentDesign, nil
	}

	// Converte design para JSON para aplicar patches
	codec := io.JSONCodec{}
	designJSON, err := codec.EncodeDesign(currentDesign)
	if err != nil {
		return io.DesignDoc{}, fmt.Errorf("erro ao codificar design: %w", err)
	}

	// Aplica patches usando JSON
	updatedJSON, err := s.applyPatches(designJSON, patches)
	if err != nil {
		return io.DesignDoc{}, fmt.Errorf("erro ao aplicar patches: %w", err)
	}

	// Parse do design atualizado
	updatedDesign, err := codec.DecodeDesign(updatedJSON)
	if err != nil {
		return io.DesignDoc{}, fmt.Errorf("erro ao decodificar design atualizado: %w", err)
	}

	// Valida o design atualizado
	result, err := s.validationService.Validate(ctx, updatedJSON, adapterName)
	if err != nil {
		return io.DesignDoc{}, fmt.Errorf("erro na validação: %w", err)
	}

	if !result.Valid {
		return io.DesignDoc{}, fmt.Errorf("design atualizado inválido: %d issues encontrados", len(result.Issues))
	}

	return updatedDesign, nil
}

// AddNode adiciona um novo nó ao design
func (s *DesignService) AddNode(ctx context.Context, design io.DesignDoc, node flow.Node, adapterName string) (io.DesignDoc, error) {
	patches := []PatchOperation{
		{
			Op:    "add",
			Path:  "/graph/nodes/-",
			Value: node,
		},
	}

	return s.Update(ctx, design, patches, adapterName)
}

// RemoveNode remove um nó do design
func (s *DesignService) RemoveNode(ctx context.Context, design io.DesignDoc, nodeID string, adapterName string) (io.DesignDoc, error) {
	// Encontra índice do nó
	nodeIndex := -1
	for i, node := range design.Graph.Nodes {
		if string(node.ID) == nodeID {
			nodeIndex = i
			break
		}
	}

	if nodeIndex == -1 {
		return io.DesignDoc{}, fmt.Errorf("nó %s não encontrado", nodeID)
	}

	patches := []PatchOperation{
		{
			Op:   "remove",
			Path: fmt.Sprintf("/graph/nodes/%d", nodeIndex),
		},
	}

	// Remove edges relacionadas
	edgeRemovePatches := s.getEdgeRemovalPatches(design, nodeID)
	patches = append(patches, edgeRemovePatches...)

	return s.Update(ctx, design, patches, adapterName)
}

// UpdateNode atualiza um nó existente
func (s *DesignService) UpdateNode(ctx context.Context, design io.DesignDoc, nodeID string, updatedNode flow.Node, adapterName string) (io.DesignDoc, error) {
	// Encontra índice do nó
	nodeIndex := -1
	for i, node := range design.Graph.Nodes {
		if string(node.ID) == nodeID {
			nodeIndex = i
			break
		}
	}

	if nodeIndex == -1 {
		return io.DesignDoc{}, fmt.Errorf("nó %s não encontrado", nodeID)
	}

	patches := []PatchOperation{
		{
			Op:    "replace",
			Path:  fmt.Sprintf("/graph/nodes/%d", nodeIndex),
			Value: updatedNode,
		},
	}

	return s.Update(ctx, design, patches, adapterName)
}

// AddEdge adiciona uma nova edge ao design
func (s *DesignService) AddEdge(ctx context.Context, design io.DesignDoc, edge flow.Edge, adapterName string) (io.DesignDoc, error) {
	patches := []PatchOperation{
		{
			Op:    "add",
			Path:  "/graph/edges/-",
			Value: edge,
		},
	}

	return s.Update(ctx, design, patches, adapterName)
}

// RemoveEdge remove uma edge do design
func (s *DesignService) RemoveEdge(ctx context.Context, design io.DesignDoc, edgeID string, adapterName string) (io.DesignDoc, error) {
	// Encontra índice da edge
	edgeIndex := -1
	for i, edge := range design.Graph.Edges {
		if string(edge.From)+"-"+string(edge.To) == edgeID {
			edgeIndex = i
			break
		}
	}

	if edgeIndex == -1 {
		return io.DesignDoc{}, fmt.Errorf("edge %s não encontrada", edgeID)
	}

	patches := []PatchOperation{
		{
			Op:   "remove",
			Path: fmt.Sprintf("/graph/edges/%d", edgeIndex),
		},
	}

	return s.Update(ctx, design, patches, adapterName)
}

// UpdateProfile atualiza informações do perfil
func (s *DesignService) UpdateProfile(ctx context.Context, design io.DesignDoc, profile io.Profile, adapterName string) (io.DesignDoc, error) {
	patches := []PatchOperation{
		{
			Op:    "replace",
			Path:  "/profile",
			Value: profile,
		},
	}

	return s.Update(ctx, design, patches, adapterName)
}

// AddProp adiciona uma nova propriedade ao design
func (s *DesignService) AddProp(ctx context.Context, design io.DesignDoc, propID string, prop map[string]interface{}, adapterName string) (io.DesignDoc, error) {
	patches := []PatchOperation{
		{
			Op:    "add",
			Path:  fmt.Sprintf("/props/%s", propID),
			Value: prop,
		},
	}

	return s.Update(ctx, design, patches, adapterName)
}

// Validate valida um design usando o ValidationService
func (s *DesignService) Validate(ctx context.Context, design io.DesignDoc, adapterName string) (*ValidationResult, error) {
	return s.validationService.ValidateDesign(ctx, design, adapterName)
}

// applyPatches aplica patches RFC 6902 a um JSON (implementação simplificada)
func (s *DesignService) applyPatches(jsonData []byte, patches []PatchOperation) ([]byte, error) {
	// Deserializa JSON
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	// Aplica cada patch (implementação básica)
	for _, patch := range patches {
		var err error
		data, err = s.applyPatch(data, patch)
		if err != nil {
			return nil, fmt.Errorf("erro ao aplicar patch %s: %w", patch.Op, err)
		}
	}

	// Serializa resultado
	return json.Marshal(data)
}

// applyPatch aplica um único patch (implementação básica para demonstração)
func (s *DesignService) applyPatch(data interface{}, patch PatchOperation) (interface{}, error) {
	// Esta é uma implementação simplificada
	// Em produção, usar biblioteca como github.com/evanphx/json-patch

	switch patch.Op {
	case "add":
		return s.applyAdd(data, patch.Path, patch.Value)
	case "remove":
		return s.applyRemove(data, patch.Path)
	case "replace":
		return s.applyReplace(data, patch.Path, patch.Value)
	case "test":
		return s.applyTest(data, patch.Path, patch.Value)
	default:
		return nil, fmt.Errorf("operação não suportada: %s", patch.Op)
	}
}

// Implementações básicas dos patches (para demonstração)
func (s *DesignService) applyAdd(data interface{}, path string, value interface{}) (interface{}, error) {
	// Implementação simplificada - em produção usar lib RFC 6902
	return data, nil
}

func (s *DesignService) applyRemove(data interface{}, path string) (interface{}, error) {
	// Implementação simplificada - em produção usar lib RFC 6902
	return data, nil
}

func (s *DesignService) applyReplace(data interface{}, path string, value interface{}) (interface{}, error) {
	// Implementação simplificada - em produção usar lib RFC 6902
	return data, nil
}

func (s *DesignService) applyTest(data interface{}, path string, value interface{}) (interface{}, error) {
	// Implementação simplificada - em produção usar lib RFC 6902
	return data, nil
}

// getEdgeRemovalPatches gera patches para remover edges relacionadas a um nó
func (s *DesignService) getEdgeRemovalPatches(design io.DesignDoc, nodeID string) []PatchOperation {
	var patches []PatchOperation

	// Remove edges que conectam ao nó (em ordem reversa para não afetar índices)
	for i := len(design.Graph.Edges) - 1; i >= 0; i-- {
		edge := design.Graph.Edges[i]
		if string(edge.From) == nodeID || string(edge.To) == nodeID {
			patches = append(patches, PatchOperation{
				Op:   "remove",
				Path: fmt.Sprintf("/graph/edges/%d", i),
			})
		}
	}

	return patches
}
