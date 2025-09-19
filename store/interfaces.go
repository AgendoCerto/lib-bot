// Package store provides interfaces for data storage and versioning.
package store

import (
	"context"
)

// Reader provides read access to versioned data.
type Reader interface {
	GetActiveProduction(ctx context.Context, botID string) (Versioned, error)
	GetDraft(ctx context.Context, botID string) (Versioned, error)
}

// Writer provides write access to versioned data.
type Writer interface {
	CommitDraft(ctx context.Context, botID string, version Versioned) error
	Promote(ctx context.Context, botID, versionID string) error
}

// Repository combines read and write access to versioned data.
type Repository interface {
	Reader
	Writer
}

// PatchApplier applies JSON patches to documents.
type PatchApplier interface {
	ApplyJSONPatch(ctx context.Context, doc []byte, patchOps []byte) ([]byte, error)
}

// DesignCompiler compiles design documents into execution plans.
type DesignCompiler interface {
	Compile(ctx context.Context, design interface{}, registry interface{}, adapter interface{}) (interface{}, string, []ValidationIssue, error)
}

// ComponentRegistry provides component registry functionality.
type ComponentRegistry interface {
	GetRegistry() interface{}
}

// Adapter provides adapter functionality.
type Adapter interface {
	GetCapabilities() []string
}

// VersionIDGenerator generates version identifiers.
type VersionIDGenerator interface {
	Generate() string
}

// JSONNormalizer normalizes JSON documents.
type JSONNormalizer interface {
	Normalize(data []byte) []byte
}

// ValidationChecker checks validation issues for errors.
type ValidationChecker interface {
	HasErrors(issues []ValidationIssue) bool
}

// ValidationIssue represents a validation issue.
type ValidationIssue struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Path     string `json:"path"`
}
