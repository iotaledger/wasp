// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dss

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/dss"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestMsgPartialSigSerialization(t *testing.T) {
	s := tcrypto.DefaultEd25519Suite()
	nbParticipants := 7
	partPubs := make([]kyber.Point, nbParticipants)
	partSec := make([]kyber.Scalar, nbParticipants)
	for i := 0; i < nbParticipants; i++ {
		sec, pub := GenPair(s)
		partPubs[i] = pub
		partSec[i] = sec
	}
	longterms := GenDistSecret(s, nbParticipants, partSec, partPubs)
	randoms := GenDistSecret(s, nbParticipants, partSec, partPubs)
	dss, err := dss.NewDSS(s, partSec[0], partPubs, longterms[0], randoms[0], []byte("hello"), 4)
	require.NoError(t, err)
	partialSig, err := dss.PartialSig()
	require.NoError(t, err)

	msg := &msgPartialSig{
		gpa.BasicMessage{},
		s,
		partialSig,
	}

	newobj := new(msgPartialSig)
	newobj.suite = s
	rwutil.ReadWriteTest(t, msg, newobj)
}
