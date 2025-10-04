package component

import (
	"context"

	"github.com/AgendoCerto/lib-bot/liquid"
	"github.com/AgendoCerto/lib-bot/persistence"
	"github.com/AgendoCerto/lib-bot/runtime"
)

// UnitFinder componente para buscar e selecionar unidades próximas (spec v2.2)
type UnitFinder struct {
	radiusKmDefault      float64
	radiusKmExpandOnFail float64
	pageSize             int
	sort                 string
	showPreferredFirst   bool
	autoPersist          *persistence.Config
	det                  liquid.Detector
}

// NewUnitFinder cria nova instância
func NewUnitFinder(det liquid.Detector) *UnitFinder {
	return &UnitFinder{
		det:                  det,
		radiusKmDefault:      8.0,
		radiusKmExpandOnFail: 20.0,
		pageSize:             10,
		sort:                 "distance",
		showPreferredFirst:   true,
	}
}

func (uf *UnitFinder) Kind() string { return "unit_finder" }

// WithRadius define raios de busca
func (uf *UnitFinder) WithRadius(defaultKm, expandKm float64) *UnitFinder {
	cp := *uf
	cp.radiusKmDefault = defaultKm
	cp.radiusKmExpandOnFail = expandKm
	return &cp
}

// WithPagination define tamanho da página
func (uf *UnitFinder) WithPagination(pageSize int) *UnitFinder {
	cp := *uf
	cp.pageSize = pageSize
	return &cp
}

// WithSort define ordenação
func (uf *UnitFinder) WithSort(sort string) *UnitFinder {
	cp := *uf
	cp.sort = sort
	return &cp
}

// WithPreferred define se mostra preferida primeiro
func (uf *UnitFinder) WithPreferred(show bool) *UnitFinder {
	cp := *uf
	cp.showPreferredFirst = show
	return &cp
}

// WithAutoPersist define persistência automática
func (uf *UnitFinder) WithAutoPersist(config *persistence.Config) *UnitFinder {
	cp := *uf
	cp.autoPersist = config
	return &cp
}

// Spec gera o ComponentSpec
func (uf *UnitFinder) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	metaData := map[string]any{
		"radius_km_default":        uf.radiusKmDefault,
		"radius_km_expand_on_fail": uf.radiusKmExpandOnFail,
		"page_size":                uf.pageSize,
		"sort":                     uf.sort,
		"show_preferred_first":     uf.showPreferredFirst,
		"component_type":           "unit_finder",
		// Outputs: selected | no_results | more
	}

	if uf.autoPersist != nil {
		metaData["auto_persist"] = map[string]string{
			"scope": string(uf.autoPersist.Scope),
			"key":   uf.autoPersist.Key,
		}
	}

	return ComponentSpec{
		Kind: "unit_finder",
		Meta: metaData,
	}, nil
}

// UnitFinderFactory factory
type UnitFinderFactory struct{ det liquid.Detector }

func NewUnitFinderFactory(det liquid.Detector) *UnitFinderFactory {
	return &UnitFinderFactory{det: det}
}

func (f *UnitFinderFactory) New(_ string, props map[string]any) (Component, error) {
	uf := NewUnitFinder(f.det)

	// Radius
	if radiusDefault, ok := props["radius_km_default"].(float64); ok {
		radiusExpand, _ := props["radius_km_expand_on_fail"].(float64)
		if radiusExpand == 0 {
			radiusExpand = radiusDefault * 2
		}
		uf = uf.WithRadius(radiusDefault, radiusExpand)
	}

	// Page size
	if pageSize, ok := props["page_size"].(float64); ok {
		uf = uf.WithPagination(int(pageSize))
	} else if pageSize, ok := props["page_size"].(int); ok {
		uf = uf.WithPagination(pageSize)
	}

	// Sort
	if sort, ok := props["sort"].(string); ok {
		uf = uf.WithSort(sort)
	}

	// Preferred
	if showPreferred, ok := props["show_preferred_first"].(bool); ok {
		uf = uf.WithPreferred(showPreferred)
	}

	// Auto persist
	if autoPersistRaw, ok := props["auto_persist"].(map[string]any); ok {
		config := &persistence.Config{}
		if scope, ok := autoPersistRaw["scope"].(string); ok {
			config.Scope = persistence.Scope(scope)
		}
		if key, ok := autoPersistRaw["key"].(string); ok {
			config.Key = key
		}
		config.Enabled = true
		uf = uf.WithAutoPersist(config)
	}

	// Parse behaviors
	behavior, err := ParseBehavior(props, f.det)
	if err != nil {
		return nil, err
	}

	return &UnitFinderWithBehavior{
		unitFinder: uf,
		behavior:   behavior,
	}, nil
}

// UnitFinderWithBehavior wrapper
type UnitFinderWithBehavior struct {
	unitFinder *UnitFinder
	behavior   *ComponentBehavior
}

func (ufwb *UnitFinderWithBehavior) Kind() string {
	return ufwb.unitFinder.Kind()
}

func (ufwb *UnitFinderWithBehavior) Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) {
	spec, err := ufwb.unitFinder.Spec(ctx, rctx)
	if err != nil {
		return spec, err
	}

	spec.Behavior = ufwb.behavior
	if ufwb.unitFinder.autoPersist != nil {
		spec.Persistence = ufwb.unitFinder.autoPersist
	}

	return spec, nil
}
