package isc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/core/types"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmnames"
)

// evmOffLedgerTxRequest is used to wrap an EVM tx
type evmOffLedgerTxRequest struct {
	chainID ChainID
	tx      *types.Transaction
	sender  *EthereumAddressAgentID // not serialized
}

var _ OffLedgerRequest = &evmOffLedgerTxRequest{}

func NewEVMOffLedgerTxRequest(chainID ChainID, tx *types.Transaction) (OffLedgerRequest, error) {
	sender, err := evmutil.GetSender(tx)
	if err != nil {
		return nil, err
	}
	return &evmOffLedgerTxRequest{
		chainID: chainID,
		tx:      tx,
		sender:  NewEthereumAddressAgentID(chainID, sender),
	}, nil
}

func (req *evmOffLedgerTxRequest) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadKindAndVerify(rwutil.Kind(requestKindOffLedgerEVMTx))
	rr.Read(&req.chainID)
	txData := rr.ReadBytes()
	if rr.Err != nil {
		return rr.Err
	}
	req.tx, rr.Err = evmtypes.DecodeTransaction(txData)
	if rr.Err != nil {
		return rr.Err
	}
	// derive req.sender from req.tx
	sender, err := evmutil.GetSender(req.tx)
	if err != nil {
		return err
	}
	req.sender = NewEthereumAddressAgentID(req.chainID, sender)
	return rr.Err
}

func (req *evmOffLedgerTxRequest) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteKind(rwutil.Kind(requestKindOffLedgerEVMTx))
	ww.Write(&req.chainID)
	if ww.Err == nil {
		txData := evmtypes.EncodeTransaction(req.tx)
		ww.WriteBytes(txData)
	}
	// no need to write req.sender, it can be derived from req.tx
	return ww.Err
}

func (req *evmOffLedgerTxRequest) Allowance() *Assets {
	return NewEmptyAssets()
}

func (req *evmOffLedgerTxRequest) Assets() *Assets {
	return NewEmptyAssets()
}

func (req *evmOffLedgerTxRequest) Bytes() []byte {
	return rwutil.WriteToBytes(req)
}

func (req *evmOffLedgerTxRequest) CallTarget() CallTarget {
	return CallTarget{
		Contract:   Hn(evmnames.Contract),
		EntryPoint: Hn(evmnames.FuncSendTransaction),
	}
}

func (req *evmOffLedgerTxRequest) ChainID() ChainID {
	return req.chainID
}

func (req *evmOffLedgerTxRequest) GasBudget() (gas uint64, isEVM bool) {
	return req.tx.Gas(), true
}

func (req *evmOffLedgerTxRequest) ID() RequestID {
	return RequestIDFromEVMTxHash(req.tx.Hash())
}

func (req *evmOffLedgerTxRequest) IsOffLedger() bool {
	return true
}

func (req *evmOffLedgerTxRequest) NFT() *NFT {
	return nil
}

func (req *evmOffLedgerTxRequest) Nonce() uint64 {
	return req.tx.Nonce()
}

func (req *evmOffLedgerTxRequest) Params() dict.Dict {
	return dict.Dict{evmnames.FieldTransaction: evmtypes.EncodeTransaction(req.tx)}
}

func (req *evmOffLedgerTxRequest) SenderAccount() AgentID {
	if req.sender == nil {
		panic("could not determine sender from ethereum tx")
	}
	return req.sender
}

func (req *evmOffLedgerTxRequest) String() string {
	// ignore error so String does not crash the app
	data, _ := json.MarshalIndent(req.tx, " ", " ")
	return fmt.Sprintf("%T::{ ID: %s, Tx: %s }",
		req,
		req.ID(),
		data,
	)
}

func (req *evmOffLedgerTxRequest) TargetAddress() iotago.Address {
	return req.chainID.AsAddress()
}

func (req *evmOffLedgerTxRequest) VerifySignature() error {
	sender, err := evmutil.GetSender(req.tx)
	if err != nil {
		return fmt.Errorf("cannot verify Ethereum tx sender: %w", err)
	}
	if sender != req.sender.EthAddress() {
		return errors.New("sender mismatch in EVM off-ledger request")
	}
	return nil
}

func (req *evmOffLedgerTxRequest) EVMTransaction() *types.Transaction {
	return req.tx
}
