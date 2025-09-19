// Package store provides RFC6902 JSON Patch implementation.
package store

import (
	"context"
	"errors"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"
)

// Static errors for better error handling.
var (
	ErrInvalidPatchFormat = errors.New("invalid JSON patch format")
	ErrPatchApplyFailed   = errors.New("failed to apply JSON patch")
	ErrInvalidDocument    = errors.New("invalid JSON document")
)

// RFC6902Patcher provides RFC6902 JSON Patch implementation using github.com/evanphx/json-patch.
// This implementation follows Go best practices with proper error handling and validation.
type RFC6902Patcher struct{}

// NewRFC6902Patcher creates a new RFC6902Patcher.
func NewRFC6902Patcher() *RFC6902Patcher {
	return &RFC6902Patcher{}
}

// ApplyJSONPatch applies JSON patch operations to a document following RFC6902 specification.
// It validates both the document and patch operations before applying them.
// Returns the patched document or an error if the operation fails.
func (p *RFC6902Patcher) ApplyJSONPatch(_ context.Context, doc []byte, patchOps []byte) ([]byte, error) {
	// Handle empty patch operations - return original document
	if len(patchOps) == 0 {
		return doc, nil
	}

	// Validate input document
	if len(doc) == 0 {
		return nil, fmt.Errorf("%w: document is empty", ErrInvalidDocument)
	}

	// Parse and validate patch operations
	patch, err := jsonpatch.DecodePatch(patchOps)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPatchFormat, err)
	}

	// Apply patch operations to the document
	result, err := patch.Apply(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPatchApplyFailed, err)
	}

	return result, nil
}

// Compile-time interface implementation check.
var _ PatchApplier = (*RFC6902Patcher)(nil)
