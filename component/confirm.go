package component

import (
	"context"

	"lib-bot/liquid"
	"lib-bot/persistence"
	"lib-bot/runtime"
)

type Confirm struct {
	title string
	yes   string
	no    string
	det   liquid.Detector
}

func NewConfirm(det liquid.Detector) *Confirm { return &Confirm{det: det} }
func (c *Confirm) Kind() string               { return "confirm" }
func (c *Confirm) WithText(title, yes, no string) *Confirm {
	cp := *c
	cp.title, cp.yes, cp.no = title, yes, no
	return &cp
}

func (c *Confirm) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	mt, err := c.det.Parse(ctx, c.title)
	if err != nil {
		return ComponentSpec{}, err
	}
	my, err := c.det.Parse(ctx, c.yes)
	if err != nil {
		return ComponentSpec{}, err
	}
	mn, err := c.det.Parse(ctx, c.no)
	if err != nil {
		return ComponentSpec{}, err
	}

	return ComponentSpec{
		Kind: "confirm",
		Text: &TextValue{Raw: c.title, Template: mt.IsTemplate, Liquid: mt},
		Buttons: []Button{
			{Label: TextValue{Raw: c.yes, Template: my.IsTemplate, Liquid: my}, Payload: "yes", Kind: "reply"},
			{Label: TextValue{Raw: c.no, Template: mn.IsTemplate, Liquid: mn}, Payload: "no", Kind: "reply"},
		},
	}, nil
}

// Factory

type ConfirmFactory struct{ det liquid.Detector }

func NewConfirmFactory(det liquid.Detector) *ConfirmFactory { return &ConfirmFactory{det: det} }

func (f *ConfirmFactory) New(_ string, props map[string]any) (Component, error) {
	c := NewConfirm(f.det)
	if title, _ := props["title"].(string); title != "" {
		// Aceita tanto yes/no quanto positive/negative
		yes, _ := props["yes"].(string)
		if yes == "" {
			yes, _ = props["positive"].(string)
		}
		no, _ := props["no"].(string)
		if no == "" {
			no, _ = props["negative"].(string)
		}
		c = c.WithText(title, yes, no)
	}

	// Parse behaviors
	behavior, err := ParseBehavior(props, f.det)
	if err != nil {
		return nil, err
	}

	// Parse persistence
	persistence, err := ParsePersistence(props)
	if err != nil {
		return nil, err
	}

	return &ConfirmWithBehaviorAndPersistence{
		confirm:     c,
		behavior:    behavior,
		persistence: persistence,
	}, nil
}

// ConfirmWithBehaviorAndPersistence é um wrapper que inclui behaviors e persistência
type ConfirmWithBehaviorAndPersistence struct {
	confirm     *Confirm
	behavior    *ComponentBehavior
	persistence *persistence.PersistenceConfig
}

func (cwbp *ConfirmWithBehaviorAndPersistence) Kind() string {
	return cwbp.confirm.Kind()
}

func (cwbp *ConfirmWithBehaviorAndPersistence) Spec(ctx context.Context, rctx runtime.Context) (ComponentSpec, error) {
	spec, err := cwbp.confirm.Spec(ctx, rctx)
	if err != nil {
		return spec, err
	}

	spec.Behavior = cwbp.behavior
	spec.Persistence = cwbp.persistence

	return spec, nil
}
