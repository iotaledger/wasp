package model

import (
	"encoding/json"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
)

// Color is the base58 representation of ledgerstate.Color
type Color string

func NewColor(color ledgerstate.Color) Color {
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
	_, err := ledgerstate.ColorFromBase58EncodedString(s)
	*c = Color(s)
	return err
}

func (c Color) Color() ledgerstate.Color {
	col, err := ledgerstate.ColorFromBase58EncodedString(string(c))
	if err != nil {
		panic(err)
	}
	return col
}
