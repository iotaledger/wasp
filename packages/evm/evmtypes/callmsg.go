// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package evmtypes defines data structures and types for EVM compatibility and operations.
package evmtypes

import (
	"github.com/ethereum/go-ethereum"

	bcs "github.com/iotaledger/bcs-go"
)

func init() {
	bcs.AddCustomEncoder(func(e *bcs.Encoder, msg ethereum.CallMsg) error {
		e.Encode(msg.From)
		e.EncodeOptional(msg.To)
		e.WriteCompactUint64(msg.Gas)
		e.EncodeOptional(msg.Value)
		e.Encode(msg.Data)
		return nil
	})

	bcs.AddCustomDecoder(func(d *bcs.Decoder, msg *ethereum.CallMsg) error {
		d.Decode(&msg.From)
		_ = d.DecodeOptional(&msg.To)
		msg.Gas = d.ReadCompactUint64()
		_ = d.DecodeOptional(&msg.Value)
		d.Decode(&msg.Data)
		return nil
	})
}

func EncodeCallMsg(args ethereum.CallMsg) []byte {
	return bcs.MustMarshal(&args)
}

func DecodeCallMsg(data []byte) ethereum.CallMsg {
	return bcs.MustUnmarshal[ethereum.CallMsg](data)
}
