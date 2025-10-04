package component

import (
	"context"

	"github.com/AgendoCerto/lib-bot/liquid"
	"github.com/AgendoCerto/lib-bot/runtime"
)

// SlotPicker componente para seleção de slots de agendamento (spec v2.2)
type SlotPicker struct {
	unitID            string
	serviceID         string
	professionalID    string
	windowH           int // Janela de horários em horas
	pageSize          int
	preferLastPro     bool // Preferir último profissional
	preferLastService bool // Preferir último serviço
	det               liquid.Detector
}

// NewSlotPicker cria nova instância
func NewSlotPicker(det liquid.Detector) *SlotPicker {
	return &SlotPicker{
		det:               det,
		windowH:           72,
		pageSize:          9,
		preferLastPro:     true,
		preferLastService: true,
	}
}

func (sp *SlotPicker) Kind() string { return "slot_picker" }

// WithUnit define unidade
func (sp *SlotPicker) WithUnit(id string) *SlotPicker {
	cp := *sp
	cp.unitID = id
	return &cp
}

// WithService define serviço
func (sp *SlotPicker) WithService(id string) *SlotPicker {
	cp := *sp
	cp.serviceID = id
	return &cp
}

// WithProfessional define profissional
func (sp *SlotPicker) WithProfessional(id string) *SlotPicker {
	cp := *sp
	cp.professionalID = id
	return &cp
}

// WithWindow define janela de tempo
func (sp *SlotPicker) WithWindow(hours int) *SlotPicker {
	cp := *sp
	cp.windowH = hours
	return &cp
}

// WithPagination define paginação
func (sp *SlotPicker) WithPagination(size int) *SlotPicker {
	cp := *sp
	cp.pageSize = size
	return &cp
}

// WithPreferences define preferências
func (sp *SlotPicker) WithPreferences(preferPro, preferService bool) *SlotPicker {
	cp := *sp
	cp.preferLastPro = preferPro
	cp.preferLastService = preferService
	return &cp
}

// Spec gera o ComponentSpec
func (sp *SlotPicker) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	metaData := map[string]any{
		"unit_id":                  sp.unitID,
		"service_id":               sp.serviceID,
		"professional_id":          sp.professionalID,
		"window_h":                 sp.windowH,
		"page_size":                sp.pageSize,
		"prefer_last_professional": sp.preferLastPro,
		"prefer_last_service":      sp.preferLastService,
		"component_type":           "slot_picker",
		// Outputs: chosen | no_slots | next_page | prev_page
	}

	return ComponentSpec{
		Kind: "slot_picker",
		Meta: metaData,
	}, nil
}

// SlotPickerFactory factory
type SlotPickerFactory struct{ det liquid.Detector }

func NewSlotPickerFactory(det liquid.Detector) *SlotPickerFactory {
	return &SlotPickerFactory{det: det}
}

func (f *SlotPickerFactory) New(_ string, props map[string]any) (Component, error) {
	sp := NewSlotPicker(f.det)

	if unitID, ok := props["unit_id"].(string); ok {
		sp = sp.WithUnit(unitID)
	}

	if serviceID, ok := props["service_id"].(string); ok {
		sp = sp.WithService(serviceID)
	}

	if professionalID, ok := props["professional_id"].(string); ok {
		sp = sp.WithProfessional(professionalID)
	}

	if windowH, ok := props["window_h"].(float64); ok {
		sp = sp.WithWindow(int(windowH))
	} else if windowH, ok := props["window_h"].(int); ok {
		sp = sp.WithWindow(windowH)
	}

	if pageSize, ok := props["page_size"].(float64); ok {
		sp = sp.WithPagination(int(pageSize))
	} else if pageSize, ok := props["page_size"].(int); ok {
		sp = sp.WithPagination(pageSize)
	}

	preferPro, _ := props["prefer_last_professional"].(bool)
	preferService, _ := props["prefer_last_service"].(bool)
	sp = sp.WithPreferences(preferPro, preferService)

	behavior, err := ParseBehavior(props, f.det)
	if err != nil {
		return nil, err
	}

	return &SlotPickerWithBehavior{
		slotPicker: sp,
		behavior:   behavior,
	}, nil
}

// SlotPickerWithBehavior wrapper
type SlotPickerWithBehavior struct {
	slotPicker *SlotPicker
	behavior   *ComponentBehavior
}

func (spwb *SlotPickerWithBehavior) Kind() string {
	return spwb.slotPicker.Kind()
}

func (spwb *SlotPickerWithBehavior) Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) {
	spec, err := spwb.slotPicker.Spec(ctx, rctx)
	if err != nil {
		return spec, err
	}

	spec.Behavior = spwb.behavior
	return spec, nil
}

// PaymentLink componente para geração de link de pagamento (spec v2.2)
type PaymentLink struct {
	amount       string
	currency     string
	expiresInMin int
	lockSlotTTL  int
	metadata     map[string]string
	det          liquid.Detector
}

// NewPaymentLink cria nova instância
func NewPaymentLink(det liquid.Detector) *PaymentLink {
	return &PaymentLink{
		det:          det,
		currency:     "BRL",
		expiresInMin: 15,
		lockSlotTTL:  900, // 15 minutos
		metadata:     make(map[string]string),
	}
}

func (pl *PaymentLink) Kind() string { return "payment_link" }

// WithAmount define valor
func (pl *PaymentLink) WithAmount(amount string) *PaymentLink {
	cp := *pl
	cp.amount = amount
	return &cp
}

// WithCurrency define moeda
func (pl *PaymentLink) WithCurrency(currency string) *PaymentLink {
	cp := *pl
	cp.currency = currency
	return &cp
}

// WithExpiration define expiração
func (pl *PaymentLink) WithExpiration(minutes int) *PaymentLink {
	cp := *pl
	cp.expiresInMin = minutes
	return &cp
}

// WithSlotLock define TTL de trava do slot
func (pl *PaymentLink) WithSlotLock(seconds int) *PaymentLink {
	cp := *pl
	cp.lockSlotTTL = seconds
	return &cp
}

// WithMetadata define metadados
func (pl *PaymentLink) WithMetadata(meta map[string]string) *PaymentLink {
	cp := *pl
	cp.metadata = meta
	return &cp
}

// Spec gera o ComponentSpec
func (pl *PaymentLink) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	metaData := map[string]any{
		"amount":          pl.amount,
		"currency":        pl.currency,
		"expires_in_min":  pl.expiresInMin,
		"lock_slot_ttl_s": pl.lockSlotTTL,
		"metadata":        pl.metadata,
		"component_type":  "payment_link",
		// Outputs: paid | expired | failed | abandoned
	}

	return ComponentSpec{
		Kind: "payment_link",
		Meta: metaData,
	}, nil
}

// PaymentLinkFactory factory
type PaymentLinkFactory struct{ det liquid.Detector }

func NewPaymentLinkFactory(det liquid.Detector) *PaymentLinkFactory {
	return &PaymentLinkFactory{det: det}
}

func (f *PaymentLinkFactory) New(_ string, props map[string]any) (Component, error) {
	pl := NewPaymentLink(f.det)

	if amount, ok := props["amount"].(string); ok {
		pl = pl.WithAmount(amount)
	}

	if currency, ok := props["currency"].(string); ok {
		pl = pl.WithCurrency(currency)
	}

	if expiresIn, ok := props["expires_in_min"].(float64); ok {
		pl = pl.WithExpiration(int(expiresIn))
	} else if expiresIn, ok := props["expires_in_min"].(int); ok {
		pl = pl.WithExpiration(expiresIn)
	}

	if lockTTL, ok := props["lock_slot_ttl_s"].(float64); ok {
		pl = pl.WithSlotLock(int(lockTTL))
	} else if lockTTL, ok := props["lock_slot_ttl_s"].(int); ok {
		pl = pl.WithSlotLock(lockTTL)
	}

	if metaRaw, ok := props["metadata"].(map[string]any); ok {
		meta := make(map[string]string)
		for k, v := range metaRaw {
			if str, ok := v.(string); ok {
				meta[k] = str
			}
		}
		pl = pl.WithMetadata(meta)
	}

	behavior, err := ParseBehavior(props, f.det)
	if err != nil {
		return nil, err
	}

	return &PaymentLinkWithBehavior{
		paymentLink: pl,
		behavior:    behavior,
	}, nil
}

// PaymentLinkWithBehavior wrapper
type PaymentLinkWithBehavior struct {
	paymentLink *PaymentLink
	behavior    *ComponentBehavior
}

func (plwb *PaymentLinkWithBehavior) Kind() string {
	return plwb.paymentLink.Kind()
}

func (plwb *PaymentLinkWithBehavior) Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) {
	spec, err := plwb.paymentLink.Spec(ctx, rctx)
	if err != nil {
		return spec, err
	}

	spec.Behavior = plwb.behavior
	return spec, nil
}

// HumanHandoff componente para transferência para humano (spec v2.2)
type HumanHandoff struct {
	slaMinutes       int
	offHoursMessage  string
	requeueIfTimeout bool
	det              liquid.Detector
}

// NewHumanHandoff cria nova instância
func NewHumanHandoff(det liquid.Detector) *HumanHandoff {
	return &HumanHandoff{
		det:              det,
		slaMinutes:       10,
		requeueIfTimeout: true,
	}
}

func (hh *HumanHandoff) Kind() string { return "human_handoff" }

// WithSLA define SLA
func (hh *HumanHandoff) WithSLA(minutes int) *HumanHandoff {
	cp := *hh
	cp.slaMinutes = minutes
	return &cp
}

// WithOffHoursMessage define mensagem fora do horário
func (hh *HumanHandoff) WithOffHoursMessage(msg string) *HumanHandoff {
	cp := *hh
	cp.offHoursMessage = msg
	return &cp
}

// WithRequeue define se recoloca na fila ao timeout
func (hh *HumanHandoff) WithRequeue(requeue bool) *HumanHandoff {
	cp := *hh
	cp.requeueIfTimeout = requeue
	return &cp
}

// Spec gera o ComponentSpec
func (hh *HumanHandoff) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	var textVal *TextValue
	if hh.offHoursMessage != "" {
		meta, err := hh.det.Parse(ctx, hh.offHoursMessage)
		if err != nil {
			return ComponentSpec{}, err
		}
		textVal = &TextValue{
			Raw:      hh.offHoursMessage,
			Template: meta.IsTemplate,
			Liquid:   meta,
		}
	}

	metaData := map[string]any{
		"sla_minutes":        hh.slaMinutes,
		"requeue_if_timeout": hh.requeueIfTimeout,
		"component_type":     "human_handoff",
		// Outputs: queued | agent_joined | closed_by_agent | timeout_to_bot
	}

	return ComponentSpec{
		Kind: "human_handoff",
		Text: textVal,
		Meta: metaData,
	}, nil
}

// HumanHandoffFactory factory
type HumanHandoffFactory struct{ det liquid.Detector }

func NewHumanHandoffFactory(det liquid.Detector) *HumanHandoffFactory {
	return &HumanHandoffFactory{det: det}
}

func (f *HumanHandoffFactory) New(_ string, props map[string]any) (Component, error) {
	hh := NewHumanHandoff(f.det)

	if slaMinutes, ok := props["sla_minutes"].(float64); ok {
		hh = hh.WithSLA(int(slaMinutes))
	} else if slaMinutes, ok := props["sla_minutes"].(int); ok {
		hh = hh.WithSLA(slaMinutes)
	}

	if offHoursMsg, ok := props["off_hours_message"].(string); ok {
		hh = hh.WithOffHoursMessage(offHoursMsg)
	}

	if requeue, ok := props["requeue_if_timeout"].(bool); ok {
		hh = hh.WithRequeue(requeue)
	}

	behavior, err := ParseBehavior(props, f.det)
	if err != nil {
		return nil, err
	}

	return &HumanHandoffWithBehavior{
		humanHandoff: hh,
		behavior:     behavior,
	}, nil
}

// HumanHandoffWithBehavior wrapper
type HumanHandoffWithBehavior struct {
	humanHandoff *HumanHandoff
	behavior     *ComponentBehavior
}

func (hhwb *HumanHandoffWithBehavior) Kind() string {
	return hhwb.humanHandoff.Kind()
}

func (hhwb *HumanHandoffWithBehavior) Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) {
	spec, err := hhwb.humanHandoff.Spec(ctx, rctx)
	if err != nil {
		return spec, err
	}

	spec.Behavior = hhwb.behavior
	return spec, nil
}
