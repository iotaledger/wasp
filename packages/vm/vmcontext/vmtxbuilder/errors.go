package vmtxbuilder

import (
	"github.com/iotaledger/wasp/packages/vm/core/errors/commonerrors"
)

// error codes used for handled panics
var (
	ErrOverflow                             = commonerrors.RegisterGlobalError("overflow").Create()
	ErrNotEnoughIotaBalance                 = commonerrors.RegisterGlobalError("not enough iota balance").Create()
	ErrNotEnoughNativeAssetBalance          = commonerrors.RegisterGlobalError("not enough native assets balance").Create()
	ErrCreateFoundryMaxSupplyMustBePositive = commonerrors.RegisterGlobalError("max supply must be positive").Create()
	ErrCreateFoundryMaxSupplyTooBig         = commonerrors.RegisterGlobalError("max supply is too big").Create()
	ErrFoundryDoesNotExist                  = commonerrors.RegisterGlobalError("foundry does not exist").Create()
	ErrCantModifySupplyOfTheToken           = commonerrors.RegisterGlobalError("supply of the token is not controlled by the chain").Create()
	ErrNativeTokenSupplyOutOffBounds        = commonerrors.RegisterGlobalError("token supply is out of bounds").Create()
	ErrFatalTxBuilderNotBalanced            = commonerrors.RegisterGlobalError("fatal: tx builder is not balanced").Create()
	ErrInconsistentL2LedgerWithL1TxBuilder  = commonerrors.RegisterGlobalError("fatal: L2 ledger is not consistent with the L1 tx builder").Create()
	ErrCantDestroyFoundryBeingCreated       = commonerrors.RegisterGlobalError("can't destroy foundry which is being created").Create()
)
