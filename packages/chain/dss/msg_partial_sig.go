// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dss

import (
	"bytes"

	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util"
)

type msgPartialSig struct {
	suite      suites.Suite // Transient, for un-marshalling only.
	sender     gpa.NodeID   // Transient.
	recipient  gpa.NodeID   // Transient.
	partialSig *dss.PartialSig
}

var _ gpa.Message = &msgPartialSig{}

func (m *msgPartialSig) Recipient() gpa.NodeID {
	return m.recipient
}

func (m *msgPartialSig) SetSender(sender gpa.NodeID) {
	m.sender = sender
}

func (m *msgPartialSig) MarshalBinary() ([]byte, error) {
	w := &bytes.Buffer{}
	if err := util.WriteByte(w, msgTypePartialSig); err != nil {
		return nil, xerrors.Errorf("cannot marshal type=msgTypePartialSig: %w", err)
	}
	if err := util.WriteUint16(w, uint16(m.partialSig.Partial.I)); err != nil { // TODO: Resolve it from the context, instead of marshalling.
		return nil, xerrors.Errorf("cannot marshal partialSig.Partial.I: %w", err)
	}
	if err := util.WriteMarshaled(w, m.partialSig.Partial.V); err != nil {
		return nil, xerrors.Errorf("cannot marshal partialSig.Partial.V: %w", err)
	}
	if err := util.WriteBytes16(w, m.partialSig.SessionID); err != nil {
		return nil, xerrors.Errorf("cannot marshal m.partialSig.SessionID: %w", err)
	}
	if err := util.WriteBytes16(w, m.partialSig.Signature); err != nil {
		return nil, xerrors.Errorf("cannot marshal partialSig.Signature: %w", err)
	}
	return w.Bytes(), nil
}

func (m *msgPartialSig) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	msgType, err := util.ReadByte(r)
	if err != nil {
		return err
	}
	if msgType != msgTypePartialSig {
		return xerrors.Errorf("unexpected msgType=%v in dss.msgPartialSig", msgType)
	}
	var partialI uint16
	if err := util.ReadUint16(r, &partialI); err != nil {
		return err
	}
	partialV := m.suite.Scalar()
	if err := util.ReadMarshaled(r, partialV); err != nil {
		return xerrors.Errorf("cannot unmarshal partialSig.V: %w", err)
	}
	m.partialSig = &dss.PartialSig{
		Partial: &share.PriShare{I: int(partialI), V: partialV},
	}
	m.partialSig.SessionID, err = util.ReadBytes16(r)
	if err != nil {
		return err
	}
	m.partialSig.Signature, err = util.ReadBytes16(r)
	if err != nil {
		return err
	}
	return nil
}
