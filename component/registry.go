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

	// Registra todos os componentes padrão
	reg.Register("message", NewMessageFactory(det))
	reg.Register("text", NewTextFactory(det))
	reg.Register("confirm", NewConfirmFactory(det))
	reg.Register("buttons", NewButtonsFactory(det))
	reg.Register("listpicker", NewListPickerFactory(det))
	reg.Register("media", NewMediaFactory(det))
	reg.Register("carousel", NewCarouselFactory(det))
	reg.Register("delay", NewSimpleFactory(DelayFactory))

	return reg
}
