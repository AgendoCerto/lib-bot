package hsm

// Reference e Catalog são usados em validação (design-time).
type Reference struct {
	ID          string
	Locale      string
	Namespace   string
	ParamsArity int
	ButtonKinds []string // tipos/posições aprovadas (simples)
}

type Catalog interface {
	Exists(id, locale string) bool
	Lookup(id, locale string) (Reference, bool)
}
