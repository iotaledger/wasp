// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util"
	"go.dedis.ch/kyber/v3/suites"
)

// This message is used as a payload of the RBC:
//
// > RBC(C||E)
//
type msgRBCCEPayload struct {
	suite suites.Suite
	data  []byte
}

func (m *msgRBCCEPayload) MarshalBinary() ([]byte, error) {
	w := &bytes.Buffer{}
	//
	// Write data.
	if err := util.WriteBytes16(w, m.data); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (m *msgRBCCEPayload) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	//
	// Read data
	var err error
	m.data, err = util.ReadBytes16(r)
	if err != nil {
		return err
	}
	return nil
}

//
// An event to self.
type msgRBCCEOutput struct {
	me      gpa.NodeID
	payload *msgRBCCEPayload
}

var _ gpa.Message = &msgRBCCEOutput{}

func (m *msgRBCCEOutput) Recipient() gpa.NodeID {
	return m.me
}

func (m *msgRBCCEOutput) SetSender(sender gpa.NodeID) {
	// Don't care the sender.
}

func (m *msgRBCCEOutput) MarshalBinary() ([]byte, error) {
	return nil, nil // TODO: Implement.
}
