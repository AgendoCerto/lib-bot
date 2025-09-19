// Package service fornece services especializados para diferentes responsabilidades
package service

import (
	"context"
	"time"

	"github.com/AgendoCerto/lib-bot/adapter"
	"github.com/AgendoCerto/lib-bot/adapter/whatsapp"
	"github.com/AgendoCerto/lib-bot/compile"
	"github.com/AgendoCerto/lib-bot/component"
	"github.com/AgendoCerto/lib-bot/io"
	"github.com/AgendoCerto/lib-bot/validate"
)

// ValidationService é responsável apenas por validar designs de bots
type ValidationService struct {
	registry *component.Registry
	adapters map[string]adapter.Adapter
	compiler *compile.DefaultCompiler
}

// ValidationResult representa o resultado simples de uma validação
type ValidationResult struct {
	Valid    bool             `json:"valid"`
	Plan     *io.RuntimePlan  `json:"plan,omitempty"`
	Checksum string           `json:"checksum"`
	Issues   []validate.Issue `json:"issues"`
	Duration time.Duration    `json:"duration"`
}

// NewValidationService cria um novo serviço de validação
func NewValidationService() *ValidationService {
	service := &ValidationService{
		registry: component.DefaultRegistry(),
		adapters: make(map[string]adapter.Adapter),
		compiler: &compile.DefaultCompiler{},
	}

	// Registra adapters padrão
	service.RegisterAdapter("whatsapp", whatsapp.New())

	return service
}

// RegisterAdapter registra um adapter no serviço
func (s *ValidationService) RegisterAdapter(name string, adapter adapter.Adapter) {
	s.adapters[name] = adapter
}

// Validate valida um design e retorna resultado simples
func (s *ValidationService) Validate(ctx context.Context, designJSON []byte, adapterName string) (*ValidationResult, error) {
	start := time.Now()

	// Parse do design
	codec := io.JSONCodec{}
	design, err := codec.DecodeDesign(designJSON)
	if err != nil {
		return &ValidationResult{
			Valid:    false,
			Issues:   []validate.Issue{{Code: "parse_error", Severity: "error", Msg: err.Error()}},
			Duration: time.Since(start),
		}, err
	}

	// Seleciona adapter
	if adapterName == "" {
		adapterName = "whatsapp"
	}

	adapter, exists := s.adapters[adapterName]
	if !exists {
		return &ValidationResult{
			Valid:    false,
			Issues:   []validate.Issue{{Code: "adapter_not_found", Severity: "error", Msg: "adapter not found: " + adapterName}},
			Duration: time.Since(start),
		}, nil
	}

	// Compila e valida
	plan, checksum, issues, err := s.compiler.Compile(ctx, design, s.registry, adapter)
	if err != nil {
		return &ValidationResult{
			Valid:    false,
			Issues:   []validate.Issue{{Code: "compile_error", Severity: "error", Msg: err.Error()}},
			Duration: time.Since(start),
		}, err
	}

	// Verifica se há errors críticos
	valid := true
	for _, issue := range issues {
		if issue.Severity == "error" {
			valid = false
			break
		}
	}

	result := &ValidationResult{
		Valid:    valid,
		Checksum: checksum,
		Issues:   issues,
		Duration: time.Since(start),
	}

	// Inclui plano apenas se válido
	if valid {
		result.Plan = &plan
	}

	return result, nil
}

// ValidateDesign método de conveniência que aceita io.DesignDoc
func (s *ValidationService) ValidateDesign(ctx context.Context, design io.DesignDoc, adapterName string) (*ValidationResult, error) {
	codec := io.JSONCodec{}
	designJSON, err := codec.EncodeDesign(design)
	if err != nil {
		return nil, err
	}

	return s.Validate(ctx, designJSON, adapterName)
}

// GetIssuesByPath retorna issues agrupados por path para facilitar análise
func (s *ValidationService) GetIssuesByPath(issues []validate.Issue) map[string][]validate.Issue {
	byPath := make(map[string][]validate.Issue)

	for _, issue := range issues {
		path := issue.Path
		if path == "" {
			path = "root"
		}
		byPath[path] = append(byPath[path], issue)
	}

	return byPath
}

// CountIssuesBySeverity conta issues por severidade
func (s *ValidationService) CountIssuesBySeverity(issues []validate.Issue) map[string]int {
	counts := make(map[string]int)

	for _, issue := range issues {
		counts[string(issue.Severity)]++
	}

	return counts
}
