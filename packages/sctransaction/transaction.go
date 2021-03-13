package sctransaction

import (
	"bytes"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"time"
)

// region ParsedTransaction //////////////////////////////////////////////////////////////////

// ParsedTransaction is a wrapper of ledgerstate.Transaction. It provides additional validation
// and methods for ISCP. Represents a set of pre-parsed outputs with target of specific chainID
type ParsedTransaction struct{
	*ledgerstate.Transaction
	receivingChainID coretypes.ChainID
	senderAddr       ledgerstate.Address
	chainOutput      *ledgerstate.ChainOutput
	requests         []RequestData
}


// ParseNew
func ParseNew(tx *ledgerstate.Transaction, sender ledgerstate.Address, receivingChainID coretypes.ChainID) *ParsedTransaction{
	ret := &ParsedTransaction{
		Transaction:      tx,
		receivingChainID: receivingChainID,
		senderAddr:       sender,
		requests:         make([]RequestData, 0),
	}
	for _, out := range tx.Essence().Outputs(){
		if !out.Address().Equals(receivingChainID.AsAddress()){
			continue
		}
		switch o := out.(type) {
		case *ledgerstate.ExtendedLockedOutput:
			ret.requests = append(ret.requests, RequestDataFromOutput(o))
		case *ledgerstate.ChainOutput:
			ret.chainOutput = o
		default:
			continue
		}
	}
	return ret
}

func (tx *ParsedTransaction) SenderAddress() ledgerstate.Address{
	return tx.senderAddr
}

func (tx *ParsedTransaction) Requests() []RequestData{
	return tx.requests
}


// endregion /////////////////////////////////////////////////////////////////

// region RequestData //////////////////////////////////////////////////////////////////

type RequestData struct {
	output *ledgerstate.ExtendedLockedOutput
	parsedOk            bool
	// senderAddress contract index
	// - if state block present, it is hname of the sending contract in the chain of which state transaction it is
	// - if state block is absent, it is uninterpreted (it means requests are sent by the wallet)
	senderContractHname coretypes.Hname
	// ID of the target smart contract
	targetContractHname coretypes.Hname
	// entry point code
	entryPoint          coretypes.Hname
	// request arguments, not decoded yet wrt blobRefs
	args requestargs.RequestArgs
	// decoded args, if not nil. If nil, it means it wasn't
	// successfully decoded yet and can't be used in the batch for calculations in VM
	solidArgs dict.Dict
}

// RequestDataFromOutput
func RequestDataFromOutput(output *ledgerstate.ExtendedLockedOutput) RequestData{
	ret := RequestData{output: output}
	if len(output.GetPayload()) < 3 * coretypes.HnameLength{
		return ret
	}
	buf := bytes.NewReader(output.GetPayload())
	_ = ret.senderContractHname.Read(buf)
	_ = ret.targetContractHname.Read(buf)
	_ = ret.entryPoint.Read(buf)
	ret.args = requestargs.New(nil)
	if err := ret.args.Read(buf); err != nil {
		return ret
	}
	ret.parsedOk = true
	return ret
}

func (req *RequestData)Output() *ledgerstate.ExtendedLockedOutput{
	return req.output
}

func (req *RequestData)ParsedOk() bool{
	return req.parsedOk
}

func (req *RequestData)SenderContractHname() coretypes.Hname{
	return req.senderContractHname
}

func (req *RequestData)TargetContractHname() coretypes.Hname{
	return req.targetContractHname
}

func (req *RequestData)EntryPoint() coretypes.Hname{
	return req.entryPoint
}

// SolidArgs returns solid args if decoded already or nil otherwise
func (req *RequestData) SolidArgs() dict.Dict {
	return req.solidArgs
}

// SolidifyArgs return true if solidified successfully
func (req *RequestData) SolidifyArgs(reg coretypes.BlobCache) (bool, error) {
	if req.solidArgs != nil {
		return true, nil
	}
	solid, ok, err := req.args.SolidifyRequestArguments(reg)
	if err != nil || !ok {
		return ok, err
	}
	req.solidArgs = solid
	if req.solidArgs == nil {
		panic("req.solidArgs == nil")
	}
	return true, nil
}

func (req *RequestData)UnlockedNow(nowis time.Time) bool{
	req.output.
}

// endregion /////////////////////////////////////////////////////////////////
