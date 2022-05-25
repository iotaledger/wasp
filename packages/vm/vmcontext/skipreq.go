package vmcontext

import (
	"fmt"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmexceptions"
	"golang.org/x/xerrors"
)

const (
	// OffLedgerNonceStrictOrderTolerance how many steps back the nonce is considered too old
	// within this limit order of nonces is not checked
	OffLedgerNonceStrictOrderTolerance = 10000
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

	var err error
	if vmctx.req.IsOffLedger() {
		err = vmctx.checkReasonToSkipOffLedger()
	} else {
		err = vmctx.checkReasonToSkipOnLedger()
	}
	return err
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

	nonce := vmctx.req.AsOffLedger().Nonce()
	vmctx.Debugf("vmctx.validateRequest - nonce check - maxAssumed: %d, tolerance: %d, request nonce: %d ",
		maxAssumed, OffLedgerNonceStrictOrderTolerance, nonce)

	if maxAssumed < OffLedgerNonceStrictOrderTolerance {
		return nil
	}
	if nonce <= maxAssumed-OffLedgerNonceStrictOrderTolerance {
		return fmt.Errorf("nonce %d is too old", nonce)
	}
	return nil
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
	if vmctx.req.AsOnLedger().IsInternalUTXO(vmctx.ChainID()) {
		return xerrors.New("it is an internal output")
	}
	return nil
}

// checkReasonTimeLock checking timelock conditions based on time assumptions.
// VM must ensure that the UTXO can be unlocked
func (vmctx *VMContext) checkReasonTimeLock() error {
	lock := vmctx.req.AsOnLedger().Features().TimeLock()
	if lock != nil {
		if !lock.Time.IsZero() {
			if vmctx.finalStateTimestamp.Before(lock.Time) {
				return xerrors.Errorf("can't be consumed due to lock until %v", vmctx.finalStateTimestamp)
			}
		}
		if lock.MilestoneIndex != 0 && vmctx.task.TimeAssumption.MilestoneIndex < lock.MilestoneIndex {
			return xerrors.Errorf("can't be consumed due to lock until milestone index #%v", vmctx.task.TimeAssumption.MilestoneIndex)
		}
	}
	return nil
}

// checkReasonExpiry checking expiry conditions based on time assumptions.
// VM must ensure that the UTXO can be unlocked
func (vmctx *VMContext) checkReasonExpiry() error {
	expiry, _ := vmctx.req.AsOnLedger().Features().Expiry()

	if expiry == nil {
		return nil
	}

	// Validate time window
	windowFrom := vmctx.finalStateTimestamp.Add(-ExpiryUnlockSafetyWindowDuration)
	windowTo := vmctx.finalStateTimestamp.Add(ExpiryUnlockSafetyWindowDuration)

	if expiry.Time.After(windowFrom) && expiry.Time.Before(windowTo) {
		return xerrors.Errorf("can't be consumed in the expire safety window close to v", expiry.Time)
	}

	// Validate milestone window
	milestoneFrom := vmctx.task.TimeAssumption.MilestoneIndex - ExpiryUnlockSafetyWindowMilestone
	milestoneTo := vmctx.task.TimeAssumption.MilestoneIndex + ExpiryUnlockSafetyWindowMilestone

	if milestoneFrom <= expiry.MilestoneIndex && expiry.MilestoneIndex <= milestoneTo {
		return xerrors.Errorf("can't be consumed in the expire safety window between milestones #%d and #%d",
			milestoneFrom, milestoneTo)
	}

	// General unlock validation
	output, _ := vmctx.req.AsOnLedger().Output().(iotago.TransIndepIdentOutput)

	unlockable := output.UnlockableBy(vmctx.task.AnchorOutput.AliasID.ToAddress(), &iotago.ExternalUnlockParameters{
		ConfUnix:    uint32(vmctx.finalStateTimestamp.Unix()),
		ConfMsIndex: vmctx.task.TimeAssumption.MilestoneIndex,
	})

	if !unlockable {
		return xerrors.Errorf("can't be consumed", expiry.Time)
	}

	return nil
}

// checkReasonReturnAmount skipping anything with return amounts in this version. There's no risk to lose funds
func (vmctx *VMContext) checkReasonReturnAmount() error {
	if _, ok := vmctx.req.AsOnLedger().Features().ReturnAmount(); ok {
		return xerrors.New("return amount feature not supported in this version")
	}
	return nil
}
