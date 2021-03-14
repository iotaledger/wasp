package sctransaction

import (
	"bytes"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"golang.org/x/crypto/blake2b"
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
	minted        uint64
	requestData   RequestMetadata
	solidArgs     dict.Dict
}

// RequestDataFromOutput
func RequestFromOutput(output *ledgerstate.ExtendedLockedOutput, senderAddr ledgerstate.Address, minted ...uint64) *Request {
	ret := &Request{output: output, senderAddress: senderAddr}
	r, err := RequestPayloadFromBytes(output.GetPayload())
	if err != nil {
		return ret
	}
	ret.requestData = r
	if len(minted) > 0 {
		ret.minted = minted[0]
	}
	ret.parsedOk = true
	return ret
}

func (req *Request) Output() *ledgerstate.ExtendedLockedOutput {
	return req.output
}

func (req *Request) SenderAddress() ledgerstate.Address {
	return req.senderAddress
}

func (req *Request) SenderAgentID() (ret coretypes.AgentID) {
	if req.senderAddress.Type() == ledgerstate.AliasAddressType {
		chainID, err := coretypes.NewChainIDFromAddress(req.senderAddress)
		if err != nil {
			panic(err)
		}
		senderContractID := coretypes.NewContractID(chainID, req.requestData.SenderContractHname)
		ret = coretypes.NewAgentIDFromContractID(senderContractID)
	} else {
		var err error
		ret, err = coretypes.NewAgentIDFromAddress(req.senderAddress)
		if err != nil {
			panic(err)
		}
	}
	return
}

func (req *Request) ParsedOk() bool {
	return req.parsedOk
}

func (req *Request) SetMetadata(d *RequestMetadata) {
	req.requestData = *d
	req.requestData.Args = d.Args.Clone()
}

func (req *Request) GetMetadata() *RequestMetadata {
	ret := req.requestData
	ret.Args = req.requestData.Args.Clone()
	return &ret
}

func (req *Request) MintColor() ledgerstate.Color {
	return blake2b.Sum256(req.Output().ID().Bytes())
}

func (req *Request) MintedAmount() uint64 {
	return req.minted
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

func OutputsFromRequests(requests ...*Request) []ledgerstate.Output {
	ret := make([]ledgerstate.Output, len(requests))
	for i, req := range requests {
		ret[i] = req.Output()
	}
	return ret
}

// endregion /////////////////////////////////////////////////////////////////

// region RequestMetadata  ///////////////////////////////////////////////////////

// RequestMetadata represents content of the data payload of the output
type RequestMetadata struct {
	SenderContractHname coretypes.Hname
	// ID of the target smart contract
	TargetContractHname coretypes.Hname
	// entry point code
	EntryPoint coretypes.Hname
	// request arguments, not decoded yet wrt blobRefs
	Args requestargs.RequestArgs
}

func RequestPayloadFromBytes(data []byte) (ret RequestMetadata, err error) {
	err = ret.Read(bytes.NewReader(data))
	return
}

func (p *RequestMetadata) Bytes() []byte {
	var buf bytes.Buffer
	_ = p.Write(&buf)
	return buf.Bytes()
}

func (p *RequestMetadata) Write(w io.Writer) error {
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

func (p *RequestMetadata) Read(r io.Reader) error {
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

// endregion
