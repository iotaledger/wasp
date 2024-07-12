package suiclient

import (
	"errors"

	"github.com/fardream/go-bcs/bcs"
)

// UnmarshalBCS is a shortcut for bcs.Unmarshal that also verifies
// that the consumed bytes is exactly len(data).
func UnmarshalBCS(data []byte, obj any) error {
	n, err := bcs.Unmarshal(data, &obj)
	if err != nil {
		return err
	}
	if n != len(data) {
		return errors.New("excess bytes")
	}
	return nil
}