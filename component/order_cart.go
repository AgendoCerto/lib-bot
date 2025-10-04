package component

import (
	"context"
	"fmt"

	"github.com/AgendoCerto/lib-bot/liquid"
	"github.com/AgendoCerto/lib-bot/runtime"
)

// OrderCart componente para gerenciar carrinho de compras (spec v2.2)
type OrderCart struct {
	mode             string         // action | view | checkout
	action           string         // add | remove | clear
	item             map[string]any // Item a adicionar/remover
	showSummary      bool           // Mostrar resumo no canal
	currency         string         // Moeda (BRL, USD, etc)
	autoPersist      bool           // Auto-persistir carrinho
	checkoutProvider string         // Provider de checkout (stripe, mercadopago, etc)
	det              liquid.Detector
}

// NewOrderCart cria nova instância
func NewOrderCart(det liquid.Detector) *OrderCart {
	return &OrderCart{
		det:              det,
		mode:             "action",
		action:           "add",
		showSummary:      true,
		currency:         "BRL",
		autoPersist:      true,
		checkoutProvider: "default",
	}
}

func (oc *OrderCart) Kind() string { return "order_cart" }

// WithMode define o modo de operação
func (oc *OrderCart) WithMode(mode string) *OrderCart {
	cp := *oc
	cp.mode = mode
	return &cp
}

// WithAction define a ação (add/remove/clear)
func (oc *OrderCart) WithAction(action string) *OrderCart {
	cp := *oc
	cp.action = action
	return &cp
}

// WithItem define o item a adicionar/remover
func (oc *OrderCart) WithItem(item map[string]any) *OrderCart {
	cp := *oc
	cp.item = item
	return &cp
}

// WithSummary define se mostra resumo
func (oc *OrderCart) WithSummary(show bool) *OrderCart {
	cp := *oc
	cp.showSummary = show
	return &cp
}

// WithCurrency define a moeda
func (oc *OrderCart) WithCurrency(currency string) *OrderCart {
	cp := *oc
	cp.currency = currency
	return &cp
}

// WithAutoPersist define se persiste automaticamente
func (oc *OrderCart) WithAutoPersist(persist bool) *OrderCart {
	cp := *oc
	cp.autoPersist = persist
	return &cp
}

// WithCheckoutProvider define o provider
func (oc *OrderCart) WithCheckoutProvider(provider string) *OrderCart {
	cp := *oc
	cp.checkoutProvider = provider
	return &cp
}

// Spec gera o ComponentSpec
func (oc *OrderCart) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	metaData := map[string]any{
		"mode":              oc.mode,
		"action":            oc.action,
		"show_summary":      oc.showSummary,
		"currency":          oc.currency,
		"auto_persist":      oc.autoPersist,
		"checkout_provider": oc.checkoutProvider,
		"component_type":    "order_cart",
	}

	if oc.item != nil {
		metaData["item"] = oc.item
	}

	// Outputs conforme spec v2.2:
	// - mode "action" → added | removed | cleared | error
	// - mode "view" → viewed
	// - mode "checkout" → checkout_ready | empty
	var outputsComment string
	switch oc.mode {
	case "action":
		outputsComment = "added | removed | cleared | error"
	case "view":
		outputsComment = "viewed"
	case "checkout":
		outputsComment = "checkout_ready | empty"
	default:
		outputsComment = "added | removed | cleared | error"
	}
	metaData["outputs"] = outputsComment

	return ComponentSpec{
		Kind: "order_cart",
		Meta: metaData,
	}, nil
}

// OrderCartFactory factory
type OrderCartFactory struct{ det liquid.Detector }

func NewOrderCartFactory(det liquid.Detector) *OrderCartFactory {
	return &OrderCartFactory{det: det}
}

func (f *OrderCartFactory) New(_ string, props map[string]any) (Component, error) {
	oc := NewOrderCart(f.det)

	// Mode
	if mode, ok := props["mode"].(string); ok {
		oc = oc.WithMode(mode)
	}

	// Action
	if action, ok := props["action"].(string); ok {
		oc = oc.WithAction(action)
	}

	// Item
	if item, ok := props["item"].(map[string]any); ok {
		oc = oc.WithItem(item)
	}

	// Show summary
	if showSummary, ok := props["show_summary_in_channel"].(bool); ok {
		oc = oc.WithSummary(showSummary)
	}

	// Currency
	if currency, ok := props["currency"].(string); ok {
		oc = oc.WithCurrency(currency)
	}

	// Auto persist
	if autoPersist, ok := props["auto_persist"].(bool); ok {
		oc = oc.WithAutoPersist(autoPersist)
	}

	// Checkout provider
	if provider, ok := props["checkout_provider"].(string); ok {
		oc = oc.WithCheckoutProvider(provider)
	}

	// Validações
	if oc.mode == "action" && oc.action != "clear" && oc.item == nil {
		return nil, fmt.Errorf("order_cart: action '%s' requires 'item' property", oc.action)
	}

	// Parse behavior
	behavior, err := ParseBehavior(props, f.det)
	if err != nil {
		return nil, err
	}

	return &OrderCartWithBehavior{
		orderCart: oc,
		behavior:  behavior,
	}, nil
}

// OrderCartWithBehavior wrapper
type OrderCartWithBehavior struct {
	orderCart *OrderCart
	behavior  *ComponentBehavior
}

func (ocwb *OrderCartWithBehavior) Kind() string {
	return ocwb.orderCart.Kind()
}

func (ocwb *OrderCartWithBehavior) Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) {
	spec, err := ocwb.orderCart.Spec(ctx, rctx)
	if err != nil {
		return spec, err
	}

	spec.Behavior = ocwb.behavior
	return spec, nil
}
