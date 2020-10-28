package runvm

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"github.com/iotaledger/wasp/packages/vm"
)

// createVMContext:
// 1. creates state block in the tx builder, including moving the SC token
// 2. handles request tokens by moving them either to the
// reward address or sending it back to the requestor
// All request tokens are handled for the whole batch
func createVMContext(ctx *vm.VMTask, txb *txbuilder.Builder) (*vm.VMContext, error) {
	// create state block and move smart contract token
	if err := txb.CreateStateBlock(ctx.Color); err != nil {
		ctx.Log.Errorf("createVMContext: %v\nDump txbuilder accounts:\n%s\n", err, txb.Dump())
		return nil, fmt.Errorf("createVMContext: %v", err)
	}

	// handle request tokens.
	// recolor request tokens back to iota color
	// if node rewards are enabled, send request tokens to it. Otherwise send them to the request originator
	nodeRewardsEnabled := ctx.RewardAddress[0] != 0 && ctx.MinimumReward > 0

	var targetAddress address.Address

	for _, reqRef := range ctx.Requests {
		// if rewards enabled, request tokens are erased to the reward address
		// otherwise all erased tokens (iotas) are returned back to teh corresponding
		// addresses
		if nodeRewardsEnabled {
			targetAddress = ctx.RewardAddress
		} else {
			targetAddress = *reqRef.Tx.Sender()
		}
		reqTxId := reqRef.Tx.ID()
		reqColor := (balance.Color)(reqTxId)
		// one request token is uncolored back to iotas for each request
		if err := txb.EraseColor(targetAddress, reqColor, 1); err != nil {
			ctx.Log.Errorf("createVMContext: %v\nDump txbuilder accounts:\n%s\n", err, txb.Dump())
			return nil, fmt.Errorf("createVMContext: %v", err)
		}
		ctx.Log.Debugf("$$$$$$$ erased 1 request token color %s to addr %s. Remains %d",
			reqColor.String(), targetAddress.String(),
			txb.GetInputBalanceFromTransaction(reqColor, reqTxId))
	}

	vmctx := &vm.VMContext{
		ContractID:         coretypes.NewContractID(ctx.ChainID, 0), // TODO
		OwnerAddress:       ctx.OwnerAddress,
		RewardAddress:      ctx.RewardAddress,
		ProgramHash:        ctx.ProgramHash,
		MinimumReward:      ctx.MinimumReward,
		NodeRewardsEnabled: nodeRewardsEnabled,
		Entropy:            *hashing.HashData(ctx.Entropy[:]), // mutates deterministically
		TxBuilder:          txb,                               // mutates when tokens are moved
		Timestamp:          ctx.Timestamp,                     // mutates by increments of 1 nanosecond
		VirtualState:       ctx.VirtualState.Clone(),
		Log:                ctx.Log,
	}
	return vmctx, nil
}

// handleNodeRewards return true if to continue with request processing
// rewards are "rewards for node", so smart contract sending request to itself might need
// to pay rewards too
func handleNodeRewards(vmctx *vm.VMContext) bool {
	if !vmctx.NodeRewardsEnabled {
		// nothing to do
		return true
	}
	var err error

	reqTxId := vmctx.RequestRef.Tx.ID()
	// determining how many iotas have been left in the request transaction
	availableIotas := vmctx.TxBuilder.GetInputBalanceFromTransaction(balance.ColorIOTA, reqTxId)

	var proceed bool
	// taking into account 1 request token which will be recolored back to iota
	// and will be send to the node reward address (if enabled)
	var sendToRewardAddress int64
	if availableIotas+1 >= vmctx.MinimumReward {
		sendToRewardAddress = vmctx.MinimumReward - 1
		proceed = true
	} else {
		sendToRewardAddress = availableIotas
		// if reward is not enough, the state update will be empty, i.e. NOP (the fee will be taken)
		proceed = false
	}
	err = vmctx.TxBuilder.MoveToAddressFromTransaction(vmctx.RewardAddress, balance.ColorIOTA, sendToRewardAddress, reqTxId)

	if err != nil {
		vmctx.Log.Error("can't move reward tokens: %v", err)
		proceed = false
	}
	return proceed
}
