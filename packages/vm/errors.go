package vm

import (
	"github.com/iotaledger/wasp/packages/vm/core/errors/commonerrors"
)

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

	ErrContractNotFound                   = commonerrors.RegisterGlobalError("contract not found id:%v")
	ErrTargetEntryPointNotFound           = commonerrors.RegisterGlobalError("entry point not found").Create()
	ErrEntryPointCantBeAView              = commonerrors.RegisterGlobalError("'init' entry point can't be a view").Create()
	ErrTargetContractNotFound             = commonerrors.RegisterGlobalError("target contract not found").Create()
	ErrTransferTargetAccountDoesNotExists = commonerrors.RegisterGlobalError("transfer target account does not exist").Create()
	ErrRepeatingInitCall                  = commonerrors.RegisterGlobalError("repeating init call").Create()
	ErrInconsistentDustAssumptions        = commonerrors.RegisterGlobalError("dust deposit requirements are not consistent with the chain assumptions").Create()
	ErrTooManyEvents                      = commonerrors.RegisterGlobalError("too many events issued for contract").Create()
	ErrTooLargeEvent                      = commonerrors.RegisterGlobalError("event data is too large").Create()
	ErrPrivilegedCallFailed               = commonerrors.RegisterGlobalError("privileged call failed").Create()
	ErrExceededPostedOutputLimit          = commonerrors.RegisterGlobalError("exceeded maximum number of %d posted outputs in one request").Create(42)
	ErrGasBudgetExceeded                  = commonerrors.RegisterGlobalError("gas budget exceeded").Create()
	ErrSenderUnknown                      = commonerrors.RegisterGlobalError("sender unknown").Create()
	ErrGasBudgetDetail                    = commonerrors.RegisterGlobalError("%v: burned (budget) = %d (%d)")
)
