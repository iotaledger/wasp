// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dss

import (
	"fmt"

	"fortio.org/safecast"

	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/suites"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/gpa"
)

type msgPartialSig struct {
	gpa.BasicMessage
	suite      suites.Suite // Transient, for un-marshaling only.
	partialSig *dss.PartialSig
}

var _ gpa.Message = new(msgPartialSig)

func (m *msgPartialSig) MsgType() gpa.MessageType {
	return msgTypePartialSig
}

func (m *msgPartialSig) MarshalBCS(e *bcs.Encoder) error {
	val, err := safecast.Convert[uint16](m.partialSig.Partial.I)
	if err != nil {
		return err
	}
	e.WriteUint16(val) // TODO: Resolve it from the context, instead of marshaling.

	if _, err := m.partialSig.Partial.V.MarshalTo(e); err != nil {
		return fmt.Errorf("marshaling PartialSig.Partial.V: %w", err)
	}

	e.Encode(m.partialSig.SessionID)
	e.Encode(m.partialSig.Signature)

	return nil
}

func (m *msgPartialSig) UnmarshalBCS(d *bcs.Decoder) error {
	m.partialSig = &dss.PartialSig{Partial: &share.PriShare{}}
	m.partialSig.Partial.I = int(d.ReadUint16())

	m.partialSig.Partial.V = m.suite.Scalar()
	if _, err := m.partialSig.Partial.V.UnmarshalFrom(d); err != nil {
		return fmt.Errorf("unmarshaling PartialSig.Partial.V: %w", err)
	}

	m.partialSig.SessionID = bcs.Decode[[]byte](d)
	m.partialSig.Signature = bcs.Decode[[]byte](d)

	return nil
}
