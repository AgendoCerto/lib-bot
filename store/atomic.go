// Package store provides atomic operations for document storage and versioning.
package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Service provides atomic operations for document management.
type Service struct {
	repository Repository
	compiler   DesignCompiler
	patcher    PatchApplier
	versionGen VersionIDGenerator
	normalizer JSONNormalizer
	validator  ValidationChecker
}

// NewService creates a new service with all required dependencies.
func NewService(
	repo Repository,
	compiler DesignCompiler,
	patcher PatchApplier,
	versionGen VersionIDGenerator,
	normalizer JSONNormalizer,
	validator ValidationChecker,
) *Service {
	return &Service{
		repository: repo,
		compiler:   compiler,
		patcher:    patcher,
		versionGen: versionGen,
		normalizer: normalizer,
		validator:  validator,
	}
}

// ApplyAtomic applies JSON patch operations atomically to a bot's draft.
func (s *Service) ApplyAtomic(
	ctx context.Context,
	botID string,
	patchOps []byte,
	registry ComponentRegistry,
	adapter Adapter,
) (newVersionID string, plan []byte, issues []ValidationIssue, err error) {
	if err := s.validateDependencies(); err != nil {
		return "", nil, nil, err
	}

	draft, err := s.repository.GetDraft(ctx, botID)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to get draft: %w", err)
	}

	patchedDoc, err := s.patcher.ApplyJSONPatch(ctx, draft.Data, patchOps)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to apply patch: %w", err)
	}

	planData, issues, err := s.compileAndValidate(ctx, patchedDoc, registry, adapter)
	if err != nil {
		return "", nil, issues, err
	}

	if s.validator.HasErrors(issues) {
		return "", nil, issues, ValidationError{Issues: issues}
	}

	newVersionID, err = s.commitNewVersion(ctx, botID, patchedDoc)
	if err != nil {
		return "", nil, issues, fmt.Errorf("failed to commit version: %w", err)
	}

	return newVersionID, planData, issues, nil
}

// validateDependencies checks if all required dependencies are configured.
func (s *Service) validateDependencies() error {
	if s.repository == nil || s.compiler == nil || s.patcher == nil ||
		s.versionGen == nil || s.normalizer == nil || s.validator == nil {
		return ErrServiceNotConfigured
	}

	return nil
}

// compileAndValidate compiles the design and validates it.
func (s *Service) compileAndValidate(
	ctx context.Context,
	docData []byte,
	registry ComponentRegistry,
	adapter Adapter,
) ([]byte, []ValidationIssue, error) {
	var design interface{}
	if err := json.Unmarshal(docData, &design); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal design: %w", err)
	}

	planData, _, issues, err := s.compiler.Compile(ctx, design, registry.GetRegistry(), adapter)
	if err != nil {
		return nil, issues, fmt.Errorf("compilation failed: %w", err)
	}

	planJSON, err := json.Marshal(planData)
	if err != nil {
		return nil, issues, fmt.Errorf("failed to marshal plan: %w", err)
	}

	return planJSON, issues, nil
}

// commitNewVersion creates and commits a new version.
func (s *Service) commitNewVersion(ctx context.Context, botID string, docData []byte) (string, error) {
	newVersion := Versioned{
		ID:       s.versionGen.Generate(),
		Status:   "development",
		Checksum: "", // Optional; Compiler already returns checksum in plan
		Data:     s.normalizer.Normalize(docData),
	}

	if err := s.repository.CommitDraft(ctx, botID, newVersion); err != nil {
		return "", fmt.Errorf("failed to commit draft: %w", err)
	}

	return newVersion.ID, nil
}

// DefaultVersionIDGenerator provides default version ID generation.
type DefaultVersionIDGenerator struct{}

// Generate creates a new version ID.
func (g DefaultVersionIDGenerator) Generate() string {
	// Simple ULID-like ID (timestamp base36 + nanos base36) – replace with oklog/ulid if needed.
	timestamp := time.Now().UTC().UnixNano()
	return "01" + base36(uint64(timestamp))
}

// DefaultJSONNormalizer provides default JSON normalization.
type DefaultJSONNormalizer struct{}

// Normalize normalizes JSON data.
func (n DefaultJSONNormalizer) Normalize(data []byte) []byte {
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return data // Return original data if unmarshal fails
	}

	normalized, err := json.Marshal(value)
	if err != nil {
		return data // Return original data if marshal fails
	}

	return normalized
}

// DefaultValidationChecker provides default validation checking.
type DefaultValidationChecker struct{}

// HasErrors checks if validation issues contain errors.
func (c DefaultValidationChecker) HasErrors(issues []ValidationIssue) bool {
	for _, issue := range issues {
		if issue.Severity == "error" {
			return true
		}
	}

	return false
}

// base36 converts a uint64 to base36 string.
func base36(number uint64) string {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyz"
	if number == 0 {
		return "0"
	}

	var result []byte
	for number > 0 {
		result = append([]byte{chars[number%36]}, result...)
		number /= 36
	}

	return string(result)
}
