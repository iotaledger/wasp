// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtypes

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/iotaledger/hive.go/core/marshalutil"
)

func EncodeLog(log *types.Log, includeDerivedFields bool) []byte {
	m := marshalutil.New()
	{
		var b bytes.Buffer
		err := log.EncodeRLP(&b)
		if err != nil {
			panic(err)
		}
		writeBytes(m, b.Bytes())
	}
	if includeDerivedFields {
		m.WriteUint64(log.BlockNumber)
		m.Write(log.TxHash)
		m.WriteUint32(uint32(log.TxIndex))
		m.Write(log.BlockHash)
		m.WriteUint32(uint32(log.Index))
	}
	return m.Bytes()
}

func DecodeLog(b []byte, includeDerivedFields bool) (*types.Log, error) {
	log := new(types.Log)
	m := marshalutil.New(b)
	{
		logBytes, err := readBytes(m)
		if err != nil {
			return nil, err
		}
		err = log.DecodeRLP(rlp.NewStream(bytes.NewReader(logBytes), 0))
		if err != nil {
			return nil, err
		}
	}
	if includeDerivedFields {
		var err error
		if log.BlockNumber, err = m.ReadUint64(); err != nil {
			return nil, err
		}
		{
			hashBytes, err := m.ReadBytes(common.HashLength)
			if err != nil {
				return nil, err
			}
			copy(log.TxHash[:], hashBytes)
		}
		{
			n, err := m.ReadUint32()
			if err != nil {
				return nil, err
			}
			log.TxIndex = uint(n)
		}
		{
			hashBytes, err := m.ReadBytes(common.HashLength)
			if err != nil {
				return nil, err
			}
			copy(log.BlockHash[:], hashBytes)
		}
		{
			n, err := m.ReadUint32()
			if err != nil {
				return nil, err
			}
			log.Index = uint(n)
		}
	}
	return log, nil
}

func EncodeLogs(logs []*types.Log) []byte {
	m := marshalutil.New()
	m.WriteUint32(uint32(len(logs)))
	for _, log := range logs {
		writeBytes(m, EncodeLog(log, true))
	}
	return m.Bytes()
}

func DecodeLogs(b []byte) ([]*types.Log, error) {
	m := marshalutil.New(b)
	n, err := m.ReadUint32()
	if err != nil {
		return nil, err
	}
	if int(n) > len(b) {
		// using len(b) as an upper bound to prevent DoS attack allocating the array
		return nil, fmt.Errorf("DecodeLogs: invalid length")
	}
	logs := make([]*types.Log, n)
	for i := uint32(0); i < n; i++ {
		b, err := readBytes(m)
		if err != nil {
			return nil, err
		}
		log, err := DecodeLog(b, true)
		if err != nil {
			return nil, err
		}
		logs[i] = log
	}
	return logs, nil
}

func EncodeFilterQuery(q *ethereum.FilterQuery) []byte {
	buf := new(bytes.Buffer)
	// TODO: using gob temporarily until we decide on a proper binary codec format
	err := gob.NewEncoder(buf).Encode(q)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func DecodeFilterQuery(b []byte) (*ethereum.FilterQuery, error) {
	var q ethereum.FilterQuery
	err := gob.NewDecoder(bytes.NewReader(b)).Decode(&q)
	return &q, err
}
