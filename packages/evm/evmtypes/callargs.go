// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtypes

import (
	"github.com/ethereum/go-ethereum"

	"github.com/iotaledger/wasp/packages/util/bcs"
)

func init() {
	bcs.AddCustomEncoder(func(e *bcs.Encoder, msg ethereum.CallMsg) error {
		e.Encode(msg.From)
		e.EncodeOptional(msg.To)
		e.WriteUint64(msg.Gas)
		e.Encode(msg.Value)
		e.Encode(msg.Data)

		return e.Err()
	})

	bcs.AddCustomDecoder(func(d *bcs.Decoder, msg *ethereum.CallMsg) error {
		d.Decode(&msg.From)
		d.DecodeOptional(&msg.To)
		msg.Gas = d.ReadUint64()
		d.Decode(&msg.Value)
		d.Decode(&msg.Data)

		return d.Err()
	})
}

func EncodeCallMsg(args ethereum.CallMsg) []byte {
	return bcs.MustMarshal(&args)
}

func DecodeCallMsg(data []byte) ethereum.CallMsg {
	return bcs.MustUnmarshal[ethereum.CallMsg](data)
}
