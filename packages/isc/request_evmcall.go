package isc

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmnames"
)

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

func (req *evmOffLedgerCallRequest) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadKindAndVerify(rwutil.Kind(requestKindOffLedgerEVMTx))
	rr.Read(&req.chainID)
	data := rr.ReadBytes()
	if rr.Err == nil {
		req.callMsg, rr.Err = evmtypes.DecodeCallMsg(data)
	}
	return rr.Err
}

func (req *evmOffLedgerCallRequest) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteKind(rwutil.Kind(requestKindOffLedgerEVMTx))
	ww.Write(&req.chainID)
	if ww.Err == nil {
		data := evmtypes.EncodeCallMsg(req.callMsg)
		ww.WriteBytes(data)
	}
	return ww.Err
}

func (req *evmOffLedgerCallRequest) Allowance() *Assets {
	return NewEmptyAssets()
}

func (req *evmOffLedgerCallRequest) Assets() *Assets {
	return NewEmptyAssets()
}

func (req *evmOffLedgerCallRequest) Bytes() []byte {
	return rwutil.WriteToBytes(req)
}

func (req *evmOffLedgerCallRequest) CallTarget() CallTarget {
	return CallTarget{
		Contract:   Hn(evmnames.Contract),
		EntryPoint: Hn(evmnames.FuncCallContract),
	}
}

func (req *evmOffLedgerCallRequest) ChainID() ChainID {
	return req.chainID
}

func (req *evmOffLedgerCallRequest) GasBudget() (gas uint64, isEVM bool) {
	return req.callMsg.Gas, true
}

func (req *evmOffLedgerCallRequest) ID() RequestID {
	return NewRequestID(iotago.TransactionID(hashing.HashData(req.Bytes())), 0)
}

func (req *evmOffLedgerCallRequest) IsOffLedger() bool {
	return true
}

func (req *evmOffLedgerCallRequest) NFT() *NFT {
	return nil
}

func (req *evmOffLedgerCallRequest) Nonce() uint64 {
	return 0
}

func (req *evmOffLedgerCallRequest) Params() dict.Dict {
	return dict.Dict{evmnames.FieldCallMsg: evmtypes.EncodeCallMsg(req.callMsg)}
}

func (req *evmOffLedgerCallRequest) SenderAccount() AgentID {
	return NewEthereumAddressAgentID(req.chainID, req.callMsg.From)
}

func (req *evmOffLedgerCallRequest) String() string {
	// ignore error so String does not crash the app
	data, _ := json.MarshalIndent(req.callMsg, " ", " ")
	return fmt.Sprintf("%T::{ ID: %s, callMsg: %s }",
		req,
		req.ID(),
		data,
	)
}

func (req *evmOffLedgerCallRequest) TargetAddress() iotago.Address {
	return req.chainID.AsAddress()
}

func (req *evmOffLedgerCallRequest) VerifySignature() error {
	return fmt.Errorf("%T should never be used to send regular requests", req)
}

func (*evmOffLedgerCallRequest) EVMTransaction() *types.Transaction {
	return nil
}
