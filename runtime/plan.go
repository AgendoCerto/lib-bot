package runtime

// Tipos leves usados pelo sender (opcional complementar ao io.RuntimePlan).
type PlanConstraints struct {
	MaxTextLen int `json:"max_text_len,omitempty"`
	MaxButtons int `json:"max_buttons,omitempty"`
}
