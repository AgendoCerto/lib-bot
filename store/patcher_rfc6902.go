package store

import "errors"

// Stub do Patcher RFC6902. Em produção, plugue uma lib real (ex.: github.com/evanphx/json-patch).
type RFC6902Patcher struct{}

func (RFC6902Patcher) ApplyJSONPatch(doc []byte, _ []byte) ([]byte, error) {
	// Para manter zero dependências, apenas retorna doc (no-op).
	// Troque por uma implementação real quando desejar.
	return doc, nil
}

var _ = errors.New
