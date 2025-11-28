// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsign

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/dss"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/tcrypto"
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

	msg, err := NewMsgPartialSig(partialSig)
	require.NoError(t, err)

	msgEnc := bcs.MustMarshal(&msg)
	msgDec := bcs.MustUnmarshal[MsgPartialSig](msgEnc)
	require.Equal(t, msg, msgDec)
	require.Equal(t, partialSig, lo.Must(msgDec.PartialSig(s)))
}
