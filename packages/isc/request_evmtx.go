package isc

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"

	bcs "github.com/iotaledger/bcs-go"
	_ "github.com/iotaledger/wasp/packages/evm/evmtypes" // register BCS custom encoder for Transaction
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmnames"
)

// evmOffLedgerTxRequest is used to wrap an EVM tx
type evmOffLedgerTxRequest struct {
	chainID ChainID                 `bcs:"export"`
	tx      *types.Transaction      `bcs:"export"`
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
		sender:  NewEthereumAddressAgentID(sender),
	}, nil
}

func (req *evmOffLedgerTxRequest) BCSInit() error {
	// derive req.sender from req.tx
	sender, err := evmutil.GetSender(req.tx)
	if err != nil {
		return err
	}
	req.sender = NewEthereumAddressAgentID(sender)
	return nil
}

func (req *evmOffLedgerTxRequest) Allowance() (*Assets, error) {
	return NewEmptyAssets(), nil
}

func (req *evmOffLedgerTxRequest) Assets() *Assets {
	return NewEmptyAssets()
}

func (req *evmOffLedgerTxRequest) Bytes() []byte {
	var r Request = req
	return bcs.MustMarshal(&r)
}

func (req *evmOffLedgerTxRequest) Message() Message {
	return NewMessage(
		Hn(evmnames.Contract),
		Hn(evmnames.FuncSendTransaction),
		NewCallArguments(bcs.MustMarshal(req.tx)),
	)
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

func (req *evmOffLedgerTxRequest) Nonce() uint64 {
	return req.tx.Nonce()
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

func (req *evmOffLedgerTxRequest) EVMCallMsg() *ethereum.CallMsg {
	return EVMCallDataFromTx(req.tx)
}

func (req *evmOffLedgerTxRequest) TxValue() *big.Int {
	return req.tx.Value()
}

func (req *evmOffLedgerTxRequest) GasPrice() *big.Int {
	return req.tx.GasPrice()
}
