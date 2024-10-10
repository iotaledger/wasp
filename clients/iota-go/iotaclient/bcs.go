package iotaclient

import (
	"bytes"
	"errors"

	"github.com/iotaledger/wasp/packages/util/bcs"
)

// UnmarshalBCS is a shortcut for bcs.Unmarshal that also verifies
// that the consumed bytes is exactly len(data).
func UnmarshalBCS[Obj any](data []byte, obj *Obj) error {
	r := bytes.NewReader(data)

	if _, err := bcs.UnmarshalStreamInto(r, obj); err != nil {
		return err
	}
	if r.Len() != 0 {
		return errors.New("excess bytes")
	}
	return nil
}
