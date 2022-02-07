package vmtxbuilder

import (
	"golang.org/x/xerrors"
)

// error codes used for handled panics
var (
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
	ErrCantDestroyFoundryBeingCreated       = xerrors.New("can't destroy foundry which is being created")
)
