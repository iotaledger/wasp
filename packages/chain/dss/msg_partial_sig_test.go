// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dss

import (
	"testing"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	dkg "go.dedis.ch/kyber/v3/share/dkg/rabin"
	"go.dedis.ch/kyber/v3/sign/dss"
)

func TestMsgPartialSigSerialization(t *testing.T) {
	// FIXME
	t.Skip()
	s := tcrypto.DefaultEd25519Suite()
	var nbParticipants = 7
	partPubs := make([]kyber.Point, nbParticipants)
	partSec := make([]kyber.Scalar, nbParticipants)
	for i := 0; i < nbParticipants; i++ {
		sec, pub := genPair(s)
		partPubs[i] = pub
		partSec[i] = sec
	}
	longterms := genDistSecret(s, nbParticipants, partSec, partPubs)
	randoms := genDistSecret(s, nbParticipants, partSec, partPubs)
	dss, err := dss.NewDSS(s, partSec[0], partPubs, longterms[0], randoms[0], []byte("hello"), 4)
	require.NoError(t, err)
	partialSig, err := dss.PartialSig()
	require.NoError(t, err)
	msg := &msgPartialSig{
		gpa.BasicMessage{},
		s,
		partialSig,
	}

	rwutil.ReadWriteTest(t, msg, new(msgPartialSig))
}

func genDistSecret(suite dkg.Suite, nbParticipants int, partSec []kyber.Scalar, partPubs []kyber.Point) []*dkg.DistKeyShare {
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
func genPair(suite dkg.Suite) (kyber.Scalar, kyber.Point) {
	sc := suite.Scalar().Pick(suite.RandomStream())
	return sc, suite.Point().Mul(sc, nil)
}
