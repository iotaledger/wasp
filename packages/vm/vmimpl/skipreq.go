package vmimpl

import (
	"errors"
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

// earlyCheckReasonToSkip checks if request must be ignored without even modifying the state
func (reqctx *requestContext) earlyCheckReasonToSkip(maintenanceMode bool) error {
	if maintenanceMode && reqctx.req.Message().Target.Contract != governance.Contract.Hname() {
		return errors.New("skipped due to maintenance mode")
	}

	if reqctx.req.IsOffLedger() {
		return reqctx.checkReasonToSkipOffLedger()
	}
	return reqctx.checkReasonToSkipOnLedger()
}

// checkReasonRequestProcessed checks if request ID is already in the blocklog
func (reqctx *requestContext) checkReasonRequestProcessed() error {
	reqid := reqctx.req.ID()
	isProcessed := lo.Must(blocklog.NewStateReaderFromChainState(reqctx.uncommittedState).IsRequestProcessed(reqid))
	if isProcessed {
		return errors.New("already processed")
	}
	return nil
}

// checkReasonToSkipOffLedger checks reasons to skip off ledger request
func (reqctx *requestContext) checkReasonToSkipOffLedger() error {
	if reqctx.vm.task.EstimateGasMode {
		return nil
	}
	offledgerReq := reqctx.req.(isc.OffLedgerRequest)
	if err := offledgerReq.VerifySignature(); err != nil {
		return err
	}
	senderAccount := offledgerReq.SenderAccount()

	reqNonce := offledgerReq.Nonce()
	var expectedNonce uint64
	if evmAgentID, ok := senderAccount.(*isc.EthereumAddressAgentID); ok {
		expectedNonce = evmimpl.Nonce(evm.Contract.StateSubrealm(reqctx.uncommittedState), evmAgentID.EthAddress())
	} else {
		expectedNonce = reqctx.accountsStateWriter(false).AccountNonce(senderAccount)
	}
	if reqNonce != expectedNonce {
		return fmt.Errorf(
			"invalid nonce (%s): expected %d, got %d",
			offledgerReq.SenderAccount(), expectedNonce, reqNonce,
		)
	}

	if gasPrice := isc.RequestGasPrice(reqctx.req); gasPrice != nil {
		if err := evmutil.CheckGasPrice(gasPrice, reqctx.vm.chainInfo.GasFeePolicy); err != nil {
			return err
		}
	}
	return nil
}

// checkReasonToSkipOnLedger check reasons to skip UTXO request
func (reqctx *requestContext) checkReasonToSkipOnLedger() error {
	if err := reqctx.checkReasonRequestProcessed(); err != nil {
		return err
	}
	return nil
}

// checkReasonTimeLock checking timelock conditions based on time assumptions.
// VM must ensure that the UTXO can be unlocked
/* func (reqctx *requestContext) checkReasonTimeLock() error {
	timeLock := reqctx.req.(isc.OnLedgerRequest).Features().TimeLock()
	if !timeLock.IsZero() {
		if reqctx.vm.task.FinalStateTimestamp().Before(timeLock) {
			return fmt.Errorf("can't be consumed due to lock until %v", reqctx.vm.task.FinalStateTimestamp())
		}
	}
	return nil
}*/

// checkReasonExpiry checking expiry conditions based on time assumptions.
// VM must ensure that the UTXO can be unlocked
/* func (reqctx *requestContext) checkReasonExpiry() error {
	expiry, _ := reqctx.req.(isc.OnLedgerRequest).Features().Expiry()

	if expiry.IsZero() {
		return nil
	}

	// Validate time window
	finalStateTimestamp := reqctx.vm.task.FinalStateTimestamp()
	windowFrom := finalStateTimestamp.Add(-ExpiryUnlockSafetyWindowDuration)
	windowTo := finalStateTimestamp.Add(ExpiryUnlockSafetyWindowDuration)

	if expiry.After(windowFrom) && expiry.Before(windowTo) {
		return fmt.Errorf("can't be consumed in the expire safety window close to %v", expiry)
	}

	// General unlock validation
	output, _ := reqctx.req.(isc.OnLedgerRequest).Output().(iotago.TransIndepIdentOutput)

	unlockable := output.UnlockableBy(reqctx.vm.task.AnchorOutput.AliasID.ToAddress(), &iotago.ExternalUnlockParameters{
		ConfUnix: uint32(finalStateTimestamp.Unix()),
	})

	if !unlockable {
		return fmt.Errorf("can't be consumed, expiry: %v", expiry)
	}

	return nil
}
*/

// checkReasonReturnAmount skipping anything with return amounts in this version. There's no risk to lose funds
/*func (reqctx *requestContext) checkReasonReturnAmount() error {
	if _, ok := reqctx.req.(isc.OnLedgerRequest).Features().ReturnAmount(); ok {
		return errors.New("return amount feature not supported in this version")
	}
	return nil
}*/
