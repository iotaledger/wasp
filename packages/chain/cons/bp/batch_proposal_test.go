// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bp

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmovetest"
	sui2 "github.com/iotaledger/wasp/clients/iota-go/sui"
	"github.com/iotaledger/wasp/clients/iota-go/sui/suitest"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/isctest"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestBatchProposal1Serialization(t *testing.T) {
	var reqRefs []*isc.RequestRef
	for i := uint64(0); i < 5; i++ {
		req := isc.NewOffLedgerRequest(isctest.RandomChainID(), isc.NewMessage(3, 14), i, 200).Sign(cryptolib.NewKeyPair())
		reqRefs = append(reqRefs, &isc.RequestRef{
			ID:   req.ID(),
			Hash: hashing.PseudoRandomHash(nil),
		})
	}

	anchor := iscmovetest.RandomAnchor()

	// TODO: how to properly generate digest?
	var digest sui2.Base58
	_, err := rand.Read(digest)
	require.NoError(t, err)

	stateAnchor := isc.NewStateAnchor(&iscmove.AnchorWithRef{
		ObjectRef: sui2.ObjectRef{
			ObjectID: &anchor.ID,
			Version:  13,
			Digest:   &digest,
		},
		Object: &anchor,
	}, cryptolib.NewEmptyAddress(), *suitest.RandomAddress())

	batchProposal := NewBatchProposal(10, &stateAnchor, util.NewFixedSizeBitVector(11), time.Now(), isctest.NewRandomAgentID(), reqRefs)

	bpEncoded := lo.Must1(bcs.Marshal(batchProposal))
	bpDecoded, err := bcs.Unmarshal[BatchProposal](bpEncoded)
	require.NoError(t, err)
	require.Equal(t, batchProposal.nodeIndex, bpDecoded.nodeIndex)
	require.Equal(t, batchProposal.baseAliasOutput, bpDecoded.baseAliasOutput)
	require.Equal(t, batchProposal.dssIndexProposal, bpDecoded.dssIndexProposal)
	require.Equal(t, batchProposal.timeData.UnixNano(), bpDecoded.timeData.UnixNano())
	require.Equal(t, batchProposal.validatorFeeDestination, bpDecoded.validatorFeeDestination)
	require.Equal(t, batchProposal.requestRefs, bpDecoded.requestRefs)
}
