package vmimpl

import (
	"errors"
	"fmt"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/vmexceptions"
)

const (
	// ExpiryUnlockSafetyWindowDuration creates safety window around time assumption,
	// the UTXO won't be consumed to avoid race conditions
	ExpiryUnlockSafetyWindowDuration = 1 * time.Minute
)

// earlyCheckReasonToSkip checks if request must be ignored without even modifying the state
func (vmctx *vmContext) earlyCheckReasonToSkip() error {
	if vmctx.task.AnchorOutput.StateIndex == 0 {
		if len(vmctx.task.AnchorOutput.NativeTokens) > 0 {
			return errors.New("can't init chain with native assets on the origin alias output")
		}
	} else {
		if len(vmctx.task.AnchorOutput.NativeTokens) > 0 {
			panic("inconsistency: native assets on the anchor output")
		}
	}

	if vmctx.task.MaintenanceModeEnabled &&
		vmctx.reqCtx.req.CallTarget().Contract != governance.Contract.Hname() {
		return errors.New("skipped due to maintenance mode")
	}

	if vmctx.reqCtx.req.IsOffLedger() {
		return vmctx.checkReasonToSkipOffLedger()
	}
	return vmctx.checkReasonToSkipOnLedger()
}

// checkReasonRequestProcessed checks if request ID is already in the blocklog
func (vmctx *vmContext) checkReasonRequestProcessed() error {
	reqid := vmctx.reqCtx.req.ID()
	var isProcessed bool
	vmctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		isProcessed = blocklog.MustIsRequestProcessed(s, reqid)
	})
	if isProcessed {
		return errors.New("already processed")
	}
	return nil
}

// checkReasonToSkipOffLedger checks reasons to skip off ledger request
func (vmctx *vmContext) checkReasonToSkipOffLedger() error {
	// first checks if it is already in backlog
	if err := vmctx.checkReasonRequestProcessed(); err != nil {
		return err
	}

	// skip ISC nonce check for EVM requests
	senderAccount := vmctx.reqCtx.req.SenderAccount()
	if senderAccount.Kind() == isc.AgentIDKindEthereumAddress {
		return nil
	}

	var nonceErr error
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		nonceErr = accounts.CheckNonce(s, senderAccount, vmctx.reqCtx.req.(isc.OffLedgerRequest).Nonce())
	})
	return nonceErr
}

// checkReasonToSkipOnLedger check reasons to skip UTXO request
func (vmctx *vmContext) checkReasonToSkipOnLedger() error {
	if err := vmctx.checkInternalOutput(); err != nil {
		return err
	}
	if err := vmctx.checkReasonReturnAmount(); err != nil {
		return err
	}
	if err := vmctx.checkReasonTimeLock(); err != nil {
		return err
	}
	if err := vmctx.checkReasonExpiry(); err != nil {
		return err
	}
	if vmctx.txbuilder.InputsAreFull() {
		return vmexceptions.ErrInputLimitExceeded
	}
	if err := vmctx.checkReasonRequestProcessed(); err != nil {
		return err
	}
	return nil
}

func (vmctx *vmContext) checkInternalOutput() error {
	// internal outputs are used for internal accounting of assets inside the chain. They are not interpreted as requests
	if vmctx.reqCtx.req.(isc.OnLedgerRequest).IsInternalUTXO(vmctx.ChainID()) {
		return errors.New("it is an internal output")
	}
	return nil
}

// checkReasonTimeLock checking timelock conditions based on time assumptions.
// VM must ensure that the UTXO can be unlocked
func (vmctx *vmContext) checkReasonTimeLock() error {
	timeLock := vmctx.reqCtx.req.(isc.OnLedgerRequest).Features().TimeLock()
	if !timeLock.IsZero() {
		if vmctx.task.FinalStateTimestamp().Before(timeLock) {
			return fmt.Errorf("can't be consumed due to lock until %v", vmctx.task.FinalStateTimestamp())
		}
	}
	return nil
}

// checkReasonExpiry checking expiry conditions based on time assumptions.
// VM must ensure that the UTXO can be unlocked
func (vmctx *vmContext) checkReasonExpiry() error {
	expiry, _ := vmctx.reqCtx.req.(isc.OnLedgerRequest).Features().Expiry()

	if expiry.IsZero() {
		return nil
	}

	// Validate time window
	finalStateTimestamp := vmctx.task.FinalStateTimestamp()
	windowFrom := finalStateTimestamp.Add(-ExpiryUnlockSafetyWindowDuration)
	windowTo := finalStateTimestamp.Add(ExpiryUnlockSafetyWindowDuration)

	if expiry.After(windowFrom) && expiry.Before(windowTo) {
		return fmt.Errorf("can't be consumed in the expire safety window close to %v", expiry)
	}

	// General unlock validation
	output, _ := vmctx.reqCtx.req.(isc.OnLedgerRequest).Output().(iotago.TransIndepIdentOutput)

	unlockable := output.UnlockableBy(vmctx.task.AnchorOutput.AliasID.ToAddress(), &iotago.ExternalUnlockParameters{
		ConfUnix: uint32(finalStateTimestamp.Unix()),
	})

	if !unlockable {
		return fmt.Errorf("can't be consumed, expiry: %v", expiry)
	}

	return nil
}

// checkReasonReturnAmount skipping anything with return amounts in this version. There's no risk to lose funds
func (vmctx *vmContext) checkReasonReturnAmount() error {
	if _, ok := vmctx.reqCtx.req.(isc.OnLedgerRequest).Features().ReturnAmount(); ok {
		return errors.New("return amount feature not supported in this version")
	}
	return nil
}
