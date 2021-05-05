package consensus1imp

import (
	"testing"

	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/chain/mock_chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestConsensus1Impl(t *testing.T) {
	log := testlogger.NewLogger(t)
	chainID := coretypes.RandomChainID()
	chainCore := mock_chain.NewMockedChainCore(*chainID, log)
	mpool := mempool.New()
	consensus := New(chainCore)
}
