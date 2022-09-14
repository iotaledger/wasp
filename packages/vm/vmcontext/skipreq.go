package vmcontext

import (
	"fmt"
	"time"

	"golang.org/x/xerrors"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmexceptions"
)

const (
	// OffLedgerNonceStrictOrderTolerance how many steps back the nonce is considered too old
	// within this limit order of nonces is not checked
	OffLedgerNonceStrictOrderTolerance = 10_000
	// ExpiryUnlockSafetyWindowDuration creates safety window around time assumption,
	// the UTXO won't be consumed to avoid race conditions
	ExpiryUnlockSafetyWindowDuration  = 1 * time.Minute
	ExpiryUnlockSafetyWindowMilestone = 3
)

// earlyCheckReasonToSkip checks if request must be ignored without even modifying the state
func (vmctx *VMContext) earlyCheckReasonToSkip() error {
	if vmctx.task.AnchorOutput.StateIndex == 0 {
		if len(vmctx.task.AnchorOutput.NativeTokens) > 0 {
			return xerrors.New("can't init chain with native assets on the origin alias output")
		}
	} else {
		if len(vmctx.task.AnchorOutput.NativeTokens) > 0 {
			panic("inconsistency: native assets on the anchor output")
		}
	}

	if vmctx.task.MaintenanceModeEnabled &&
		vmctx.req.CallTarget().Contract != governance.Contract.Hname() {
		return fmt.Errorf("skipped due to maintenance mode")
	}

	if vmctx.req.IsOffLedger() {
		return vmctx.checkReasonToSkipOffLedger()
	}
	return vmctx.checkReasonToSkipOnLedger()
}

// checkReasonRequestProcessed checks if request ID is already in the blocklog
func (vmctx *VMContext) checkReasonRequestProcessed() error {
	reqid := vmctx.req.ID()
	var isProcessed bool
	vmctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		isProcessed = blocklog.MustIsRequestProcessed(vmctx.State(), &reqid)
	})
	if isProcessed {
		return xerrors.New("already processed")
	}
	return nil
}

func CheckNonce(req isc.OffLedgerRequest, maxAssumedNonce uint64) error {
	if maxAssumedNonce <= OffLedgerNonceStrictOrderTolerance {
		return nil
	}
	nonce := req.Nonce()
	if nonce < maxAssumedNonce-OffLedgerNonceStrictOrderTolerance {
		return fmt.Errorf("nonce %d is too old", nonce)
	}
	return nil
}

// checkReasonToSkipOffLedger checks reasons to skip off ledger request
func (vmctx *VMContext) checkReasonToSkipOffLedger() error {
	// first checks if it is already in backlog
	if err := vmctx.checkReasonRequestProcessed(); err != nil {
		return err
	}

	var maxAssumed uint64
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		// this is a replay protection measure for off-ledger requests assuming in the batch order of requests is random.
		// It is checking if nonce is not too old. See replay-off-ledger.md
		maxAssumed = accounts.GetMaxAssumedNonce(vmctx.State(), vmctx.req.SenderAccount())
	})

	return CheckNonce(vmctx.req.(isc.OffLedgerRequest), maxAssumed)
}

// checkReasonToSkipOnLedger check reasons to skip UTXO request
func (vmctx *VMContext) checkReasonToSkipOnLedger() error {
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

func (vmctx *VMContext) checkInternalOutput() error {
	// internal outputs are used for internal accounting of assets inside the chain. They are not interpreted as requests
	if vmctx.req.(isc.OnLedgerRequest).IsInternalUTXO(vmctx.ChainID()) {
		return xerrors.New("it is an internal output")
	}
	return nil
}

// checkReasonTimeLock checking timelock conditions based on time assumptions.
// VM must ensure that the UTXO can be unlocked
func (vmctx *VMContext) checkReasonTimeLock() error {
	timeLock := vmctx.req.(isc.OnLedgerRequest).Features().TimeLock()
	if !timeLock.IsZero() {
		if vmctx.finalStateTimestamp.Before(timeLock) {
			return xerrors.Errorf("can't be consumed due to lock until %v", vmctx.finalStateTimestamp)
		}
	}
	return nil
}

// checkReasonExpiry checking expiry conditions based on time assumptions.
// VM must ensure that the UTXO can be unlocked
func (vmctx *VMContext) checkReasonExpiry() error {
	expiry, _ := vmctx.req.(isc.OnLedgerRequest).Features().Expiry()

	if expiry.IsZero() {
		return nil
	}

	// Validate time window
	windowFrom := vmctx.finalStateTimestamp.Add(-ExpiryUnlockSafetyWindowDuration)
	windowTo := vmctx.finalStateTimestamp.Add(ExpiryUnlockSafetyWindowDuration)

	if expiry.After(windowFrom) && expiry.Before(windowTo) {
		return xerrors.Errorf("can't be consumed in the expire safety window close to %v", expiry)
	}

	// General unlock validation
	output, _ := vmctx.req.(isc.OnLedgerRequest).Output().(iotago.TransIndepIdentOutput)

	unlockable := output.UnlockableBy(vmctx.task.AnchorOutput.AliasID.ToAddress(), &iotago.ExternalUnlockParameters{
		ConfUnix: uint32(vmctx.finalStateTimestamp.Unix()),
	})

	if !unlockable {
		return xerrors.Errorf("can't be consumed, expiry: %v", expiry)
	}

	return nil
}

// checkReasonReturnAmount skipping anything with return amounts in this version. There's no risk to lose funds
func (vmctx *VMContext) checkReasonReturnAmount() error {
	if _, ok := vmctx.req.(isc.OnLedgerRequest).Features().ReturnAmount(); ok {
		return xerrors.New("return amount feature not supported in this version")
	}
	return nil
}
