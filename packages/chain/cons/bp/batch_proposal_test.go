// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bp

import (
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/isctest"
	"github.com/iotaledger/wasp/v2/packages/parameters/parameterstest"
	"github.com/iotaledger/wasp/v2/packages/util"
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

	stateAnchor := isctest.RandomStateAnchor()
	batchProposal := NewBatchProposal(
		10,
		&stateAnchor,
		util.NewFixedSizeBitVector(11),
		nil,
		time.Now(),
		isctest.NewRandomAgentID(),
		reqRefs,
		[]*coin.CoinWithRef{{
			Type:  coin.BaseTokenType,
			Value: coin.Value(100),
			Ref:   iotatest.RandomObjectRef(),
		}},
		parameterstest.L1Mock,
	)

	bpEncoded := lo.Must1(bcs.Marshal(batchProposal))
	bpDecoded, err := bcs.Unmarshal[BatchProposal](bpEncoded)
	require.NoError(t, err)
	require.Equal(t, batchProposal.nodeIndex, bpDecoded.nodeIndex)
	require.Equal(t, batchProposal.baseAliasOutput, bpDecoded.baseAliasOutput)
	require.Equal(t, batchProposal.dssIndexProposal, bpDecoded.dssIndexProposal)
	require.Equal(t, batchProposal.timeData.UnixNano(), bpDecoded.timeData.UnixNano())
	require.Equal(t, batchProposal.validatorFeeDestination, bpDecoded.validatorFeeDestination)
	require.Equal(t, batchProposal.requestRefs, bpDecoded.requestRefs)
	require.Equal(t, batchProposal.gasCoins, bpDecoded.gasCoins)
	require.Equal(t, batchProposal.l1params, bpDecoded.l1params)
}
