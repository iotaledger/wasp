// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtypes

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
)

func DecodeGenesisAlloc(b []byte) (core.GenesisAlloc, error) {
	m := marshalutil.New(b)
	var err error
	r := core.GenesisAlloc{}

	var n uint32
	if n, err = m.ReadUint32(); err != nil {
		return nil, err
	}

	for i := 0; i < int(n); i++ {
		var b []byte

		if b, err = readBytes(m); err != nil {
			return nil, err
		}
		address := common.BytesToAddress(b)

		g := core.GenesisAccount{}

		if g.Code, err = readBytes(m); err != nil {
			return nil, err
		}

		var storageCount uint32
		if storageCount, err = m.ReadUint32(); err != nil {
			return nil, err
		}
		g.Storage = map[common.Hash]common.Hash{}
		for j := 0; j < int(storageCount); j++ {
			var k, v []byte
			if k, err = readBytes(m); err != nil {
				return nil, err
			}
			if v, err = readBytes(m); err != nil {
				return nil, err
			}
			g.Storage[common.BytesToHash(k)] = common.BytesToHash(v)
		}

		if b, err = readBytes(m); err != nil {
			return nil, err
		}
		g.Balance = big.NewInt(0)
		g.Balance.SetBytes(b)

		if g.Nonce, err = m.ReadUint64(); err != nil {
			return nil, err
		}

		if g.PrivateKey, err = readBytes(m); err != nil {
			return nil, err
		}

		r[address] = g
	}
	return r, nil
}

func EncodeGenesisAlloc(alloc core.GenesisAlloc) []byte {
	m := marshalutil.New()
	m.WriteUint32(uint32(len(alloc)))
	for addr, account := range alloc {
		writeBytes(m, addr.Bytes())
		writeBytes(m, account.Code)
		m.WriteUint32(uint32(len(account.Storage)))
		for k, v := range account.Storage {
			writeBytes(m, k.Bytes())
			writeBytes(m, v.Bytes())
		}
		writeBytes(m, account.Balance.Bytes())
		m.WriteUint64(account.Nonce)
		writeBytes(m, account.PrivateKey)
	}
	return m.Bytes()
}

func readBytes(m *marshalutil.MarshalUtil) (b []byte, err error) {
	var n uint32
	if n, err = m.ReadUint32(); err != nil {
		return nil, err
	}
	return m.ReadBytes(int(n))
}

func writeBytes(m *marshalutil.MarshalUtil, b []byte) {
	m.WriteUint32(uint32(len(b)))
	m.WriteBytes(b)
}
