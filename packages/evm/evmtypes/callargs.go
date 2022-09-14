// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtypes

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/hive.go/core/marshalutil"
)

func EncodeCallMsg(c ethereum.CallMsg) []byte {
	m := marshalutil.New()
	m.WriteBytes(c.From.Bytes())
	m.WriteBool(c.To != nil)
	if c.To != nil {
		m.WriteBytes(c.To.Bytes())
	}
	m.WriteUint64(c.Gas)
	m.WriteBool(c.GasPrice != nil)
	if c.GasPrice != nil {
		writeBytes(m, c.GasPrice.Bytes())
	}
	m.WriteBool(c.Value != nil)
	if c.Value != nil {
		writeBytes(m, c.Value.Bytes())
	}
	writeBytes(m, c.Data)
	return m.Bytes()
}

func DecodeCallMsg(callArgsBytes []byte) (ret ethereum.CallMsg, err error) {
	m := marshalutil.New(callArgsBytes)
	var b []byte
	var exists bool

	if b, err = m.ReadBytes(common.AddressLength); err != nil {
		return
	}
	ret.From.SetBytes(b)

	if exists, err = m.ReadBool(); err != nil {
		return
	}
	if exists {
		if b, err = m.ReadBytes(common.AddressLength); err != nil {
			return
		}
		ret.To = &common.Address{}
		ret.To.SetBytes(b)
	}

	if ret.Gas, err = m.ReadUint64(); err != nil {
		return
	}

	if exists, err = m.ReadBool(); err != nil {
		return
	}
	if exists {
		if b, err = readBytes(m); err != nil {
			return
		}
		ret.GasPrice = new(big.Int)
		ret.GasPrice.SetBytes(b)
	}

	if exists, err = m.ReadBool(); err != nil {
		return
	}
	if exists {
		if b, err = readBytes(m); err != nil {
			return
		}
		ret.Value = new(big.Int)
		ret.Value.SetBytes(b)
	}

	if ret.Data, err = readBytes(m); err != nil {
		return
	}
	return ret, err
}
