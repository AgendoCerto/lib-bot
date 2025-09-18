package hsm

import "errors"

// HSM (Highly Structured Messages) - Templates pré-aprovados
// O usuário só precisa informar o 'name' do template HSM.
// A validação e execução é responsabilidade do microserviço.

// HSMTemplate representa um template HSM simplificado
type HSMTemplate struct {
	Name string `json:"name"` // Nome do template (único campo obrigatório)
}

// Validate valida se o HSM template tem o nome obrigatório
func (h HSMTemplate) Validate() error {
	if h.Name == "" {
		return errors.New("hsm: name is required")
	}
	return nil
}
