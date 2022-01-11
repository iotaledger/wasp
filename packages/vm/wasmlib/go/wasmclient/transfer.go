// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

// The color string can be a base58-encoded 32-byte color, or "IOTA"

type Transfer struct {
	xfer map[string]uint64
}

func NewTransfer() *Transfer {
	return &Transfer{xfer: make(map[string]uint64)}
}

func TransferIotas(amount uint64) *Transfer {
	return TransferTokens("IOTA", amount)
}

func TransferTokens(color string, amount uint64) *Transfer {
	transfer := NewTransfer()
	transfer.Set(color, amount)
	return transfer
}

func (t *Transfer) Set(color string, amount uint64) {
	if color == COLOR_IOTA {
		color = COLOR_IOTA_BASE58
	}
	t.xfer[color] = amount
}
