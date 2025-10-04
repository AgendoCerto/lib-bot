package component

import (
	"context"

	"github.com/AgendoCerto/lib-bot/liquid"
	"github.com/AgendoCerto/lib-bot/runtime"
)

// LocationCapture componente para capturar localização do usuário (spec v2.2)
type LocationCapture struct {
	modes          []string // Modos de captura: use_last, share_location, type_address
	requireConfirm bool     // Requer confirmação da localização
	det            liquid.Detector
}

// NewLocationCapture cria nova instância
func NewLocationCapture(det liquid.Detector) *LocationCapture {
	return &LocationCapture{
		det:   det,
		modes: []string{"use_last", "share_location", "type_address"},
	}
}

func (lc *LocationCapture) Kind() string { return "location_capture" }

// WithModes define os modos de captura permitidos
func (lc *LocationCapture) WithModes(modes []string) *LocationCapture {
	cp := *lc
	cp.modes = modes
	return &cp
}

// WithConfirmation define se requer confirmação
func (lc *LocationCapture) WithConfirmation(require bool) *LocationCapture {
	cp := *lc
	cp.requireConfirm = require
	return &cp
}

// Spec gera o ComponentSpec
func (lc *LocationCapture) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	metaData := map[string]any{
		"modes":           lc.modes,
		"require_confirm": lc.requireConfirm,
		"component_type":  "location_capture",
		// Outputs: captured | invalid | timeout
	}

	return ComponentSpec{
		Kind: "location_capture",
		Meta: metaData,
	}, nil
}

// LocationCaptureFactory factory
type LocationCaptureFactory struct{ det liquid.Detector }

func NewLocationCaptureFactory(det liquid.Detector) *LocationCaptureFactory {
	return &LocationCaptureFactory{det: det}
}

func (f *LocationCaptureFactory) New(_ string, props map[string]any) (Component, error) {
	lc := NewLocationCapture(f.det)

	// Modos
	if modesRaw, ok := props["modes"].([]any); ok {
		modes := make([]string, 0, len(modesRaw))
		for _, m := range modesRaw {
			if str, ok := m.(string); ok {
				modes = append(modes, str)
			}
		}
		if len(modes) > 0 {
			lc = lc.WithModes(modes)
		}
	}

	// Confirmação
	if requireConfirm, ok := props["require_confirm"].(bool); ok {
		lc = lc.WithConfirmation(requireConfirm)
	}

	// Parse behaviors
	behavior, err := ParseBehavior(props, f.det)
	if err != nil {
		return nil, err
	}

	return &LocationCaptureWithBehavior{
		locationCapture: lc,
		behavior:        behavior,
	}, nil
}

// LocationCaptureWithBehavior wrapper
type LocationCaptureWithBehavior struct {
	locationCapture *LocationCapture
	behavior        *ComponentBehavior
}

func (lcwb *LocationCaptureWithBehavior) Kind() string {
	return lcwb.locationCapture.Kind()
}

func (lcwb *LocationCaptureWithBehavior) Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) {
	spec, err := lcwb.locationCapture.Spec(ctx, rctx)
	if err != nil {
		return spec, err
	}

	spec.Behavior = lcwb.behavior
	return spec, nil
}

// GeoResolve componente para normalizar/resolver endereço (spec v2.2)
type GeoResolve struct {
	qualityMin float64
	useCache   bool
	det        liquid.Detector
}

// NewGeoResolve cria nova instância
func NewGeoResolve(det liquid.Detector) *GeoResolve {
	return &GeoResolve{
		det:        det,
		qualityMin: 0.6,
		useCache:   true,
	}
}

func (gr *GeoResolve) Kind() string { return "geo_resolve" }

// WithQualityMin define qualidade mínima aceita
func (gr *GeoResolve) WithQualityMin(min float64) *GeoResolve {
	cp := *gr
	cp.qualityMin = min
	return &cp
}

// WithCache define se usa cache
func (gr *GeoResolve) WithCache(use bool) *GeoResolve {
	cp := *gr
	cp.useCache = use
	return &cp
}

// Spec gera o ComponentSpec
func (gr *GeoResolve) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	metaData := map[string]any{
		"quality_min":    gr.qualityMin,
		"use_cache":      gr.useCache,
		"component_type": "geo_resolve",
		// Outputs: resolved | no_match | error
	}

	return ComponentSpec{
		Kind: "geo_resolve",
		Meta: metaData,
	}, nil
}

// GeoResolveFactory factory
type GeoResolveFactory struct{ det liquid.Detector }

func NewGeoResolveFactory(det liquid.Detector) *GeoResolveFactory {
	return &GeoResolveFactory{det: det}
}

func (f *GeoResolveFactory) New(_ string, props map[string]any) (Component, error) {
	gr := NewGeoResolve(f.det)

	if qualityMin, ok := props["quality_min"].(float64); ok {
		gr = gr.WithQualityMin(qualityMin)
	}

	if useCache, ok := props["use_cache"].(bool); ok {
		gr = gr.WithCache(useCache)
	}

	// Parse behaviors
	behavior, err := ParseBehavior(props, f.det)
	if err != nil {
		return nil, err
	}

	return &GeoResolveWithBehavior{
		geoResolve: gr,
		behavior:   behavior,
	}, nil
}

// GeoResolveWithBehavior wrapper
type GeoResolveWithBehavior struct {
	geoResolve *GeoResolve
	behavior   *ComponentBehavior
}

func (grwb *GeoResolveWithBehavior) Kind() string {
	return grwb.geoResolve.Kind()
}

func (grwb *GeoResolveWithBehavior) Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) {
	spec, err := grwb.geoResolve.Spec(ctx, rctx)
	if err != nil {
		return spec, err
	}

	spec.Behavior = grwb.behavior
	return spec, nil
}
