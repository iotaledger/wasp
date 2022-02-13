package vmtxbuilder

import (
	"github.com/iotaledger/wasp/packages/vm/core/errors/commonerrors"
)

// error codes used for handled panics
var (
	ErrOverflow                             = commonerrors.RegisterGlobalError("overflow").CreateTyped()
	ErrNotEnoughIotaBalance                 = commonerrors.RegisterGlobalError("not enough iota balance").CreateTyped()
	ErrNotEnoughNativeAssetBalance          = commonerrors.RegisterGlobalError("not enough native assets balance").CreateTyped()
	ErrCreateFoundryMaxSupplyMustBePositive = commonerrors.RegisterGlobalError("max supply must be positive").CreateTyped()
	ErrCreateFoundryMaxSupplyTooBig         = commonerrors.RegisterGlobalError("max supply is too big").CreateTyped()
	ErrFoundryDoesNotExist                  = commonerrors.RegisterGlobalError("foundry does not exist").CreateTyped()
	ErrCantModifySupplyOfTheToken           = commonerrors.RegisterGlobalError("supply of the token is not controlled by the chain").CreateTyped()
	ErrNativeTokenSupplyOutOffBounds        = commonerrors.RegisterGlobalError("token supply is out of bounds").CreateTyped()
	ErrFatalTxBuilderNotBalanced            = commonerrors.RegisterGlobalError("fatal: tx builder is not balanced").CreateTyped()
	ErrInconsistentL2LedgerWithL1TxBuilder  = commonerrors.RegisterGlobalError("fatal: L2 ledger is not consistent with the L1 tx builder").CreateTyped()
	ErrCantDestroyFoundryBeingCreated       = commonerrors.RegisterGlobalError("can't destroy foundry which is being created").CreateTyped()
)
