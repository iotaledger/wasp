package vmimpl

import (
	"errors"
	"fmt"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/vmexceptions"
)

const (
	// ExpiryUnlockSafetyWindowDuration creates safety window around time assumption,
	// the UTXO won't be consumed to avoid race conditions
	ExpiryUnlockSafetyWindowDuration = 1 * time.Minute
)

// earlyCheckReasonToSkip checks if request must be ignored without even modifying the state
func (reqctx *requestContext) earlyCheckReasonToSkip(maintenanceMode bool) error {
	if reqctx.vm.task.AnchorOutput.StateIndex == 0 {
		if len(reqctx.vm.task.AnchorOutput.NativeTokens) > 0 {
			return errors.New("can't init chain with native assets on the origin alias output")
		}
	} else {
		if len(reqctx.vm.task.AnchorOutput.NativeTokens) > 0 {
			panic("inconsistency: native assets on the anchor output")
		}
	}

	if maintenanceMode &&
		reqctx.req.CallTarget().Contract != governance.Contract.Hname() {
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
	var isProcessed bool
	withContractState(reqctx.uncommittedState, blocklog.Contract, func(s kv.KVStore) {
		isProcessed = blocklog.MustIsRequestProcessed(s, reqid)
	})
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
		withContractState(reqctx.uncommittedState, evm.Contract, func(s kv.KVStore) {
			expectedNonce = evmimpl.Nonce(s, evmAgentID.EthAddress())
		})
	} else {
		withContractState(reqctx.uncommittedState, accounts.Contract, func(s kv.KVStore) {
			expectedNonce = accounts.AccountNonce(
				s,
				senderAccount,
				reqctx.ChainID(),
			)
		})
	}
	if reqNonce != expectedNonce {
		return fmt.Errorf(
			"invalid nonce (%s): expected %d, got %d",
			offledgerReq.SenderAccount(), expectedNonce, reqNonce,
		)
	}

	if evmTx := offledgerReq.EVMTransaction(); evmTx != nil {
		if err := evmutil.CheckGasPrice(evmTx, reqctx.vm.chainInfo.GasFeePolicy); err != nil {
			return err
		}
	}
	return nil
}

// checkReasonToSkipOnLedger check reasons to skip UTXO request
func (reqctx *requestContext) checkReasonToSkipOnLedger() error {
	if err := reqctx.checkInternalOutput(); err != nil {
		return err
	}
	if err := reqctx.checkReasonReturnAmount(); err != nil {
		return err
	}
	if err := reqctx.checkReasonTimeLock(); err != nil {
		return err
	}
	if err := reqctx.checkReasonExpiry(); err != nil {
		return err
	}
	if reqctx.vm.txbuilder.InputsAreFull() {
		return vmexceptions.ErrInputLimitExceeded
	}
	if err := reqctx.checkReasonRequestProcessed(); err != nil {
		return err
	}
	return nil
}

func (reqctx *requestContext) checkInternalOutput() error {
	// internal outputs are used for internal accounting of assets inside the chain. They are not interpreted as requests
	if reqctx.req.(isc.OnLedgerRequest).IsInternalUTXO(reqctx.ChainID()) {
		return errors.New("it is an internal output")
	}
	return nil
}

// checkReasonTimeLock checking timelock conditions based on time assumptions.
// VM must ensure that the UTXO can be unlocked
func (reqctx *requestContext) checkReasonTimeLock() error {
	timeLock := reqctx.req.(isc.OnLedgerRequest).Features().TimeLock()
	if !timeLock.IsZero() {
		if reqctx.vm.task.FinalStateTimestamp().Before(timeLock) {
			return fmt.Errorf("can't be consumed due to lock until %v", reqctx.vm.task.FinalStateTimestamp())
		}
	}
	return nil
}

// checkReasonExpiry checking expiry conditions based on time assumptions.
// VM must ensure that the UTXO can be unlocked
func (reqctx *requestContext) checkReasonExpiry() error {
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

// checkReasonReturnAmount skipping anything with return amounts in this version. There's no risk to lose funds
func (reqctx *requestContext) checkReasonReturnAmount() error {
	if _, ok := reqctx.req.(isc.OnLedgerRequest).Features().ReturnAmount(); ok {
		return errors.New("return amount feature not supported in this version")
	}
	return nil
}
