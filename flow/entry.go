package flow

// EntryKind define os tipos de pontos de entrada do fluxo
type EntryKind string

const (
	EntryGlobalStart  EntryKind = "global_start"  // Entrada padrão para todos os canais
	EntryChannelStart EntryKind = "channel_start" // Entrada específica para um canal
	EntryForced       EntryKind = "forced"        // Entrada forçada (bypass de condições)
)

// Entry representa um ponto de entrada no fluxo de conversação
type Entry struct {
	Kind      EntryKind `json:"kind"`                 // Tipo de entrada
	ChannelID string    `json:"channel_id,omitempty"` // ID do canal (se específico)
	Target    ID        `json:"target"`               // ID do nó de destino
}
