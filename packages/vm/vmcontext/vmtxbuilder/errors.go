package vmtxbuilder

import (
	"github.com/iotaledger/wasp/packages/vm/core/errors"
)

// error codes used for handled panics
var (
	ErrOverflow, _                             = errors.RegisterGlobalError(errors.GLOBAL_ERROR_START+1, "overflow")
	ErrNotEnoughIotaBalance, _                 = errors.RegisterGlobalError(errors.GLOBAL_ERROR_START+1, "not enough iota balance")
	ErrNotEnoughNativeAssetBalance, _          = errors.RegisterGlobalError(errors.GLOBAL_ERROR_START+1, "not enough native assets balance")
	ErrCreateFoundryMaxSupplyMustBePositive, _ = errors.RegisterGlobalError(errors.GLOBAL_ERROR_START+1, "max supply must be positive")
	ErrCreateFoundryMaxSupplyTooBig, _         = errors.RegisterGlobalError(errors.GLOBAL_ERROR_START+1, "max supply is too big")
	ErrFoundryDoesNotExist, _                  = errors.RegisterGlobalError(errors.GLOBAL_ERROR_START+1, "foundry does not exist")
	ErrCantModifySupplyOfTheToken, _           = errors.RegisterGlobalError(errors.GLOBAL_ERROR_START+1, "supply of the token is not controlled by the chain")
	ErrNativeTokenSupplyOutOffBounds, _        = errors.RegisterGlobalError(errors.GLOBAL_ERROR_START+1, "token supply is out of bounds")
	ErrFatalTxBuilderNotBalanced, _            = errors.RegisterGlobalError(errors.GLOBAL_ERROR_START+1, "fatal: tx builder is not balanced")
	ErrInconsistentL2LedgerWithL1TxBuilder, _  = errors.RegisterGlobalError(errors.GLOBAL_ERROR_START+1, "fatal: L2 ledger is not consistent with the L1 tx builder")
	ErrCantDestroyFoundryBeingCreated, _       = errors.RegisterGlobalError(errors.GLOBAL_ERROR_START+1, "can't destroy foundry which is being created")
)
