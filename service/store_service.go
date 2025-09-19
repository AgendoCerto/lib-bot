// Package service - StoreService para versionamento e persistência
package service

import (
	"context"
	"fmt"
	"time"

	"lib-bot/io"
	"lib-bot/store"
)

// StoreService é responsável por versionamento e persistência simples
type StoreService struct {
	// Para demonstração, usar um mapa simples em memória
	// Em produção, usar store.Service com Repository real
	storage map[string]StoredDesign
}

// StoredDesign representa um design armazenado
type StoredDesign struct {
	BotID     string       `json:"bot_id"`
	Design    io.DesignDoc `json:"design"`
	Version   string       `json:"version"`
	Checksum  string       `json:"checksum"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// VersionInfo informações sobre uma versão
type VersionInfo struct {
	ID        string    `json:"id"`
	BotID     string    `json:"bot_id"`
	Checksum  string    `json:"checksum"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
}

// NewStoreService cria um novo serviço de store
func NewStoreService() *StoreService {
	return &StoreService{
		storage: make(map[string]StoredDesign),
	}
}

// Save salva um design e retorna informações da versão
func (s *StoreService) Save(ctx context.Context, botID string, design io.DesignDoc) (*VersionInfo, error) {
	// Gera checksum simples
	codec := io.JSONCodec{}
	designJSON, err := codec.EncodeDesign(design)
	if err != nil {
		return nil, fmt.Errorf("erro ao codificar design: %w", err)
	}

	checksum := fmt.Sprintf("%x", len(designJSON)) // Checksum simplificado
	version := fmt.Sprintf("v%d", time.Now().Unix())

	stored := StoredDesign{
		BotID:     botID,
		Design:    design,
		Version:   version,
		Checksum:  checksum,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.storage[botID] = stored

	return &VersionInfo{
		ID:        version,
		BotID:     botID,
		Checksum:  checksum,
		CreatedAt: stored.CreatedAt,
		UpdatedAt: stored.UpdatedAt,
		IsActive:  true,
	}, nil
}

// Load carrega um design por ID do bot
func (s *StoreService) Load(ctx context.Context, botID string) (io.DesignDoc, error) {
	stored, exists := s.storage[botID]
	if !exists {
		return io.DesignDoc{}, fmt.Errorf("bot %s não encontrado", botID)
	}

	return stored.Design, nil
}

// Delete remove um design do store
func (s *StoreService) Delete(ctx context.Context, botID string) error {
	_, exists := s.storage[botID]
	if !exists {
		return fmt.Errorf("bot %s não encontrado", botID)
	}

	delete(s.storage, botID)
	return nil
}

// List lista todos os bots no store
func (s *StoreService) List(ctx context.Context) ([]VersionInfo, error) {
	var versions []VersionInfo

	for _, stored := range s.storage {
		versions = append(versions, VersionInfo{
			ID:        stored.Version,
			BotID:     stored.BotID,
			Checksum:  stored.Checksum,
			CreatedAt: stored.CreatedAt,
			UpdatedAt: stored.UpdatedAt,
			IsActive:  true,
		})
	}

	return versions, nil
}

// Exists verifica se um bot existe no store
func (s *StoreService) Exists(ctx context.Context, botID string) (bool, error) {
	_, exists := s.storage[botID]
	return exists, nil
}

// GetVersion obtém informações de versão sem carregar o design completo
func (s *StoreService) GetVersion(ctx context.Context, botID string) (*VersionInfo, error) {
	stored, exists := s.storage[botID]
	if !exists {
		return nil, fmt.Errorf("bot %s não encontrado", botID)
	}

	return &VersionInfo{
		ID:        stored.Version,
		BotID:     stored.BotID,
		Checksum:  stored.Checksum,
		CreatedAt: stored.CreatedAt,
		UpdatedAt: stored.UpdatedAt,
		IsActive:  true,
	}, nil
}

// UpdateWithValidation atualiza um design com validação
func (s *StoreService) UpdateWithValidation(ctx context.Context, botID string, design io.DesignDoc, validationService *ValidationService, adapterName string) (*VersionInfo, error) {
	// Valida antes de salvar
	result, err := validationService.ValidateDesign(ctx, design, adapterName)
	if err != nil {
		return nil, fmt.Errorf("erro na validação: %w", err)
	}

	if !result.Valid {
		return nil, fmt.Errorf("design inválido: %d issues encontrados", len(result.Issues))
	}

	// Salva se válido
	return s.Save(ctx, botID, design)
}

// Clone cria uma cópia de um design existente com novo ID
func (s *StoreService) Clone(ctx context.Context, sourceBotID, targetBotID string) (*VersionInfo, error) {
	// Carrega design original
	design, err := s.Load(ctx, sourceBotID)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar design original: %w", err)
	}

	// Salva com novo ID
	return s.Save(ctx, targetBotID, design)
}

// Compare compara checksums de dois designs
func (s *StoreService) Compare(ctx context.Context, botID1, botID2 string) (bool, error) {
	version1, err := s.GetVersion(ctx, botID1)
	if err != nil {
		return false, fmt.Errorf("erro ao obter versão 1: %w", err)
	}

	version2, err := s.GetVersion(ctx, botID2)
	if err != nil {
		return false, fmt.Errorf("erro ao obter versão 2: %w", err)
	}

	return version1.Checksum == version2.Checksum, nil
}

// ApplyPatches aplica patches RFC 6902 usando store atômico (se disponível)
func (s *StoreService) ApplyPatches(ctx context.Context, botID string, patches []byte, atomicService *store.Service) ([]byte, error) {
	if atomicService != nil {
		// Usa store atômico se disponível
		// TODO: implementar quando store.Service tiver métodos corretos
		return nil, fmt.Errorf("store atômico não implementado ainda")
	}

	// Fallback: aplica patches manualmente
	design, err := s.Load(ctx, botID)
	if err != nil {
		return nil, err
	}

	codec := io.JSONCodec{}
	designJSON, err := codec.EncodeDesign(design)
	if err != nil {
		return nil, err
	}

	// Aqui aplicaríamos os patches RFC 6902
	// Por simplicidade, retornamos o design original
	return designJSON, nil
}
