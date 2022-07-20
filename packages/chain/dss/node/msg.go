// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
)

func makeMsgData(key hashing.HashValue, index int, data []byte) ([]byte, error) {
	buf := &bytes.Buffer{}
	if _, err := buf.Write(key[:]); err != nil {
		return nil, err
	}
	if err := util.WriteUint32(buf, uint32(index)); err != nil {
		return nil, err
	}
	if err := util.WriteBytes16(buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func fromMsgData(msgPayload []byte) (hashing.HashValue, int, []byte, error) {
	r := bytes.NewReader(msgPayload)
	h := hashing.HashValue{}
	if err := h.Read(r); err != nil {
		return hashing.NilHash, 0, nil, err
	}
	var index uint32
	if err := util.ReadUint32(r, &index); err != nil {
		return hashing.NilHash, 0, nil, err
	}
	var data []byte
	var err error
	if data, err = util.ReadBytes16(r); err != nil {
		return hashing.NilHash, 0, nil, err
	}
	return h, int(index), data, nil
}
