// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/util/rwutil"
	"go.dedis.ch/kyber/v3/suites"
)

// This message is used as a payload of the RBC:
//
// > RBC(C||E)
type msgRBCCEPayload struct {
	suite suites.Suite
	data  []byte
}

func (m *msgRBCCEPayload) MarshalBinary() ([]byte, error) {
	w := &bytes.Buffer{}
	//
	// Write data.
	if err := rwutil.WriteBytes(w, m.data); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (m *msgRBCCEPayload) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	//
	// Read data
	var err error
	m.data, err = rwutil.ReadBytes(r)
	if err != nil {
		return err
	}
	return nil
}
