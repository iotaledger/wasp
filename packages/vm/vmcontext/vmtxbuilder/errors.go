package vmtxbuilder

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"golang.org/x/xerrors"
)

// error codes used for handled panics
var (
	ErrBlockGasLimitExceeded                = xerrors.Errorf("exceeded maximum gas allowed in a block. MaxGasPerBlock = %d", gas.MaxGasPerBlock)
	ErrInputLimitExceeded                   = xerrors.Errorf("exceeded maximum number of inputs in transaction. iotago.MaxInputsCount = %d", iotago.MaxInputsCount)
	ErrOutputLimitExceeded                  = xerrors.Errorf("exceeded maximum number of outputs in transaction. iotago.MaxOutputsCount = %d", iotago.MaxOutputsCount)
	ErrOutputLimitInSingleCallExceeded      = xerrors.Errorf("exceeded maximum number of outputs a contract call can produce. iotago.MaxOutputsCount = %d", iotago.MaxOutputsCount)
	ErrNumberOfNativeTokensLimitExceeded    = xerrors.Errorf("exceeded maximum number of different native tokens in transaction. iotago.MaxNativeTokensCount = %d", iotago.MaxNativeTokensCount)
	ErrNotEnoughFundsForInternalDustDeposit = xerrors.New("not enough funds for internal dust deposit")
	ErrOverflow                             = xerrors.New("overflow")
	ErrNotEnoughIotaBalance                 = xerrors.New("not enough iota balance")
	ErrNotEnoughNativeAssetBalance          = xerrors.New("not enough native assets balance")
	ErrCreateFoundryMaxSupplyMustBePositive = xerrors.New("max supply must be positive")
	ErrCreateFoundryMaxSupplyTooBig         = xerrors.New("max supply is too big")
	ErrFoundryDoesNotExist                  = xerrors.New("foundry does not exist")
	ErrCantModifySupplyOfTheToken           = xerrors.New("supply of the token is not controlled by the chain")
	ErrNativeTokenSupplyOutOffBounds        = xerrors.New("token supply is out of bounds")
	ErrFatalTxBuilderNotBalanced            = xerrors.New("fatal: tx builder is not balanced")
	ErrInconsistentL2LedgerWithL1TxBuilder  = xerrors.New("fatal: L2 ledger is not consistent with the L1 tx builder")
	ErrCantDestroyFoundryWithSupply         = xerrors.New("can't destroy foundry with non-zero circulating supply")
	ErrCantDestroyFoundryBeingCreated       = xerrors.New("can't destroy foundry which is being created")
)
