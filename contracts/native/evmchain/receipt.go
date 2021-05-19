// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmchain

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/iotaledger/hive.go/marshalutil"
)

type Receipt struct {
	TxHash            common.Hash
	TransactionIndex  uint32
	BlockHash         common.Hash
	BlockNumber       *big.Int
	From              common.Address
	To                *common.Address
	CumulativeGasUsed uint64
	GasUsed           uint64
	ContractAddress   *common.Address
	Logs              []*types.Log
	Bloom             types.Bloom
	Status            uint64
}

func NewReceipt(r *types.Receipt, from common.Address, to *common.Address) *Receipt {
	ret := &Receipt{
		TxHash:            r.TxHash,
		TransactionIndex:  uint32(r.TransactionIndex),
		BlockHash:         r.BlockHash,
		BlockNumber:       r.BlockNumber,
		From:              from,
		To:                to,
		CumulativeGasUsed: r.CumulativeGasUsed,
		GasUsed:           r.GasUsed,
		Logs:              r.Logs,
		Bloom:             r.Bloom,
		Status:            r.Status,
	}
	if r.ContractAddress != (common.Address{}) {
		ret.ContractAddress = &common.Address{}
		ret.ContractAddress.SetBytes(r.ContractAddress.Bytes())
	}
	return ret
}

func (r *Receipt) Bytes() []byte {
	m := marshalutil.New()
	m.WriteBytes(r.TxHash.Bytes())
	m.WriteUint32(r.TransactionIndex)
	m.WriteBytes(r.BlockHash.Bytes())
	writeBytes(m, r.BlockNumber.Bytes())
	m.WriteBytes(r.From.Bytes())
	m.WriteBool(r.To != nil)
	if r.To != nil {
		m.WriteBytes(r.To.Bytes())
	}
	m.WriteUint64(r.CumulativeGasUsed)
	m.WriteUint64(r.GasUsed)
	m.WriteBool(r.ContractAddress != nil)
	if r.ContractAddress != nil {
		m.WriteBytes(r.ContractAddress.Bytes())
	}
	m.WriteUint32(uint32(len(r.Logs)))
	for _, log := range r.Logs {
		writeBytes(m, EncodeLog(log))
	}
	m.WriteBytes(r.Bloom.Bytes())
	m.WriteUint64(r.Status)
	return m.Bytes()
}

func DecodeReceipt(receiptBytes []byte) (*Receipt, error) {
	m := marshalutil.New(receiptBytes)
	r := &Receipt{}
	var err error
	var b []byte
	var exists bool
	{
		if b, err = m.ReadBytes(common.HashLength); err != nil {
			return nil, err
		}
		r.TxHash.SetBytes(b)
	}
	if r.TransactionIndex, err = m.ReadUint32(); err != nil {
		return nil, err
	}
	{
		if b, err = m.ReadBytes(common.HashLength); err != nil {
			return nil, err
		}
		r.BlockHash.SetBytes(b)
	}
	{
		if b, err = readBytes(m); err != nil {
			return nil, err
		}
		r.BlockNumber = new(big.Int)
		r.BlockNumber.SetBytes(b)
	}
	{
		if b, err = m.ReadBytes(common.AddressLength); err != nil {
			return nil, err
		}
		r.From.SetBytes(b)
	}
	{
		if exists, err = m.ReadBool(); err != nil {
			return nil, err
		}
		if exists {
			if b, err = m.ReadBytes(common.AddressLength); err != nil {
				return nil, err
			}
			r.To = &common.Address{}
			r.To.SetBytes(b)
		}
	}
	if r.CumulativeGasUsed, err = m.ReadUint64(); err != nil {
		return nil, err
	}
	if r.GasUsed, err = m.ReadUint64(); err != nil {
		return nil, err
	}
	{
		if exists, err = m.ReadBool(); err != nil {
			return nil, err
		}
		if exists {
			if b, err = m.ReadBytes(common.AddressLength); err != nil {
				return nil, err
			}
			r.ContractAddress = &common.Address{}
			r.ContractAddress.SetBytes(b)
		}
	}
	{
		var n uint32
		if n, err = m.ReadUint32(); err != nil {
			return nil, err
		}
		for i := 0; i < int(n); i++ {
			if b, err = readBytes(m); err != nil {
				return nil, err
			}
			var log *types.Log
			if log, err = DecodeLog(r, uint(i), b); err != nil {
				return nil, err
			}
			r.Logs = append(r.Logs, log)
		}
	}
	{
		if b, err = m.ReadBytes(types.BloomByteLength); err != nil {
			return nil, err
		}
		r.Bloom.SetBytes(b)
	}
	if r.Status, err = m.ReadUint64(); err != nil {
		return nil, err
	}
	return r, nil
}

func EncodeLog(log *types.Log) []byte {
	var b bytes.Buffer
	err := log.EncodeRLP(&b)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func DecodeLog(r *Receipt, index uint, b []byte) (*types.Log, error) {
	log := new(types.Log)
	err := log.DecodeRLP(rlp.NewStream(bytes.NewReader(b), 0))
	log.BlockNumber = r.BlockNumber.Uint64()
	log.TxHash = r.TxHash
	log.TxIndex = uint(r.TransactionIndex)
	log.BlockHash = r.BlockHash
	log.Index = index
	return log, err
}
