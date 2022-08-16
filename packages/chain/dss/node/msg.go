// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/util"
)

func makeMsgData(key string, index int, data []byte) ([]byte, error) {
	w := &bytes.Buffer{}
	if err := util.WriteString16(w, key); err != nil {
		return nil, err
	}
	if err := util.WriteUint32(w, uint32(index)); err != nil {
		return nil, err
	}
	if err := util.WriteBytes16(w, data); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func fromMsgData(msgPayload []byte) (string, int, []byte, error) {
	var err error
	r := bytes.NewReader(msgPayload)
	var key string
	if key, err = util.ReadString16(r); err != nil {
		return "", 0, nil, err
	}
	var index uint32
	if err := util.ReadUint32(r, &index); err != nil {
		return "", 0, nil, err
	}
	var data []byte
	if data, err = util.ReadBytes16(r); err != nil {
		return "", 0, nil, err
	}
	return key, int(index), data, nil
}
