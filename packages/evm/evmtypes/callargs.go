// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtypes

import (
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func EncodeCallMsg(c ethereum.CallMsg) []byte {
	ww := rwutil.NewBytesWriter()
	ww.WriteN(c.From[:])
	ww.WriteBool(c.To != nil)
	if c.To != nil {
		ww.WriteN(c.To[:])
	}
	ww.WriteGas64(c.Gas)
	ww.WriteUint256(c.Value)
	ww.WriteBytes(c.Data)
	return ww.Bytes()
}

func DecodeCallMsg(data []byte) (ret ethereum.CallMsg, err error) {
	rr := rwutil.NewBytesReader(data)
	rr.ReadN(ret.From[:])
	hasTo := rr.ReadBool()
	if hasTo {
		ret.To = new(common.Address)
		rr.ReadN(ret.To[:])
	}
	ret.Gas = rr.ReadGas64()
	ret.Value = rr.ReadUint256()
	ret.Data = rr.ReadBytes()
	return ret, rr.Err
}
