package isc

import (
	"errors"
	"fmt"
	"math"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/iotaledger/hive.go/core/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmnames"
)

type evmOffLedgerRequest struct {
	chainID ChainID
	tx      *types.Transaction
	sender  *EthereumAddressAgentID // not serialized
}

var _ OffLedgerRequest = &evmOffLedgerRequest{}

func NewEVMOffLedgerRequest(chainID ChainID, tx *types.Transaction) (OffLedgerRequest, error) {
	sender, err := evmutil.GetSender(tx)
	if err != nil {
		return nil, err
	}
	return &evmOffLedgerRequest{
		chainID: chainID,
		tx:      tx,
		sender:  NewEthereumAddressAgentID(sender),
	}, nil
}

func (r *evmOffLedgerRequest) readFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
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

func (r *evmOffLedgerRequest) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.
		WriteByte(requestKindTagOffLedgerEVM).
		Write(r.chainID)
	b := evmtypes.EncodeTransaction(r.tx)
	mu.WriteUint32(uint32(len(b)))
	mu.WriteBytes(b)
}

func (r *evmOffLedgerRequest) Allowance() *Assets {
	return NewEmptyAssets()
}

func (r *evmOffLedgerRequest) CallTarget() CallTarget {
	return CallTarget{
		Contract:   Hn(evmnames.Contract),
		EntryPoint: Hn(evmnames.FuncSendTransaction),
	}
}

func (r *evmOffLedgerRequest) Params() dict.Dict {
	return dict.Dict{evmnames.FieldTransaction: evmtypes.EncodeTransaction(r.tx)}
}

func (r *evmOffLedgerRequest) Assets() *Assets {
	return NewEmptyAssets()
}

func (r *evmOffLedgerRequest) GasBudget() (gas uint64, isEVM bool) {
	return r.tx.Gas(), true
}

func (r *evmOffLedgerRequest) ID() RequestID {
	return NewRequestID(iotago.TransactionID(hashing.HashData(r.Bytes())), 0)
}

func (r *evmOffLedgerRequest) NFT() *NFT {
	return nil
}

func (r *evmOffLedgerRequest) SenderAccount() AgentID {
	if r.sender == nil {
		panic("could not determine sender from ethereum tx")
	}
	return r.sender
}

func (r *evmOffLedgerRequest) TargetAddress() iotago.Address {
	return r.chainID.AsAddress()
}

func (r *evmOffLedgerRequest) Bytes() []byte {
	mu := marshalutil.New()
	r.WriteToMarshalUtil(mu)
	return mu.Bytes()
}

func (r *evmOffLedgerRequest) IsOffLedger() bool {
	return true
}

func (r *evmOffLedgerRequest) String() string {
	return fmt.Sprintf("evmOffLedgerRequest(%s)", r.ID())
}

func (r *evmOffLedgerRequest) ChainID() ChainID {
	return r.chainID
}

func (r *evmOffLedgerRequest) Nonce() uint64 {
	return r.tx.Nonce()
}

func (r *evmOffLedgerRequest) VerifySignature() error {
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

type evmOffLedgerEstimateGasRequest struct {
	chainID ChainID
	callMsg ethereum.CallMsg
}

var _ OffLedgerRequest = &evmOffLedgerEstimateGasRequest{}

func NewEVMOffLedgerEstimateGasRequest(chainID ChainID, callMsg ethereum.CallMsg) OffLedgerRequest {
	return &evmOffLedgerEstimateGasRequest{
		chainID: chainID,
		callMsg: callMsg,
	}
}

func (r *evmOffLedgerEstimateGasRequest) readFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
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

func (r *evmOffLedgerEstimateGasRequest) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.
		WriteByte(requestKindTagOffLedgerEVM).
		Write(r.chainID)
	b := evmtypes.EncodeCallMsg(r.callMsg)
	mu.WriteUint32(uint32(len(b)))
	mu.WriteBytes(b)
}

func (r *evmOffLedgerEstimateGasRequest) Allowance() *Assets {
	return NewEmptyAssets()
}

func (r *evmOffLedgerEstimateGasRequest) CallTarget() CallTarget {
	return CallTarget{
		Contract:   Hn(evmnames.Contract),
		EntryPoint: Hn(evmnames.FuncEstimateGas),
	}
}

func (r *evmOffLedgerEstimateGasRequest) Params() dict.Dict {
	return dict.Dict{evmnames.FieldCallMsg: evmtypes.EncodeCallMsg(r.callMsg)}
}

func (r *evmOffLedgerEstimateGasRequest) Assets() *Assets {
	return NewEmptyAssets()
}

func (r *evmOffLedgerEstimateGasRequest) GasBudget() (gas uint64, isEVM bool) {
	if r.callMsg.Gas > 0 {
		return r.callMsg.Gas, true
	}
	// see VMContext::calculateAffordableGasBudget() when EstimateGasMode == true
	return math.MaxUint64, false
}

func (r *evmOffLedgerEstimateGasRequest) ID() RequestID {
	return NewRequestID(iotago.TransactionID(hashing.HashData(r.Bytes())), 0)
}

func (r *evmOffLedgerEstimateGasRequest) NFT() *NFT {
	return nil
}

func (r *evmOffLedgerEstimateGasRequest) SenderAccount() AgentID {
	return NewEthereumAddressAgentID(r.callMsg.From)
}

func (r *evmOffLedgerEstimateGasRequest) TargetAddress() iotago.Address {
	return r.chainID.AsAddress()
}

func (r *evmOffLedgerEstimateGasRequest) Bytes() []byte {
	mu := marshalutil.New()
	r.WriteToMarshalUtil(mu)
	return mu.Bytes()
}

func (r *evmOffLedgerEstimateGasRequest) IsOffLedger() bool {
	return true
}

func (r *evmOffLedgerEstimateGasRequest) String() string {
	return fmt.Sprintf("evmOffLedgerEstimateGasRequest(%s)", r.ID())
}

func (r *evmOffLedgerEstimateGasRequest) ChainID() ChainID {
	return r.chainID
}

func (r *evmOffLedgerEstimateGasRequest) Nonce() uint64 {
	return 0
}

func (r *evmOffLedgerEstimateGasRequest) VerifySignature() error {
	return errors.New("evmOffLedgerEstimateGasRequest should never be used to send regular requests")
}
