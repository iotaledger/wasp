// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dss

import (
	"fmt"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	dkg "go.dedis.ch/kyber/v3/share/dkg/rabin"
	"go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

type msgPartialSig struct {
	gpa.BasicMessage
	suite      suites.Suite // Transient, for un-marshaling only.
	partialSig *dss.PartialSig
}

var _ gpa.Message = new(msgPartialSig)

func (m *msgPartialSig) MarshalBCS(e *bcs.Encoder) error {
	e.WriteUint16(uint16(m.partialSig.Partial.I)) // TODO: Resolve it from the context, instead of marshaling.

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

	return d.Err()
}

//nolint:gocyclo
func GenDistSecret(suite dkg.Suite, nbParticipants int, partSec []kyber.Scalar, partPubs []kyber.Point) []*dkg.DistKeyShare {
	dkgs := make([]*dkg.DistKeyGenerator, nbParticipants)
	for i := 0; i < nbParticipants; i++ {
		dkg, err := dkg.NewDistKeyGenerator(suite, suite, partSec[i], partPubs, nbParticipants/2+1)
		if err != nil {
			panic(err)
		}
		dkgs[i] = dkg
	}
	// full secret sharing exchange
	// 1. broadcast deals
	resps := make([]*dkg.Response, 0, nbParticipants*nbParticipants)
	for _, dkg := range dkgs {
		deals, err := dkg.Deals()
		if err != nil {
			panic(err)
		}
		for i, d := range deals {
			resp, err := dkgs[i].ProcessDeal(d)
			if err != nil {
				panic(err)
			}
			if !resp.Response.Approved {
				panic("wrong approval")
			}
			resps = append(resps, resp)
		}
	}
	// 2. Broadcast responses
	for _, resp := range resps {
		for h, dkg := range dkgs {
			// ignore all messages from ourself
			if resp.Response.Index == uint32(h) {
				continue
			}
			j, err := dkg.ProcessResponse(resp)
			if err != nil || j != nil {
				panic("wrongProcessResponse")
			}
		}
	}
	// 4. Broadcast secret commitment
	for i, dkg := range dkgs {
		scs, err := dkg.SecretCommits()
		if err != nil {
			panic("wrong SecretCommits")
		}
		for j, dkg2 := range dkgs {
			if i == j {
				continue
			}
			cc, err := dkg2.ProcessSecretCommits(scs)
			if err != nil || cc != nil {
				panic("wrong ProcessSecretCommits")
			}
		}
	}

	// 5. reveal shares
	dkss := make([]*dkg.DistKeyShare, len(dkgs))
	for i, dkg := range dkgs {
		dks, err := dkg.DistKeyShare()
		if err != nil {
			panic(err)
		}
		dkss[i] = dks
	}
	return dkss
}

func GenPair(suite dkg.Suite) (kyber.Scalar, kyber.Point) {
	sc := suite.Scalar().Pick(suite.RandomStream())
	return sc, suite.Point().Mul(sc, nil)
}
