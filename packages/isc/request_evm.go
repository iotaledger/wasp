package isc

import (
	"errors"
	"fmt"
	"math"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
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
		sender:  NewEthereumAddressAgentID(sender),
	}, nil
}

func (r *evmOffLedgerTxRequest) readFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	var err error
	if r.chainID, err = ChainIDFromMarshalUtil(mu); err != nil {
		return err
	}
	var txLen uint32
	if txLen, err = mu.ReadUint32(); err != nil {
		return err
	}
	var txBytes []byte
	if txBytes, err = mu.ReadBytes(int(txLen)); err != nil {
		return err
	}
	if r.tx, err = evmtypes.DecodeTransaction(txBytes); err != nil {
		return err
	}
	sender, err := evmutil.GetSender(r.tx)
	if err != nil {
		return err
	}
	r.sender = NewEthereumAddressAgentID(sender)
	return nil
}

func (r *evmOffLedgerTxRequest) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.
		WriteByte(requestKindTagOffLedgerEVMTx).
		Write(r.chainID)
	b := evmtypes.EncodeTransaction(r.tx)
	mu.WriteUint32(uint32(len(b)))
	mu.WriteBytes(b)
}

func (r *evmOffLedgerTxRequest) Allowance() *Assets {
	return NewEmptyAssets()
}

func (r *evmOffLedgerTxRequest) CallTarget() CallTarget {
	return CallTarget{
		Contract:   Hn(evmnames.Contract),
		EntryPoint: Hn(evmnames.FuncSendTransaction),
	}
}

func (r *evmOffLedgerTxRequest) Params() dict.Dict {
	return dict.Dict{evmnames.FieldTransaction: evmtypes.EncodeTransaction(r.tx)}
}

func (r *evmOffLedgerTxRequest) Assets() *Assets {
	return NewEmptyAssets()
}

func (r *evmOffLedgerTxRequest) GasBudget() (gas uint64, isEVM bool) {
	return r.tx.Gas(), true
}

func (r *evmOffLedgerTxRequest) ID() RequestID {
	return NewRequestID(iotago.TransactionID(hashing.HashData(r.Bytes())), 0)
}

func (r *evmOffLedgerTxRequest) NFT() *NFT {
	return nil
}

func (r *evmOffLedgerTxRequest) SenderAccount() AgentID {
	if r.sender == nil {
		panic("could not determine sender from ethereum tx")
	}
	return r.sender
}

func (r *evmOffLedgerTxRequest) TargetAddress() iotago.Address {
	return r.chainID.AsAddress()
}

func (r *evmOffLedgerTxRequest) Bytes() []byte {
	mu := marshalutil.New()
	r.WriteToMarshalUtil(mu)
	return mu.Bytes()
}

func (r *evmOffLedgerTxRequest) IsOffLedger() bool {
	return true
}

func (r *evmOffLedgerTxRequest) String() string {
	return fmt.Sprintf("%T(%s)", r, r.ID())
}

func (r *evmOffLedgerTxRequest) ChainID() ChainID {
	return r.chainID
}

func (r *evmOffLedgerTxRequest) Nonce() uint64 {
	return r.tx.Nonce()
}

func (r *evmOffLedgerTxRequest) VerifySignature() error {
	sender, err := evmutil.GetSender(r.tx)
	if err != nil {
		return fmt.Errorf("cannot verify Ethereum tx sender: %w", err)
	}
	if sender != r.sender.EthAddress() {
		return errors.New("sender mismatch in EVM off-ledger request")
	}
	return nil
}

// ----------------------------------------------------------------

// evmOffLedgerCallRequest is used to wrap an EVM call (for the eth_call or eth_estimateGas jsonrpc methods)
type evmOffLedgerCallRequest struct {
	chainID ChainID
	callMsg ethereum.CallMsg
}

var _ OffLedgerRequest = &evmOffLedgerCallRequest{}

func NewEVMOffLedgerCallRequest(chainID ChainID, callMsg ethereum.CallMsg) OffLedgerRequest {
	return &evmOffLedgerCallRequest{
		chainID: chainID,
		callMsg: callMsg,
	}
}

func (r *evmOffLedgerCallRequest) readFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	var err error
	if r.chainID, err = ChainIDFromMarshalUtil(mu); err != nil {
		return err
	}
	var callMsgLen uint32
	if callMsgLen, err = mu.ReadUint32(); err != nil {
		return err
	}
	var callMsgBytes []byte
	if callMsgBytes, err = mu.ReadBytes(int(callMsgLen)); err != nil {
		return err
	}
	if r.callMsg, err = evmtypes.DecodeCallMsg(callMsgBytes); err != nil {
		return err
	}
	return nil
}

func (r *evmOffLedgerCallRequest) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.
		WriteByte(requestKindTagOffLedgerEVMTx).
		Write(r.chainID)
	b := evmtypes.EncodeCallMsg(r.callMsg)
	mu.WriteUint32(uint32(len(b)))
	mu.WriteBytes(b)
}

func (r *evmOffLedgerCallRequest) Allowance() *Assets {
	return NewEmptyAssets()
}

func (r *evmOffLedgerCallRequest) CallTarget() CallTarget {
	return CallTarget{
		Contract:   Hn(evmnames.Contract),
		EntryPoint: Hn(evmnames.FuncCallContract),
	}
}

func (r *evmOffLedgerCallRequest) Params() dict.Dict {
	return dict.Dict{evmnames.FieldCallMsg: evmtypes.EncodeCallMsg(r.callMsg)}
}

func (r *evmOffLedgerCallRequest) Assets() *Assets {
	return NewEmptyAssets()
}

func (r *evmOffLedgerCallRequest) GasBudget() (gas uint64, isEVM bool) {
	if r.callMsg.Gas > 0 {
		return r.callMsg.Gas, true
	}
	// see VMContext::calculateAffordableGasBudget() when EstimateGasMode == true
	return math.MaxUint64, false
}

func (r *evmOffLedgerCallRequest) ID() RequestID {
	return NewRequestID(iotago.TransactionID(hashing.HashData(r.Bytes())), 0)
}

func (r *evmOffLedgerCallRequest) NFT() *NFT {
	return nil
}

func (r *evmOffLedgerCallRequest) SenderAccount() AgentID {
	return NewEthereumAddressAgentID(r.callMsg.From)
}

func (r *evmOffLedgerCallRequest) TargetAddress() iotago.Address {
	return r.chainID.AsAddress()
}

func (r *evmOffLedgerCallRequest) Bytes() []byte {
	mu := marshalutil.New()
	r.WriteToMarshalUtil(mu)
	return mu.Bytes()
}

func (r *evmOffLedgerCallRequest) IsOffLedger() bool {
	return true
}

func (r *evmOffLedgerCallRequest) String() string {
	return fmt.Sprintf("%T(%s)", r, r.ID())
}

func (r *evmOffLedgerCallRequest) ChainID() ChainID {
	return r.chainID
}

func (r *evmOffLedgerCallRequest) Nonce() uint64 {
	return 0
}

func (r *evmOffLedgerCallRequest) VerifySignature() error {
	return fmt.Errorf("%T should never be used to send regular requests", r)
}
