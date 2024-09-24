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
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func TestBatchProposal1Serialization(t *testing.T) {
	var reqRefs []*isc.RequestRef
	for i := uint64(0); i < 5; i++ {
		req := isc.NewOffLedgerRequest(isc.RandomChainID(), isc.NewMessage(3, 14), i, 200).Sign(cryptolib.NewKeyPair())
		reqRefs = append(reqRefs, &isc.RequestRef{
			ID:   req.ID(),
			Hash: hashing.PseudoRandomHash(nil),
		})
	}

	anchor := iscmove.RandomAnchor()

	// TODO: how to properly generate digest?
	var digest sui.Base58
	_, err := rand.Read(digest)
	require.NoError(t, err)

	anchorRef := iscmove.AnchorWithRef{
		ObjectRef: sui.ObjectRef{
			ObjectID: &anchor.ID,
			Version:  13,
			Digest:   &digest,
		},
		Object: &anchor,
	}

	batchProposal := NewBatchProposal(10, &anchorRef, util.NewFixedSizeBitVector(11), time.Now(), isc.NewRandomAgentID(), reqRefs)

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
