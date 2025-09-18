package adapter

// Capabilities descreve as capacidades e limitações de um canal específico
// Serve como fonte de verdade para validação de componentes
type Capabilities struct {
	SupportsHSM        bool            // Suporte a Highly Structured Messages
	SupportsRichText   bool            // Suporte a texto formatado (markdown, etc.)
	MaxTextLen         int             // Limite máximo de caracteres em texto
	MaxButtons         int             // Número máximo de botões por mensagem
	ButtonKinds        map[string]bool // Tipos de botão suportados {"reply":true,"url":true,"call":true}
	SupportsCarousel   bool            // Suporte a carrosséis de cards
	SupportsListPicker bool            // Suporte a listas de seleção
}

// NewCaps cria configuração padrão conservadora de capabilities
func NewCaps() Capabilities {
	return Capabilities{
		SupportsHSM:        false,
		SupportsRichText:   false,
		MaxTextLen:         1000,
		MaxButtons:         3,
		ButtonKinds:        map[string]bool{"reply": true},
		SupportsCarousel:   false,
		SupportsListPicker: false,
	}
}
