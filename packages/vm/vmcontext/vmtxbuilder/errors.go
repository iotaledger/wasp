package vmtxbuilder

import (
	"github.com/iotaledger/wasp/packages/vm/core/errors/commonerrors"
)

// error codes used for handled panics
var (
	ErrOverflow                             = commonerrors.RegisterGlobalError("overflow")
	ErrNotEnoughIotaBalance                 = commonerrors.RegisterGlobalError("not enough iota balance")
	ErrNotEnoughNativeAssetBalance          = commonerrors.RegisterGlobalError("not enough native assets balance")
	ErrCreateFoundryMaxSupplyMustBePositive = commonerrors.RegisterGlobalError("max supply must be positive")
	ErrCreateFoundryMaxSupplyTooBig         = commonerrors.RegisterGlobalError("max supply is too big")
	ErrFoundryDoesNotExist                  = commonerrors.RegisterGlobalError("foundry does not exist")
	ErrCantModifySupplyOfTheToken           = commonerrors.RegisterGlobalError("supply of the token is not controlled by the chain")
	ErrNativeTokenSupplyOutOffBounds        = commonerrors.RegisterGlobalError("token supply is out of bounds")
	ErrFatalTxBuilderNotBalanced            = commonerrors.RegisterGlobalError("fatal: tx builder is not balanced")
	ErrInconsistentL2LedgerWithL1TxBuilder  = commonerrors.RegisterGlobalError("fatal: L2 ledger is not consistent with the L1 tx builder")
	ErrCantDestroyFoundryBeingCreated       = commonerrors.RegisterGlobalError("can't destroy foundry which is being created")
)
