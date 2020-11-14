package vmcontext

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

// context for one request
type VMContext struct {
	// same for the block
	chainID            coretypes.ChainID
	processors         *processors.ProcessorCache
	rewardAddress      address.Address
	minimumReward      int64
	nodeRewardsEnabled bool
	txBuilder          *txbuilder.Builder // mutated
	saveTxBuilder      *txbuilder.Builder // for rollback
	virtualState       state.VirtualState // mutated
	log                *logger.Logger
	// request context
	entropy     hashing.HashValue // mutates with each request
	reqRef      sctransaction.RequestRef
	timestamp   int64
	stateUpdate state.StateUpdate // mutated
	callStack   []*callContext
}

type callContext struct {
	isRequestContext bool                      // is called from the request (true) or from another SC (false)
	caller           coretypes.AgentID         // calling agent
	contract         coretypes.Hname           // called contract
	params           codec.ImmutableCodec      // params passed
	transfer         coretypes.ColoredBalances // transfer passed
}

// NewVMContext:
// - creates state block in the tx builder, including moving the SC token
// - handles request tokens by moving them either to the
// reward address or sending it back to the requester
// All request tokens are handled for the whole block
func NewVMContext(task *vm.VMTask, txb *txbuilder.Builder) (*VMContext, error) {
	// create state block and move smart contract token
	if err := txb.CreateStateSection(task.Color); err != nil {
		task.Log.Errorf("createVMContext: %v\nDump txbuilder accounts:\n%s\n", err, txb.Dump())
		return nil, fmt.Errorf("createVMContext: %v", err)
	}

	// handle request tokens.
	// recolor request tokens back to iota color
	// if node rewards are enabled, send request tokens to it. Otherwise send them to the request originator
	nodeRewardsEnabled := task.RewardAddress[0] != 0 && task.MinimumReward > 0

	var targetAddress address.Address

	for _, reqRef := range task.Requests {
		// if rewards enabled, request tokens are erased to the reward address
		// otherwise all erased tokens (iotas) are returned back to teh corresponding
		// addresses
		if nodeRewardsEnabled {
			targetAddress = task.RewardAddress
		} else {
			targetAddress = *reqRef.Tx.Sender()
		}
		reqTxId := reqRef.Tx.ID()
		reqColor := (balance.Color)(reqTxId)
		// one request token is uncolored back to iotas for each request
		if err := txb.EraseColor(targetAddress, reqColor, 1); err != nil {
			task.Log.Errorf("createVMContext: %v\nDump txbuilder accounts:\n%s\n", err, txb.Dump())
			return nil, fmt.Errorf("createVMContext: %v", err)
		}
		task.Log.Debugf("$$$$$$$ erased 1 request token color %s to addr %s. Remains %d",
			reqColor.String(), targetAddress.String(),
			txb.GetInputBalanceFromTransaction(reqColor, reqTxId))
	}
	ret := &VMContext{
		processors:         task.Processors,
		chainID:            task.ChainID,
		rewardAddress:      task.RewardAddress,
		minimumReward:      task.MinimumReward,
		nodeRewardsEnabled: nodeRewardsEnabled,
		txBuilder:          txb,
		virtualState:       task.VirtualState.Clone(),
		log:                task.Log,
		entropy:            task.Entropy,
		callStack:          make([]*callContext, 0),
	}
	return ret, nil
}
