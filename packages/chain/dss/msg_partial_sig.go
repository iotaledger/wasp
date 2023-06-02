// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dss

import (
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util"
)

type msgPartialSig struct {
	suite      suites.Suite // Transient, for un-marshaling only.
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
	mu := new(marshalutil.MarshalUtil)
	mu.WriteByte(msgTypePartialSig)
	mu.WriteUint16(uint16(m.partialSig.Partial.I))
	util.WriteMarshaledMu(mu, m.partialSig.Partial.V)
	util.WriteBytesMu(mu, m.partialSig.SessionID)
	util.WriteBytesMu(mu, m.partialSig.Signature)
	return mu.Bytes(), nil
}

func (m *msgPartialSig) UnmarshalBinary(data []byte) error {
	mu := marshalutil.New(data)
	msgType, err := mu.ReadByte()
	if err != nil {
		return err
	}
	if msgType != msgTypePartialSig {
		return fmt.Errorf("unexpected msgType=%v in dss.msgPartialSig", msgType)
	}
	partialI, err := mu.ReadUint16()
	if err != nil {
		return err
	}
	partialV := m.suite.Scalar()
	if err2 := util.ReadMarshaledMu(mu, partialV); err2 != nil {
		return fmt.Errorf("cannot unmarshal partialSig.V: %w", err2)
	}
	m.partialSig = &dss.PartialSig{
		Partial: &share.PriShare{I: int(partialI), V: partialV},
	}
	m.partialSig.SessionID, err = util.ReadBytesMu(mu)
	if err != nil {
		return err
	}
	m.partialSig.Signature, err = util.ReadBytesMu(mu)
	if err != nil {
		return err
	}
	return nil
}
