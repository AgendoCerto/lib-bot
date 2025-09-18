package flow

// VersionStatus define os estados possíveis de uma versão do fluxo
type VersionStatus string

const (
	Development VersionStatus = "development" // Versão em desenvolvimento (editável)
	Production  VersionStatus = "production"  // Versão em produção (read-only)
)

// Version representa uma versão específica do fluxo de conversação
type Version struct {
	ID     string        `json:"id"`     // Identificador único da versão
	Status VersionStatus `json:"status"` // Estado atual da versão
}
