package vm

import (
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

var (
	ErrOverflow                             = coreerrors.RegisterGlobalError("overflow").Create()
	ErrNotEnoughIotaBalance                 = coreerrors.RegisterGlobalError("not enough iota balance").Create()
	ErrNotEnoughNativeAssetBalance          = coreerrors.RegisterGlobalError("not enough native assets balance").Create()
	ErrCreateFoundryMaxSupplyMustBePositive = coreerrors.RegisterGlobalError("max supply must be positive").Create()
	ErrCreateFoundryMaxSupplyTooBig         = coreerrors.RegisterGlobalError("max supply is too big").Create()
	ErrFoundryDoesNotExist                  = coreerrors.RegisterGlobalError("foundry does not exist").Create()
	ErrCantModifySupplyOfTheToken           = coreerrors.RegisterGlobalError("supply of the token is not controlled by the chain").Create()
	ErrNativeTokenSupplyOutOffBounds        = coreerrors.RegisterGlobalError("token supply is out of bounds").Create()
	ErrFatalTxBuilderNotBalanced            = coreerrors.RegisterGlobalError("fatal: tx builder is not balanced").Create()
	ErrInconsistentL2LedgerWithL1TxBuilder  = coreerrors.RegisterGlobalError("fatal: L2 ledger is not consistent with the L1 tx builder").Create()
	ErrCantDestroyFoundryBeingCreated       = coreerrors.RegisterGlobalError("can't destroy foundry which is being created").Create()

	ErrContractNotFound                   = coreerrors.RegisterGlobalError("contract not found id:%v")
	ErrTargetEntryPointNotFound           = coreerrors.RegisterGlobalError("entry point not found").Create()
	ErrEntryPointCantBeAView              = coreerrors.RegisterGlobalError("'init' entry point can't be a view").Create()
	ErrTargetContractNotFound             = coreerrors.RegisterGlobalError("target contract not found").Create()
	ErrTransferTargetAccountDoesNotExists = coreerrors.RegisterGlobalError("transfer target account does not exist").Create()
	ErrRepeatingInitCall                  = coreerrors.RegisterGlobalError("repeating init call").Create()
	ErrInconsistentDustAssumptions        = coreerrors.RegisterGlobalError("dust deposit requirements are not consistent with the chain assumptions").Create()
	ErrTooManyEvents                      = coreerrors.RegisterGlobalError("too many events issued for contract").Create()
	ErrTooLargeEvent                      = coreerrors.RegisterGlobalError("event data is too large").Create()
	ErrPrivilegedCallFailed               = coreerrors.RegisterGlobalError("privileged call failed").Create()
	ErrExceededPostedOutputLimit          = coreerrors.RegisterGlobalError("exceeded maximum number of %d posted outputs in one request").Create(42)
	ErrGasBudgetExceeded                  = coreerrors.RegisterGlobalError("gas budget exceeded").Create()
	ErrSenderUnknown                      = coreerrors.RegisterGlobalError("sender unknown").Create()
	ErrGasBudgetDetail                    = coreerrors.RegisterGlobalError("%v: burned (budget) = %d (%d)")
)
