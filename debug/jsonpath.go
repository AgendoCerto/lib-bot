package debug

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/AgendoCerto/lib-bot/io"
)

// JSONPathResolver resolve paths no formato $.routes[3].view.text
type JSONPathResolver struct {
	doc *io.DesignDoc
}

// NewJSONPathResolver cria um novo resolver
func NewJSONPathResolver(doc *io.DesignDoc) *JSONPathResolver {
	return &JSONPathResolver{doc: doc}
}

// ResolvePath interpreta um path JSONPath e retorna a configuração atual
func (r *JSONPathResolver) ResolvePath(path string) (interface{}, error) {
	if !strings.HasPrefix(path, "$.") {
		return nil, fmt.Errorf("path deve começar com '$.'")
	}

	// Remove o prefixo "$."
	path = strings.TrimPrefix(path, "$.")

	// Converte o documento para JSON para navegação
	docJSON, err := json.Marshal(r.doc)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar documento: %w", err)
	}

	var docMap map[string]interface{}
	if err := json.Unmarshal(docJSON, &docMap); err != nil {
		return nil, fmt.Errorf("erro ao deserializar documento: %w", err)
	}

	return r.navigatePath(docMap, path)
}

// navigatePath navega pelo path no objeto JSON
func (r *JSONPathResolver) navigatePath(obj interface{}, path string) (interface{}, error) {
	if path == "" {
		return obj, nil
	}

	// Divide o path no primeiro ponto ou colchete
	var key, remaining string

	if idx := strings.IndexAny(path, ".["); idx != -1 {
		if path[idx] == '.' {
			key = path[:idx]
			remaining = path[idx+1:]
		} else { // '['
			key = path[:idx]
			// Procura o fechamento do colchete
			closingIdx := strings.Index(path[idx:], "]")
			if closingIdx == -1 {
				return nil, fmt.Errorf("colchete não fechado no path: %s", path)
			}

			arrayIndex := path[idx+1 : idx+closingIdx]
			remaining = path[idx+closingIdx+1:]
			remaining = strings.TrimPrefix(remaining, ".")

			// Navega para o array primeiro
			arrayObj, err := r.getObjectKey(obj, key)
			if err != nil {
				return nil, err
			}

			// Depois acessa o índice
			return r.navigateArrayIndex(arrayObj, arrayIndex, remaining)
		}
	} else {
		key = path
		remaining = ""
	}

	nextObj, err := r.getObjectKey(obj, key)
	if err != nil {
		return nil, err
	}

	return r.navigatePath(nextObj, remaining)
}

// getObjectKey obtém uma chave de um objeto
func (r *JSONPathResolver) getObjectKey(obj interface{}, key string) (interface{}, error) {
	switch v := obj.(type) {
	case map[string]interface{}:
		if val, ok := v[key]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("chave '%s' não encontrada no objeto", key)
	default:
		return nil, fmt.Errorf("objeto não é um map, não pode acessar chave '%s'", key)
	}
}

// navigateArrayIndex navega por um índice de array
func (r *JSONPathResolver) navigateArrayIndex(obj interface{}, indexStr string, remaining string) (interface{}, error) {
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return nil, fmt.Errorf("índice de array inválido: %s", indexStr)
	}

	switch v := obj.(type) {
	case []interface{}:
		if index < 0 || index >= len(v) {
			return nil, fmt.Errorf("índice %d fora do range do array (tamanho: %d)", index, len(v))
		}
		return r.navigatePath(v[index], remaining)
	default:
		return nil, fmt.Errorf("objeto não é um array, não pode acessar índice [%d]", index)
	}
}

// FormatValue formata o valor encontrado para exibição
func (r *JSONPathResolver) FormatValue(value interface{}) string {
	if value == nil {
		return "<nil>"
	}

	// Se for string simples, retorna direto
	if str, ok := value.(string); ok {
		return fmt.Sprintf("string: %q", str)
	}

	// Se for número
	if num, ok := value.(float64); ok {
		return fmt.Sprintf("number: %g", num)
	}

	// Se for boolean
	if b, ok := value.(bool); ok {
		return fmt.Sprintf("boolean: %t", b)
	}

	// Para objetos complexos, serializa como JSON
	jsonBytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("error formatting: %v", err)
	}
	return string(jsonBytes)
}

// GetRouteInfo obtém informações específicas sobre uma rota quando o path contém "routes[X]"
func (r *JSONPathResolver) GetRouteInfo(path string) (map[string]interface{}, error) {
	// Extrai o índice da rota do path
	routeRegex := regexp.MustCompile(`routes\[(\d+)\]`)
	matches := routeRegex.FindStringSubmatch(path)
	if len(matches) < 2 {
		return nil, fmt.Errorf("path não contém referência válida a routes[X]")
	}

	routeIndex, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("índice de rota inválido: %s", matches[1])
	}

	// Obtém informações sobre a rota específica
	info := map[string]interface{}{
		"route_index": routeIndex,
		"path":        path,
	}

	// Tenta obter o valor atual no path
	if value, err := r.ResolvePath(path); err == nil {
		info["current_value"] = value
	} else {
		info["error"] = err.Error()
	}

	// Se for um nó do grafo, adiciona informações contextuais
	if r.doc != nil && len(r.doc.Graph.Nodes) > 0 {
		if routeIndex < len(r.doc.Graph.Nodes) {
			node := r.doc.Graph.Nodes[routeIndex]
			info["node_id"] = node.ID
			info["node_kind"] = node.Kind
			info["node_title"] = node.Title
		}
	}

	return info, nil
}
