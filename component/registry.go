package component

import "github.com/AgendoCerto/lib-bot/liquid"

// SimpleFactory implementa Factory para componentes simples
type SimpleFactory struct {
	creator func(map[string]any) (Component, error)
}

// NewSimpleFactory cria uma factory simples
func NewSimpleFactory(creator func(map[string]any) (Component, error)) *SimpleFactory {
	return &SimpleFactory{creator: creator}
}

// New implementa Factory interface
func (f *SimpleFactory) New(kind string, props map[string]any) (Component, error) {
	return f.creator(props)
}

// DefaultRegistry cria um registry com todos os componentes padrão
func DefaultRegistry() *Registry {
	reg := NewRegistry()
	det := liquid.NoRenderDetector{} // Detector sem renderização

	// Registra componentes base
	reg.Register("message", NewMessageFactory(det))
	reg.Register("buttons", NewButtonsFactory(det))
	reg.Register("listpicker", NewListPickerFactory(det))
	reg.Register("media", NewMediaFactory(det))
	reg.Register("carousel", NewCarouselFactory(det))
	reg.Register("delay", NewSimpleFactory(DelayFactory))
	reg.Register("terms", NewTermsFactory(det))
	reg.Register("feedback", NewFeedbackFactory(det))
	reg.Register("global_start", NewGlobalStartFactory(det))

	// Registra componentes spec v2.2
	reg.Register("terms_gate", NewTermsGateFactory(det))
	reg.Register("hsm_trigger", NewHSMTriggerFactory(det))
	reg.Register("location_capture", NewLocationCaptureFactory(det))
	reg.Register("geo_resolve", NewGeoResolveFactory(det))
	reg.Register("unit_finder", NewUnitFinderFactory(det))
	reg.Register("slot_picker", NewSlotPickerFactory(det))
	reg.Register("payment_link", NewPaymentLinkFactory(det))
	reg.Register("order_cart", NewOrderCartFactory(det))
	reg.Register("human_handoff", NewHumanHandoffFactory(det))

	return reg
}
