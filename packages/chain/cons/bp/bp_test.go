package bp_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/v2/packages/chain/cons/bp"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/isctest"
	"github.com/iotaledger/wasp/v2/packages/parameters/parameterstest"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

func TestOffLedgerOrdering(t *testing.T) {
	log := testlogger.NewLogger(t)
	nodeIDs := gpa.MakeTestNodeIDs(1)
	//
	// Produce an anchor.
	chainID := isc.ChainIDFromObjectID(*iotatest.RandomObjectRef().ObjectID)
	anchor0 := isctest.RandomStateAnchor()

	// Create some requests.
	senderKP := cryptolib.NewKeyPair()
	contract := governance.Contract.Hname()
	entryPoint := governance.FuncAddCandidateNode.Hname()
	gasBudget := gas.LimitsDefault.MaxGasPerRequest
	r0 := isc.NewOffLedgerRequest(isc.NewMessage(contract, entryPoint, nil), 0, gasBudget).Sign(senderKP)
	r1 := isc.NewOffLedgerRequest(isc.NewMessage(contract, entryPoint, nil), 1, gasBudget).Sign(senderKP)
	r2 := isc.NewOffLedgerRequest(isc.NewMessage(contract, entryPoint, nil), 2, gasBudget).Sign(senderKP)
	r3 := isc.NewOffLedgerRequest(isc.NewMessage(contract, entryPoint, nil), 3, gasBudget).Sign(senderKP)
	rs := []isc.Request{r3, r1, r0, r2} // Out of order.
	//
	// Construct the batch proposal, and aggregate it.
	bp0 := bp.NewBatchProposal(
		0,
		&anchor0,
		util.NewFixedSizeBitVector(1).SetBits([]int{0}),
		nil,
		time.Now(),
		isctest.NewRandomAgentID(),
		isc.RequestRefsFromRequests(rs),
		[]*coin.CoinWithRef{{
			Type:  coin.BaseTokenType,
			Value: coin.Value(100),
			Ref:   iotatest.RandomObjectRef(),
		}},
		parameterstest.L1Mock,
	)
	bp0.Bytes()
	abpInputs := map[gpa.NodeID][]byte{
		nodeIDs[0]: bp0.Bytes(),
	}
	abp := bp.AggregateBatchProposals(abpInputs, nodeIDs, 0, log)
	require.NotNil(t, abp)
	require.Equal(t, len(abp.DecidedRequestRefs()), len(rs))
	//
	// ...
	rndSeed := rand.New(rand.NewSource(rand.Int63()))
	randomness := hashing.PseudoRandomHash(rndSeed)
	sortedRS := abp.OrderedRequests(rs, randomness)

	for i := range sortedRS {
		for j := range sortedRS {
			if i >= j {
				continue
			}
			oflI, okI := sortedRS[i].(isc.OffLedgerRequest)
			oflJ, okJ := sortedRS[j].(isc.OffLedgerRequest)
			if !okI || !okJ {
				continue
			}
			if !oflI.SenderAccount().Equals(oflJ.SenderAccount()) {
				continue
			}
			require.Less(t, oflI.Nonce(), oflJ.Nonce(), "i=%v, j=%v", i, j)
		}
	}
}
