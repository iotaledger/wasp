// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsign

import (
	"bytes"
	"fmt"

	"fortio.org/safecast"

	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/suites"
)

func NewMsgPartialSig(partialSig *dss.PartialSig) (MsgPartialSig, error) {
	partialI, err := safecast.Convert[uint16](partialSig.Partial.I) // TODO: Resolve it from the context, instead of marshaling.
	if err != nil {
		return MsgPartialSig{}, err
	}

	var partialV bytes.Buffer
	if _, err := partialSig.Partial.V.MarshalTo(&partialV); err != nil {
		return MsgPartialSig{}, fmt.Errorf("marshaling PartialSig.Partial.V: %w", err)
	}

	return MsgPartialSig{
		PartialI:  partialI,
		PartialV:  partialV.Bytes(),
		SessionID: partialSig.SessionID,
		Signature: partialSig.Signature,
	}, nil
}

type MsgPartialSig struct {
	PartialI  uint16
	PartialV  []byte
	SessionID []byte
	Signature []byte
}

func (m *MsgPartialSig) PartialSig(suite suites.Suite) (*dss.PartialSig, error) {
	partialSig := &dss.PartialSig{Partial: &share.PriShare{}}
	partialSig.Partial.I = int(m.PartialI)

	partialSig.Partial.V = suite.Scalar()
	if _, err := partialSig.Partial.V.UnmarshalFrom(bytes.NewReader(m.PartialV)); err != nil {
		return nil, fmt.Errorf("unmarshaling PartialSig.Partial.V: %w", err)
	}

	partialSig.SessionID = m.SessionID
	partialSig.Signature = m.Signature

	return partialSig, nil
}
