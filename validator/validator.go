package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Config representa a configuração completa do validator 2.0
// IMPORTANTE: Quando enabled=true, o componente AGUARDA automaticamente a resposta do usuário
// O Validator substitui o antigo behavior.await
type Config struct {
	Enabled       bool       `json:"enabled"`
	Normalize     *Normalize `json:"normalize,omitempty"`
	Routes        []Route    `json:"routes"`
	DefaultOutput string     `json:"default_output"`

	// Timeout (opcional) - tempo máximo para aguardar resposta
	// Se não definido, aguarda indefinidamente
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
	TimeoutOutput  string `json:"timeout_output,omitempty"` // Output quando timeout (se não definido, usa default_output)
}

// Normalize configura normalização de texto antes da validação
type Normalize struct {
	Trim             bool `json:"trim"`
	CollapseSpaces   bool `json:"collapse_spaces"`
	Lower            bool `json:"lower"`
	RemoveDiacritics bool `json:"remove_diacritics"`
}

// Route representa uma rota de validação com múltiplos modos
type Route struct {
	Go    string `json:"go"`    // Nome do output quando esta rota bater
	Modes Modes  `json:"modes"` // Modos de validação (regex, tags, rules, expr, hook)
}

// Modes agrupa os diferentes modos de validação
type Modes struct {
	Regex []RegexMode `json:"regex,omitempty"`
	Tags  []TagsMode  `json:"tags,omitempty"`
	Rules *RulesMode  `json:"rules,omitempty"`
	Expr  string      `json:"expr,omitempty"`
	Hook  *HookMode   `json:"hook,omitempty"`
}

// RegexMode valida usando expressões regulares
type RegexMode struct {
	Field   string `json:"field"`   // Campo a validar (ex: "context.user_text")
	Pattern string `json:"pattern"` // Regex pattern
	Flags   string `json:"flags"`   // Flags: i (case-insensitive), m (multiline)
}

// TagsMode valida usando matching de palavras/frases
type TagsMode struct {
	Field     string   `json:"field"`     // Campo a validar
	Match     []string `json:"match"`     // Lista de palavras/frases
	Strategy  string   `json:"strategy"`  // exact|contains|starts_with|ends_with|fuzzy
	Threshold float64  `json:"threshold"` // Para fuzzy matching (0-1)
}

// RulesMode valida usando regras declarativas
type RulesMode struct {
	Logic string `json:"logic"` // AND|OR
	All   []Rule `json:"all"`   // Lista de regras
}

// Rule representa uma regra individual
type Rule struct {
	Left  string      `json:"left"`  // Campo/variável (ex: "profile.terms.accepted")
	Op    string      `json:"op"`    // Operador
	Right interface{} `json:"right"` // Valor a comparar
}

// HookMode valida chamando serviço externo
type HookMode struct {
	URL       string            `json:"url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers,omitempty"`
	TimeoutMs int               `json:"timeout_ms"`
	CacheTTLs int               `json:"cache_ttl_s,omitempty"`
}

// HookResponse resposta esperada do hook
type HookResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
	Route   string `json:"route,omitempty"`
}

// Validator implementa a lógica de validação 2.0
type Validator struct {
	config    *Config
	variables map[string]interface{}
	client    *http.Client
	cache     map[string]*cacheEntry
}

type cacheEntry struct {
	response  *HookResponse
	expiresAt time.Time
}

// NewValidator cria novo validador
func NewValidator(config *Config, variables map[string]interface{}) *Validator {
	return &Validator{
		config:    config,
		variables: variables,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		cache: make(map[string]*cacheEntry),
	}
}

// Validate executa a validação e retorna o output apropriado
func (v *Validator) Validate(ctx context.Context) (string, error) {
	if !v.config.Enabled {
		return v.config.DefaultOutput, nil
	}

	// Avaliar rotas na ordem
	for _, route := range v.config.Routes {
		matched, err := v.evaluateRoute(ctx, route)
		if err != nil {
			return "", fmt.Errorf("error evaluating route %s: %w", route.Go, err)
		}
		if matched {
			return route.Go, nil
		}
	}

	return v.config.DefaultOutput, nil
}

// evaluateRoute avalia uma rota específica
func (v *Validator) evaluateRoute(ctx context.Context, route Route) (bool, error) {
	modes := route.Modes

	// Avaliar regex
	if len(modes.Regex) > 0 {
		for _, regex := range modes.Regex {
			matched, err := v.evaluateRegex(regex)
			if err != nil {
				return false, err
			}
			if !matched {
				return false, nil
			}
		}
	}

	// Avaliar tags
	if len(modes.Tags) > 0 {
		for _, tags := range modes.Tags {
			matched := v.evaluateTags(tags)
			if !matched {
				return false, nil
			}
		}
	}

	// Avaliar rules
	if modes.Rules != nil {
		matched, err := v.evaluateRules(modes.Rules)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}

	// Avaliar expr
	if modes.Expr != "" {
		matched, err := v.evaluateExpr(modes.Expr)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}

	// Avaliar hook
	if modes.Hook != nil {
		matched, err := v.evaluateHook(ctx, modes.Hook)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}

	return true, nil
}

// evaluateRegex avalia modo regex
func (v *Validator) evaluateRegex(mode RegexMode) (bool, error) {
	value := v.getFieldValue(mode.Field)
	if value == nil {
		return false, nil
	}

	text, ok := value.(string)
	if !ok {
		return false, nil
	}

	// Normalizar se configurado
	if v.config.Normalize != nil {
		text = v.normalize(text)
	}

	// Compilar regex com flags
	pattern := mode.Pattern
	if strings.Contains(mode.Flags, "i") {
		pattern = "(?i)" + pattern
	}
	if strings.Contains(mode.Flags, "m") {
		pattern = "(?m)" + pattern
	}

	// Timeout de 2s para prevenir ReDoS
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan bool, 1)
	var matched bool
	var err error

	go func() {
		re, compileErr := regexp.Compile(pattern)
		if compileErr != nil {
			err = compileErr
			done <- false
			return
		}
		matched = re.MatchString(text)
		done <- true
	}()

	select {
	case <-ctx.Done():
		return false, fmt.Errorf("regex timeout (possible ReDoS): %s", pattern)
	case <-done:
		return matched, err
	}
}

// evaluateTags avalia modo tags
func (v *Validator) evaluateTags(mode TagsMode) bool {
	value := v.getFieldValue(mode.Field)
	if value == nil {
		return false
	}

	text, ok := value.(string)
	if !ok {
		return false
	}

	// Normalizar
	if v.config.Normalize != nil {
		text = v.normalize(text)
	}

	// Normalizar matches também
	normalizedMatches := make([]string, len(mode.Match))
	if v.config.Normalize != nil {
		for i, m := range mode.Match {
			normalizedMatches[i] = v.normalize(m)
		}
	} else {
		normalizedMatches = mode.Match
	}

	// Aplicar estratégia
	for _, match := range normalizedMatches {
		matched := false
		switch mode.Strategy {
		case "exact":
			matched = text == match
		case "contains":
			matched = strings.Contains(text, match)
		case "starts_with":
			matched = strings.HasPrefix(text, match)
		case "ends_with":
			matched = strings.HasSuffix(text, match)
		case "fuzzy":
			similarity := v.fuzzyMatch(text, match)
			matched = similarity >= mode.Threshold
		default:
			matched = strings.Contains(text, match)
		}

		if matched {
			return true
		}
	}

	return false
}

// evaluateRules avalia modo rules
func (v *Validator) evaluateRules(mode *RulesMode) (bool, error) {
	if len(mode.All) == 0 {
		return true, nil
	}

	logic := strings.ToUpper(mode.Logic)
	if logic == "" {
		logic = "AND"
	}

	results := make([]bool, len(mode.All))
	for i, rule := range mode.All {
		matched, err := v.evaluateRule(rule)
		if err != nil {
			return false, err
		}
		results[i] = matched
	}

	// Aplicar lógica
	if logic == "AND" {
		for _, r := range results {
			if !r {
				return false, nil
			}
		}
		return true, nil
	} else if logic == "OR" {
		for _, r := range results {
			if r {
				return true, nil
			}
		}
		return false, nil
	}

	return false, fmt.Errorf("invalid logic operator: %s", mode.Logic)
}

// evaluateRule avalia uma regra individual
func (v *Validator) evaluateRule(rule Rule) (bool, error) {
	left := v.getFieldValue(rule.Left)

	switch rule.Op {
	case "exists":
		return left != nil, nil
	case "empty":
		if left == nil {
			return true, nil
		}
		str, ok := left.(string)
		return ok && str == "", nil
	case "==":
		return v.compareEquals(left, rule.Right), nil
	case "!=":
		return !v.compareEquals(left, rule.Right), nil
	case ">":
		return v.compareGreater(left, rule.Right), nil
	case ">=":
		return v.compareGreaterOrEqual(left, rule.Right), nil
	case "<":
		return v.compareLess(left, rule.Right), nil
	case "<=":
		return v.compareLessOrEqual(left, rule.Right), nil
	case "in":
		return v.compareIn(left, rule.Right), nil
	case "not_in":
		return !v.compareIn(left, rule.Right), nil
	case "contains":
		return v.compareContains(left, rule.Right), nil
	case "starts_with":
		return v.compareStartsWith(left, rule.Right), nil
	case "ends_with":
		return v.compareEndsWith(left, rule.Right), nil
	default:
		return false, fmt.Errorf("invalid operator: %s", rule.Op)
	}
}

// evaluateExpr avalia expressão booleana (implementação simplificada)
func (v *Validator) evaluateExpr(expr string) (bool, error) {
	// TODO: implementar parser completo de expressões booleanas
	return false, fmt.Errorf("expr mode not fully implemented yet")
}

// evaluateHook avalia chamando hook externo
func (v *Validator) evaluateHook(ctx context.Context, mode *HookMode) (bool, error) {
	// Verificar cache
	cacheKey := mode.URL
	if entry, found := v.cache[cacheKey]; found {
		if time.Now().Before(entry.expiresAt) {
			return entry.response.Valid, nil
		}
		delete(v.cache, cacheKey)
	}

	// Preparar request
	timeout := time.Duration(mode.TimeoutMs) * time.Millisecond
	if timeout == 0 {
		timeout = 1800 * time.Millisecond
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Criar payload com variáveis
	payload, err := json.Marshal(v.variables)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequestWithContext(ctx, mode.Method, mode.URL, strings.NewReader(string(payload)))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range mode.Headers {
		req.Header.Set(k, v)
	}

	// Executar request
	resp, err := v.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("hook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("hook returned status %d", resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var hookResp HookResponse
	if err := json.Unmarshal(body, &hookResp); err != nil {
		return false, err
	}

	// Cachear se configurado
	if mode.CacheTTLs > 0 {
		v.cache[cacheKey] = &cacheEntry{
			response:  &hookResp,
			expiresAt: time.Now().Add(time.Duration(mode.CacheTTLs) * time.Second),
		}
	}

	return hookResp.Valid, nil
}

// Helper functions

func (v *Validator) normalize(text string) string {
	if v.config.Normalize == nil {
		return text
	}

	result := text

	if v.config.Normalize.Trim {
		result = strings.TrimSpace(result)
	}

	if v.config.Normalize.Lower {
		result = strings.ToLower(result)
	}

	if v.config.Normalize.CollapseSpaces {
		result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
	}

	if v.config.Normalize.RemoveDiacritics {
		t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
		result, _, _ = transform.String(t, result)
	}

	return result
}

func (v *Validator) getFieldValue(field string) interface{} {
	parts := strings.Split(field, ".")
	if len(parts) == 0 {
		return nil
	}

	current := v.variables
	for i, part := range parts {
		if current == nil {
			return nil
		}

		val, ok := current[part]
		if !ok {
			return nil
		}

		if i == len(parts)-1 {
			return val
		}

		current, ok = val.(map[string]interface{})
		if !ok {
			return nil
		}
	}

	return nil
}

func (v *Validator) compareEquals(left, right interface{}) bool {
	return fmt.Sprintf("%v", left) == fmt.Sprintf("%v", right)
}

func (v *Validator) compareGreater(left, right interface{}) bool {
	l, lok := toFloat(left)
	r, rok := toFloat(right)
	return lok && rok && l > r
}

func (v *Validator) compareGreaterOrEqual(left, right interface{}) bool {
	l, lok := toFloat(left)
	r, rok := toFloat(right)
	return lok && rok && l >= r
}

func (v *Validator) compareLess(left, right interface{}) bool {
	l, lok := toFloat(left)
	r, rok := toFloat(right)
	return lok && rok && l < r
}

func (v *Validator) compareLessOrEqual(left, right interface{}) bool {
	l, lok := toFloat(left)
	r, rok := toFloat(right)
	return lok && rok && l <= r
}

func (v *Validator) compareIn(left, right interface{}) bool {
	arr, ok := right.([]interface{})
	if !ok {
		return false
	}
	for _, item := range arr {
		if v.compareEquals(left, item) {
			return true
		}
	}
	return false
}

func (v *Validator) compareContains(left, right interface{}) bool {
	l, lok := left.(string)
	r, rok := right.(string)
	return lok && rok && strings.Contains(l, r)
}

func (v *Validator) compareStartsWith(left, right interface{}) bool {
	l, lok := left.(string)
	r, rok := right.(string)
	return lok && rok && strings.HasPrefix(l, r)
}

func (v *Validator) compareEndsWith(left, right interface{}) bool {
	l, lok := left.(string)
	r, rok := right.(string)
	return lok && rok && strings.HasSuffix(l, r)
}

func (v *Validator) fuzzyMatch(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	distance := levenshtein(s1, s2)
	maxLen := max(len(s1), len(s2))
	return 1.0 - float64(distance)/float64(maxLen)
}

func levenshtein(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	rows := len(r1) + 1
	cols := len(r2) + 1

	dist := make([][]int, rows)
	for i := range dist {
		dist[i] = make([]int, cols)
		dist[i][0] = i
	}
	for j := range dist[0] {
		dist[0][j] = j
	}

	for i := 1; i < rows; i++ {
		for j := 1; j < cols; j++ {
			cost := 1
			if r1[i-1] == r2[j-1] {
				cost = 0
			}
			dist[i][j] = min(
				dist[i-1][j]+1,
				dist[i][j-1]+1,
				dist[i-1][j-1]+cost,
			)
		}
	}

	return dist[rows-1][cols-1]
}

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
