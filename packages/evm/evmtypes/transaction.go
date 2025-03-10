// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtypes

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/iotaledger/bcs-go"
)

func init() {
	bcs.AddCustomEncoder(func(e *bcs.Encoder, tx *types.Transaction) error {
		var b bytes.Buffer
		if err := tx.EncodeRLP(&b); err != nil {
			return fmt.Errorf("failed to RLP encode transaction: %w", err)
		}

		// We can't just do "return tx.EncodeRLP(e)" because we also need to write number of bytes (see decoding below).
		e.Encode(b.Bytes())
		if err := e.Err(); err != nil {
			return fmt.Errorf("failed to write transaction bytes: %w", err)
		}

		return nil
	})

	bcs.AddCustomDecoder(func(d *bcs.Decoder, tx *types.Transaction) error {
		// Unfortunately, we can't just do "return tx.DecodeRLP(d)" because it will consume all the bytes it can from the stream.
		// So we need to limit it - either by passing inputLimit or by feeding separate stream.
		// For some reason inputLimit was not working for me, so using separate stream.

		b := bcs.Decode[[]byte](d)
		if err := d.Err(); err != nil {
			return fmt.Errorf("failed to read transaction bytes: %w", err)
		}

		r := bytes.NewReader(b)

		if err := tx.DecodeRLP(rlp.NewStream(r, 0)); err != nil {
			return fmt.Errorf("failed to RLP decode transaction: %w", err)
		}

		return nil
	})
}

func EncodeTransaction(tx *types.Transaction) []byte {
	w := new(bytes.Buffer)
	_ = tx.EncodeRLP(w)
	return w.Bytes()
}

func DecodeTransaction(b []byte) (*types.Transaction, error) {
	tx := new(types.Transaction)
	err := tx.DecodeRLP(rlp.NewStream(bytes.NewReader(b), 0))
	return tx, err
}
