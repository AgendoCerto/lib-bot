package validate

import (
	"fmt"

	"github.com/AgendoCerto/lib-bot/flow"
	"github.com/AgendoCerto/lib-bot/io"
)

// ComponentBehaviorStep valida que componentes só usam behaviors permitidos
type ComponentBehaviorStep struct{}

// NewComponentBehaviorStep cria novo validador de behaviors por componente
func NewComponentBehaviorStep() *ComponentBehaviorStep {
	return &ComponentBehaviorStep{}
}

// ValidateDesign valida behaviors de todos os componentes
func (s *ComponentBehaviorStep) ValidateDesign(design io.DesignDoc) []Issue {
	var issues []Issue

	for i, node := range design.Graph.Nodes {
		path := fmt.Sprintf("graph.nodes[%d]", i)
		issues = append(issues, s.validateComponentBehavior(node, path)...)
	}

	return issues
}

// validateComponentBehavior valida que o componente pode ter os behaviors configurados
func (s *ComponentBehaviorStep) validateComponentBehavior(node flow.Node, path string) []Issue {
	var issues []Issue

	// Extrair behaviors das props
	propsMap := node.Props
	if propsMap == nil {
		return issues
	}

	behaviorRaw, hasBehavior := propsMap["behavior"]
	if !hasBehavior {
		return issues
	}

	behaviorMap, ok := behaviorRaw.(map[string]interface{})
	if !ok {
		return issues
	}

	// Definir quais behaviors cada componente pode ter
	allowedBehaviors := s.getAllowedBehaviors(node.Kind)

	// Verificar cada behavior configurado
	for behaviorType := range behaviorMap {
		if !contains(allowedBehaviors, behaviorType) {
			severity := Err

			// Mensagem genérica mas informativa
			msg := fmt.Sprintf("component '%s' cannot have '%s' behavior - allowed behaviors: %v",
				node.Kind, behaviorType, allowedBehaviors)

			issues = append(issues, Issue{
				Code:     fmt.Sprintf("behavior.%s.not_allowed", node.Kind),
				Severity: severity,
				Path:     fmt.Sprintf("%s.props.behavior.%s", path, behaviorType),
				Msg:      msg,
			})
		}
	}

	return issues
}

// getAllowedBehaviors retorna behaviors permitidos para cada tipo de componente
//
// REGRA INTELIGENTE:
// - TIMEOUT: Componente que ESPERA resposta do usuário
// - VALIDATION: Resposta pode ser INVÁLIDA (texto livre)
// - RETRY: Só se tiver timeout OU validation
// - FALLBACK: Usuário fez algo inesperado (digitou texto em botões, timeout, etc)
// - PERSISTENCE: Componente COLETA dados do usuário
// - DELAY: Qualquer componente (timing universal)
func (s *ComponentBehaviorStep) getAllowedBehaviors(component string) []string {
	switch component {
	case "message":
		// MESSAGE: Pode enviar OU esperar resposta (com await)
		// - timeout: limite de tempo para ENVIAR a mensagem OU aguardar resposta
		// - delay: esperar antes/depois de enviar
		// - await: aguardar resposta do usuário (novo comportamento)
		// - fallback: se await+validation e resposta inválida
		// - validation: validar resposta quando await=true
		// - retry: tentar novamente se timeout ou validation falhar
		return []string{"timeout", "delay", "await", "fallback", "validation", "retry"}

	case "media":
		// MEDIA: Pode enviar OU esperar resposta (com await)
		// - timeout: limite de tempo para ENVIAR a mídia OU aguardar resposta
		// - delay: esperar antes/depois de enviar
		// - await: aguardar resposta do usuário (novo comportamento)
		// - fallback: se await+validation e resposta inválida
		// - validation: validar resposta quando await=true
		// - retry: tentar novamente se timeout ou validation falhar
		return []string{"timeout", "delay", "await", "fallback", "validation", "retry"}

	case "delay":
		// DELAY: Apenas controla tempo, não interage
		// - delay: configuração interna de before/after
		return []string{"delay"}

	case "buttons", "listpicker", "carousel":
		// INTERACTIVE: Envia opções, ESPERA escolha do usuário
		// CENÁRIOS:
		//   1. Usuário clica/escolhe → Output normal (btn_1, item_x)
		//   2. Usuário DIGITA TEXTO → FALLBACK (não reconhecemos)
		//   3. Timeout → TIMEOUT output
		//
		// - timeout: limite de tempo para escolher
		// - fallback: usuário DIGITOU ao invés de clicar
		// - delay: esperar antes/depois
		// - persistence: salvar escolha
		// ❌ validation: Escolha sempre é válida (se clicar)
		// ❌ retry: Vai direto para fallback se digitou
		return []string{"timeout", "fallback", "delay", "persistence"}

	case "terms":
		// TERMS: Envia termos, ESPERA aceitar/rejeitar
		// CENÁRIOS:
		//   1. Usuário aceita corretamente → ACCEPTED
		//   2. Usuário rejeita → REJECTED
		//   3. Usuário digita "talvez" ou algo inválido → VALIDATION (opcional) ou FALLBACK
		//   4. Timeout → TIMEOUT
		//
		// - timeout: limite de tempo para responder
		// - validation: validar formato exato de aceite (OPCIONAL)
		// - retry: tentar novamente se validation ou timeout
		// - fallback: esgotou tentativas ou resposta não reconhecida
		// - delay: esperar antes/depois
		// - persistence: salvar decisão
		return []string{"timeout", "validation", "retry", "fallback", "delay", "persistence"}

	case "feedback":
		// FEEDBACK: Envia pergunta, ESPERA avaliação
		// CENÁRIOS:
		//   1. Usuário envia nota válida (1-10) → SUBMITTED
		//   2. Usuário envia "onze" → INVALID (validation)
		//   3. Timeout → TIMEOUT
		//   4. Esgotou retries → FALLBACK
		//
		// - timeout: limite de tempo para responder
		// - validation: validar formato/range da avaliação
		// - retry: tentar novamente se inválido ou timeout
		// - fallback: esgotou tentativas
		// - delay: esperar antes/depois
		// - persistence: salvar feedback
		return []string{"timeout", "validation", "retry", "fallback", "delay", "persistence"}

	case "text_input", "input":
		// INPUT: Envia pergunta, ESPERA texto livre
		// CENÁRIOS:
		//   1. Usuário envia texto válido → COMPLETE
		//   2. Usuário envia formato inválido → INVALID (validation)
		//   3. Timeout → TIMEOUT
		//   4. Esgotou retries → FALLBACK
		//
		// - timeout: limite de tempo para responder
		// - validation: validar formato (regex, tamanho, etc)
		// - retry: tentar novamente se inválido ou timeout
		// - fallback: esgotou tentativas
		// - delay: esperar antes/depois
		// - persistence: salvar resposta válida
		return []string{"timeout", "validation", "retry", "fallback", "delay", "persistence"}

	case "global_start", "start":
		// START: Ponto de entrada, sem behaviors
		return []string{}

	default:
		// Componentes desconhecidos: permitir todos (não bloqueamos extensões)
		return []string{"timeout", "validation", "retry", "fallback", "delay", "persistence"}
	}
}
