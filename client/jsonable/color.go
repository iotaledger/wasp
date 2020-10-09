package jsonable

import (
	"encoding/json"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/util"
)

type Color struct {
	color balance.Color
}

func NewColor(color *balance.Color) *Color {
	return &Color{color: *color}
}

func (c Color) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Color().String())
}

func (c *Color) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	col, err := util.ColorFromString(s)
	c.color = col
	return err
}

func (c Color) Color() *balance.Color {
	return &c.color
}
