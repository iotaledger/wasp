package vm

import (
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

var (
	ErrOverflow                             = coreerrors.Register("overflow").Create()
	ErrNotEnoughBaseTokensBalance           = coreerrors.Register("not enough base tokens balance").Create()
	ErrNotEnoughNativeAssetBalance          = coreerrors.Register("not enough native assets balance").Create()
	ErrNotEnoughFundsForAllowance           = coreerrors.Register("not enough funds for allowance").Create()
	ErrCreateFoundryMaxSupplyMustBePositive = coreerrors.Register("max supply must be positive").Create()
	ErrCreateFoundryMaxSupplyTooBig         = coreerrors.Register("max supply is too big").Create()
	ErrFoundryDoesNotExist                  = coreerrors.Register("foundry does not exist").Create()
	ErrCantModifySupplyOfTheToken           = coreerrors.Register("supply of the token is not controlled by the chain").Create()
	ErrNativeTokenSupplyOutOffBounds        = coreerrors.Register("token supply is out of bounds").Create()
	ErrFatalTxBuilderNotBalanced            = coreerrors.Register("fatal: tx builder is not balanced").Create()
	ErrInconsistentL2LedgerWithL1TxBuilder  = coreerrors.Register("fatal: L2 ledger is not consistent with the L1 tx builder").Create()
	ErrCantDestroyFoundryBeingCreated       = coreerrors.Register("can't destroy foundry which is being created").Create()

	ErrContractNotFound                      = coreerrors.Register("contract not found id:%d")
	ErrTargetEntryPointNotFound              = coreerrors.Register("entry point not found").Create()
	ErrEntryPointCantBeAView                 = coreerrors.Register("'init' entry point can't be a view").Create()
	ErrRepeatingInitCall                     = coreerrors.Register("repeating init call").Create()
	ErrInconsistentStorageDepositAssumptions = coreerrors.Register("storage deposit requirements are not consistent with the chain assumptions").Create()
	ErrTooManyEvents                         = coreerrors.Register("too many events issued for contract").Create()
	ErrPrivilegedCallFailed                  = coreerrors.Register("privileged call failed").Create()
	ErrExceededPostedOutputLimit             = coreerrors.Register("exceeded maximum number of %d posted outputs in one request").Create(42)
	ErrGasBudgetExceeded                     = coreerrors.Register("gas budget exceeded").Create()
	ErrSenderUnknown                         = coreerrors.Register("sender unknown").Create()
	ErrNotEnoughTokensLeftForGas             = coreerrors.Register("not enough funds left to pay for gas")
	ErrUnauthorized                          = coreerrors.Register("unauthorized access").Create()
	ErrIllegalCall                           = coreerrors.Register("illegal call - entrypoint cannot be called from contracts")
	ErrSendMultipleNFTs                      = coreerrors.Register("cannot send more than 1 NFT").Create()
)
