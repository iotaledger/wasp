package sctransaction

import (
	"bytes"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"io"
)

// region ParsedTransaction //////////////////////////////////////////////////////////////////

// ParsedTransaction is a wrapper of ledgerstate.Transaction. It provides additional validation
// and methods for ISCP. Represents a set of parsed outputs with target of specific chainID
type ParsedTransaction struct {
	*ledgerstate.Transaction
	receivingChainID coretypes.ChainID
	senderAddr       ledgerstate.Address
	chainOutput      *ledgerstate.ChainOutput
	stateHash        hashing.HashValue
	requests         []*Request
}

// Parse analyzes value transaction and parses its data
func Parse(tx *ledgerstate.Transaction, sender ledgerstate.Address, receivingChainID coretypes.ChainID) *ParsedTransaction {
	ret := &ParsedTransaction{
		Transaction:      tx,
		receivingChainID: receivingChainID,
		senderAddr:       sender,
		requests:         make([]*Request, 0),
	}
	for _, out := range tx.Essence().Outputs() {
		if !out.Address().Equals(receivingChainID.AsAddress()) {
			continue
		}
		switch o := out.(type) {
		case *ledgerstate.ExtendedLockedOutput:
			ret.requests = append(ret.requests, RequestFromOutput(o, sender))
		case *ledgerstate.ChainOutput:
			h, err := hashing.HashValueFromBytes(o.GetStateData())
			if err == nil {
				ret.stateHash = h
			}
			ret.chainOutput = o
		default:
			continue
		}
	}
	return ret
}

// ChainOutput return chain output or nil if the transaction is not a state anchor
func (tx *ParsedTransaction) ChainOutput() *ledgerstate.ChainOutput {
	return tx.chainOutput
}

func (tx *ParsedTransaction) SenderAddress() ledgerstate.Address {
	return tx.senderAddr
}

func (tx *ParsedTransaction) Requests() []*Request {
	return tx.requests
}

// endregion /////////////////////////////////////////////////////////////////

// region Request //////////////////////////////////////////////////////////////////

type Request struct {
	output        *ledgerstate.ExtendedLockedOutput
	senderAddress ledgerstate.Address
	parsedOk      bool
	requestData   RequestData
	solidArgs     dict.Dict
}

// RequestDataFromOutput
func RequestFromOutput(output *ledgerstate.ExtendedLockedOutput, senderAddr ledgerstate.Address) *Request {
	ret := &Request{output: output, senderAddress: senderAddr}
	r, err := RequestPayloadFromBytes(output.GetPayload())
	if err != nil {
		return ret
	}
	ret.requestData = *r
	ret.parsedOk = true
	return ret
}

func (req *Request) Output() *ledgerstate.ExtendedLockedOutput {
	return req.output
}

func (req *Request) ParsedOk() bool {
	return req.parsedOk
}

func (req *Request) SetData(d *RequestData) {
	req.requestData = *d
	req.requestData.Args = d.Args.Clone()
}

func (req *Request) GetData() *RequestData {
	ret := req.requestData
	ret.Args = req.requestData.Args.Clone()
	return &ret
}

// SolidArgs returns solid args if decoded already or nil otherwise
func (req *Request) SolidArgs() dict.Dict {
	return req.solidArgs
}

// SolidifyArgs return true if solidified successfully
func (req *Request) SolidifyArgs(reg coretypes.BlobCache) (bool, error) {
	if req.solidArgs != nil {
		return true, nil
	}
	solid, ok, err := req.requestData.Args.SolidifyRequestArguments(reg)
	if err != nil || !ok {
		return ok, err
	}
	req.solidArgs = solid
	if req.solidArgs == nil {
		panic("req.solidArgs == nil")
	}
	return true, nil
}

// endregion /////////////////////////////////////////////////////////////////

type RequestData struct {
	SenderContractHname coretypes.Hname
	// ID of the target smart contract
	TargetContractHname coretypes.Hname
	// entry point code
	EntryPoint coretypes.Hname
	// request arguments, not decoded yet wrt blobRefs
	Args requestargs.RequestArgs
}

func RequestPayloadFromBytes(data []byte) (*RequestData, error) {
	ret := &RequestData{}
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}

func (p *RequestData) Bytes() []byte {
	var buf bytes.Buffer
	_ = p.Write(&buf)
	return buf.Bytes()
}

func (p *RequestData) Write(w io.Writer) error {
	if err := p.SenderContractHname.Write(w); err != nil {
		return err
	}
	if err := p.TargetContractHname.Write(w); err != nil {
		return err
	}
	if err := p.EntryPoint.Write(w); err != nil {
		return err
	}
	if err := p.Args.Write(w); err != nil {
		return err
	}
	return nil
}

func (p *RequestData) Read(r io.Reader) error {
	if err := p.SenderContractHname.Read(r); err != nil {
		return err
	}
	if err := p.TargetContractHname.Read(r); err != nil {
		return err
	}
	if err := p.EntryPoint.Read(r); err != nil {
		return err
	}
	if err := p.Args.Read(r); err != nil {
		return err
	}
	return nil
}
