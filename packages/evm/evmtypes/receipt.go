// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtypes

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/iotaledger/hive.go/core/marshalutil"
)

// EncodeReceipt serializes the receipt in RLP format
func EncodeReceipt(receipt *types.Receipt) []byte {
	var b bytes.Buffer
	err := receipt.EncodeRLP(&b)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func DecodeReceipt(b []byte) (*types.Receipt, error) {
	receipt := new(types.Receipt)
	err := receipt.DecodeRLP(rlp.NewStream(bytes.NewReader(b), 0))
	return receipt, err
}

// EncodeReceiptFull encodes the receipt including fields not serialized by EncodeReceipt
func EncodeReceiptFull(r *types.Receipt) []byte {
	m := marshalutil.New()
	writeBytes(m, EncodeReceipt(r))
	m.WriteBytes(r.TxHash.Bytes())
	m.WriteBytes(r.BlockHash.Bytes())
	writeBytes(m, r.BlockNumber.Bytes())
	m.WriteBytes(r.ContractAddress.Bytes())
	m.WriteUint64(r.GasUsed)
	return m.Bytes()
}

func DecodeReceiptFull(receiptBytes []byte) (*types.Receipt, error) {
	m := marshalutil.New(receiptBytes)
	var err error
	var b []byte

	if b, err = readBytes(m); err != nil {
		return nil, err
	}
	var r *types.Receipt
	if r, err = DecodeReceipt(b); err != nil {
		return nil, err
	}

	if b, err = m.ReadBytes(common.HashLength); err != nil {
		return nil, err
	}
	r.TxHash.SetBytes(b)

	if b, err = m.ReadBytes(common.HashLength); err != nil {
		return nil, err
	}
	r.BlockHash.SetBytes(b)

	if b, err = readBytes(m); err != nil {
		return nil, err
	}
	r.BlockNumber = new(big.Int).SetBytes(b)

	if b, err = m.ReadBytes(common.AddressLength); err != nil {
		return nil, err
	}
	r.ContractAddress.SetBytes(b)

	if r.GasUsed, err = m.ReadUint64(); err != nil {
		return nil, err
	}

	return r, nil
}
