package vmtxbuilder

import (
	"github.com/iotaledger/wasp/packages/vm/core/errors"
)

// error codes used for handled panics
var (
	ErrOverflow                             = errors.RegisterGlobalError(errors.GlobalErrorStart+1, "overflow")
	ErrNotEnoughIotaBalance                 = errors.RegisterGlobalError(errors.GlobalErrorStart+2, "not enough iota balance")
	ErrNotEnoughNativeAssetBalance          = errors.RegisterGlobalError(errors.GlobalErrorStart+3, "not enough native assets balance")
	ErrCreateFoundryMaxSupplyMustBePositive = errors.RegisterGlobalError(errors.GlobalErrorStart+4, "max supply must be positive")
	ErrCreateFoundryMaxSupplyTooBig         = errors.RegisterGlobalError(errors.GlobalErrorStart+5, "max supply is too big")
	ErrFoundryDoesNotExist                  = errors.RegisterGlobalError(errors.GlobalErrorStart+6, "foundry does not exist")
	ErrCantModifySupplyOfTheToken           = errors.RegisterGlobalError(errors.GlobalErrorStart+7, "supply of the token is not controlled by the chain")
	ErrNativeTokenSupplyOutOffBounds        = errors.RegisterGlobalError(errors.GlobalErrorStart+8, "token supply is out of bounds")
	ErrFatalTxBuilderNotBalanced            = errors.RegisterGlobalError(errors.GlobalErrorStart+9, "fatal: tx builder is not balanced")
	ErrInconsistentL2LedgerWithL1TxBuilder  = errors.RegisterGlobalError(errors.GlobalErrorStart+10, "fatal: L2 ledger is not consistent with the L1 tx builder")
	ErrCantDestroyFoundryBeingCreated       = errors.RegisterGlobalError(errors.GlobalErrorStart+11, "can't destroy foundry which is being created")
	ErrBasicMessageError                    = errors.RegisterGlobalError(errors.GlobalErrorStart+12, "%v")
)
