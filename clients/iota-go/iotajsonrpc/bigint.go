package iotajsonrpc

import (
	"encoding/json"
	"fmt"
	"math/big"
)

type Uint128 = BigInt

type BigInt struct {
	*big.Int
}

func NewBigInt(v uint64) *BigInt {
	return &BigInt{new(big.Int).SetUint64(v)}
}

func NewBigIntInt64(v int64) *BigInt {
	return &BigInt{new(big.Int).SetInt64(v)}
}

func (w *BigInt) UnmarshalText(data []byte) error {
	return w.UnmarshalJSON(data)
}

func (w *BigInt) UnmarshalJSON(data []byte) error {
	if w.Int == nil {
		w.Int = new(big.Int)
	}
	// Handle string-wrapped numbers (e.g., "\"123\"" in JSON)
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		_, ok := w.Int.SetString(s, 10)
		if !ok {
			return fmt.Errorf("invalid number string: %s", s)
		}
		return nil
	}
	// Delegate to standard big.Int unmarshaling for numeric values
	return w.Int.UnmarshalJSON(data)
}

func (w *BigInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.String())
}

func (w *BigInt) Clone() *BigInt {
	if w.Int == nil {
		return NewBigInt(0)
	}
	return &BigInt{new(big.Int).Set(w.Int)}
}
