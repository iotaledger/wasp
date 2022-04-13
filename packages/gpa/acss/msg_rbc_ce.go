// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/suites"
)

// This message is used as a payload of the RBC:
//
// > RBC(C||E)
//
type msgRBCCEPayload struct {
	suite suites.Suite
	C     *share.PubPoly
	E     [][]byte
}

func (m *msgRBCCEPayload) MarshalBinary() ([]byte, error) {
	w := &bytes.Buffer{}
	//
	// Write C.
	base, commits := m.C.Info()
	if base == nil {
		base = m.suite.Point().Base()
	}
	if err := util.WriteMarshaled(w, base); err != nil {
		return nil, err
	}
	if err := util.WriteUint16(w, uint16(len(commits))); err != nil {
		return nil, err
	}
	for i := range commits {
		if err := util.WriteMarshaled(w, commits[i]); err != nil {
			return nil, err
		}
	}
	//
	// Write E.
	if err := util.WriteUint16(w, uint16(len(m.E))); err != nil {
		return nil, err
	}
	for i := range m.E {
		if err := util.WriteBytes16(w, m.E[i]); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

func (m *msgRBCCEPayload) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	//
	// Read C
	base := m.suite.Point()
	if err := util.ReadMarshaled(r, base); err != nil {
		return err
	}
	var commitsLen uint16
	if err := util.ReadUint16(r, &commitsLen); err != nil {
		return err
	}
	commits := make([]kyber.Point, commitsLen)
	for i := range commits {
		commits[i] = m.suite.Point()
		if err := util.ReadMarshaled(r, commits[i]); err != nil {
			return err
		}
	}
	m.C = share.NewPubPoly(m.suite, base, commits)
	//
	// Read E
	var eLen uint16
	if err := util.ReadUint16(r, &eLen); err != nil {
		return err
	}
	m.E = make([][]byte, eLen)
	for i := range m.E {
		var err error
		m.E[i], err = util.ReadBytes16(r)
		if err != nil {
			return err
		}
	}
	return nil
}

//
// An event to self.
type msgRBCCEOutput struct {
	me      gpa.NodeID
	payload *msgRBCCEPayload
}

func (m *msgRBCCEOutput) Recipient() gpa.NodeID {
	return m.me
}

func (m *msgRBCCEOutput) MarshalBinary() ([]byte, error) {
	return nil, nil // TODO: Implement.
}
