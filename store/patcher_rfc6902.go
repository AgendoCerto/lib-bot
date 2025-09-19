// Package store provides RFC6902 JSON Patch implementation.
package store

import (
	"context"
	"errors"
)

// Static errors for better error handling.
var (
	ErrPatchNotImplemented = errors.New("JSON patch not implemented - use a real implementation")
)

// RFC6902Patcher provides a stub implementation of JSON Patch.
// In production, replace with a real implementation (e.g., github.com/evanphx/json-patch).
type RFC6902Patcher struct{}

// NewRFC6902Patcher creates a new RFC6902Patcher.
func NewRFC6902Patcher() *RFC6902Patcher {
	return &RFC6902Patcher{}
}

// ApplyJSONPatch applies JSON patch operations to a document.
// This is a stub implementation that returns the original document.
func (p *RFC6902Patcher) ApplyJSONPatch(_ context.Context, doc []byte, patchOps []byte) ([]byte, error) {
	// To maintain zero dependencies, this returns the original document (no-op).
	// Replace with a real implementation when needed.
	if len(patchOps) > 0 {
		// If patches are provided but not implemented, return an error
		return nil, ErrPatchNotImplemented
	}

	return doc, nil
}

// Compile-time interface implementation check.
var _ PatchApplier = (*RFC6902Patcher)(nil)
