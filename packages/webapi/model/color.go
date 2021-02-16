package model

import (
	"encoding/json"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/util"
)

// Color is the base58 representation of balance.Color
type Color string

func NewColor(color *balance.Color) Color {
	return Color(color.String())
}

func (c Color) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(c))
}

func (c *Color) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	_, err := util.ColorFromString(s)
	*c = Color(s)
	return err
}

func (c Color) Color() balance.Color {
	col, err := util.ColorFromString(string(c))
	if err != nil {
		panic(err)
	}
	return col
}
