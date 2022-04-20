// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package adkg

import (
	"bytes"
	"encoding"

	"github.com/iotaledger/wasp/packages/util"
)

type msgRBCPayload struct {
	t []int
}

var (
	_ encoding.BinaryMarshaler   = &msgRBCPayload{}
	_ encoding.BinaryUnmarshaler = &msgRBCPayload{}
)

func (m *msgRBCPayload) MarshalBinary() ([]byte, error) {
	w := &bytes.Buffer{}
	if err := util.WriteUint16(w, uint16(len(m.t))); err != nil {
		return nil, err
	}
	for i := range m.t {
		if err := util.WriteUint16(w, uint16(m.t[i])); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

func (m *msgRBCPayload) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	var tLen uint16
	if err := util.ReadUint16(r, &tLen); err != nil {
		return err
	}
	m.t = make([]int, tLen)
	for i := range m.t {
		var ti uint16
		err := util.ReadUint16(r, &ti)
		if err != nil {
			return err
		}
		m.t[i] = int(ti)
	}
	return nil
}
