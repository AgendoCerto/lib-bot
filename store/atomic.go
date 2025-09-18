package store

import (
	"encoding/json"
	"errors"
	"time"

	"lib-bot/adapter"
	"lib-bot/compile"
	"lib-bot/component"
	"lib-bot/io"
	"lib-bot/validate"
)

type Versioned struct {
	ID       string
	Status   string // development|production
	Checksum string
	Data     []byte // Design JSON normalizado
}

type Repo interface {
	GetActiveProduction(botID string) (Versioned, error)
	GetDraft(botID string) (Versioned, error)
	CommitDraft(botID string, v Versioned) error // atomic write (ETag/If-Match)
	Promote(botID, versionID string) error       // single flip
}

type Patcher interface {
	ApplyJSONPatch(doc []byte, patchOps []byte) ([]byte, error) // RFC6902
}

type Service struct {
	R Repo
	C compile.Compiler
	P Patcher
}

func (s Service) ApplyAtomic(botID string, patchOps []byte, reg componentRegistry, a adapter.Adapter) (newVersionID string, plan []byte, issues []validate.Issue, err error) {
	if s.R == nil || s.C == nil || s.P == nil {
		return "", nil, nil, errors.New("store.Service not properly configured")
	}
	draft, err := s.R.GetDraft(botID)
	if err != nil {
		return "", nil, nil, err
	}

	newDoc, err := s.P.ApplyJSONPatch(draft.Data, patchOps)
	if err != nil {
		return "", nil, nil, err
	}

	var design io.DesignDoc
	if err := json.Unmarshal(newDoc, &design); err != nil {
		return "", nil, nil, err
	}

	planAny, _, issues, err := s.C.Compile(ctxNoop{}, design, reg.registry, a)
	if err != nil {
		return "", nil, issues, err
	}

	// bloqueia commit se houver erro de validação
	for _, is := range issues {
		if is.Severity == validate.Err {
			return "", nil, issues, ErrInvalid{Issues: issues}
		}
	}

	planJSON, _ := json.Marshal(planAny)

	newVersion := Versioned{
		ID:       newULID(),
		Status:   "development",
		Checksum: "", // opcional; Compiler já retorna checksum no plan
		Data:     normalizeJSON(newDoc),
	}
	if err := s.R.CommitDraft(botID, newVersion); err != nil {
		return "", nil, issues, err
	}
	return newVersion.ID, planJSON, issues, nil
}

type ErrInvalid struct{ Issues []validate.Issue }

func (e ErrInvalid) Error() string { return "validation failed" }

// Helpers (mínimos, sem deps externas)
type ctxNoop struct{}

func (ctxNoop) Deadline() (deadline time.Time, ok bool) { return }
func (ctxNoop) Done() <-chan struct{}                   { return nil }
func (ctxNoop) Err() error                              { return nil }
func (ctxNoop) Value(key any) any                       { return nil }

func newULID() string {
	// ULID simples fakeado (timestamp base36 + nanos base36) – substitua por oklog/ulid se quiser.
	ts := time.Now().UTC().UnixNano()
	return "01" + base36(uint64(ts))
}

func base36(u uint64) string {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyz"
	if u == 0 {
		return "0"
	}
	out := []byte{}
	for u > 0 {
		out = append([]byte{chars[u%36]}, out...)
		u /= 36
	}
	return string(out)
}

func normalizeJSON(b []byte) []byte {
	var v any
	_ = json.Unmarshal(b, &v)
	nb, _ := json.Marshal(v)
	return nb
}

// componentRegistry é um adaptador leve para evitar dependência cíclica
type componentRegistry struct {
	registry *component.Registry
}
