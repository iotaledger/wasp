// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtypes

import (
	"github.com/ethereum/go-ethereum"

	"github.com/iotaledger/wasp/packages/util/bcs"
)

func init() {
	bcs.AddCustomEncoder(func(e *bcs.Encoder, msg ethereum.CallMsg) error {
		_ = e.Encode(msg.From)
		_ = e.EncodeOptional(msg.To)
		_ = e.WriteCompactUint(msg.Gas)
		_ = e.EncodeOptional(msg.Value)
		_ = e.Encode(msg.Data)
		return e.Err()
	})

	bcs.AddCustomDecoder(func(d *bcs.Decoder, msg *ethereum.CallMsg) error {
		_ = d.Decode(&msg.From)
		_, _ = d.DecodeOptional(&msg.To)
		msg.Gas = d.ReadCompactUint()
		_, _ = d.DecodeOptional(&msg.Value)
		_ = d.Decode(&msg.Data)
		return d.Err()
	})
}

func EncodeCallMsg(args ethereum.CallMsg) []byte {
	return bcs.MustMarshal(&args)
}

func DecodeCallMsg(data []byte) ethereum.CallMsg {
	return bcs.MustUnmarshal[ethereum.CallMsg](data)
}
