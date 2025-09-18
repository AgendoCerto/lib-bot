package component

import (
	"context"

	"lib-bot/liquid"
	"lib-bot/runtime"
)

type Confirm struct {
	title string
	yes   string
	no    string
	hsm   *HSMView
	det   liquid.Detector
}

func NewConfirm(det liquid.Detector) *Confirm { return &Confirm{det: det} }
func (c *Confirm) Kind() string               { return "confirm" }
func (c *Confirm) WithText(title, yes, no string) *Confirm {
	cp := *c
	cp.title, cp.yes, cp.no = title, yes, no
	return &cp
}
func (c *Confirm) WithHSM(h *HSMView) *Confirm { cp := *c; cp.hsm = h; return &cp }

func (c *Confirm) Spec(ctx context.Context, _ runtime.Context) (ComponentSpec, error) {
	if c.hsm != nil {
		for i := range c.hsm.Params {
			meta, err := c.det.Parse(ctx, c.hsm.Params[i].Raw)
			if err != nil {
				return ComponentSpec{}, err
			}
			c.hsm.Params[i].Template = meta.IsTemplate
			c.hsm.Params[i].Liquid = meta
		}
		for i := range c.hsm.Buttons {
			meta, err := c.det.Parse(ctx, c.hsm.Buttons[i].Label.Raw)
			if err != nil {
				return ComponentSpec{}, err
			}
			c.hsm.Buttons[i].Label.Template = meta.IsTemplate
			c.hsm.Buttons[i].Label.Liquid = meta
		}
		return ComponentSpec{Kind: "confirm", HSM: c.hsm}, nil
	}

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
		yes, _ := props["positive"].(string)
		no, _ := props["negative"].(string)
		c = c.WithText(title, yes, no)
	}
	if raw, ok := props["hsm_ref"].(map[string]any); ok && raw != nil {
		h := decodeHSM(raw)
		c = c.WithHSM(h)
	}
	return c, nil
}
