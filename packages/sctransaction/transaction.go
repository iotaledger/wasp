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
	output          *ledgerstate.ExtendedLockedOutput
	senderAddress   ledgerstate.Address
	minted          uint64
	requestMetadata RequestMetadata
	solidArgs       dict.Dict
}

// RequestDataFromOutput
func RequestFromOutput(output *ledgerstate.ExtendedLockedOutput, senderAddr ledgerstate.Address, minted ...uint64) *Request {
	ret := &Request{output: output, senderAddress: senderAddr}
	ret.requestMetadata = *RequestMetadataFromBytes(output.GetPayload())
	if len(minted) > 0 {
		ret.minted = minted[0]
	}
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
		senderContractID := coretypes.NewContractID(chainID, req.requestMetadata.SenderContract())
		ret = coretypes.NewAgentIDFromContractID(senderContractID)
	} else {
		ret = coretypes.NewAgentIDFromAddress(req.senderAddress)
	}
	return
}

func (req *Request) SetMetadata(d *RequestMetadata) {
	req.requestMetadata = *d.Clone()
}

func (req *Request) GetMetadata() *RequestMetadata {
	return &req.requestMetadata
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
	solid, ok, err := req.requestMetadata.Args().SolidifyRequestArguments(reg)
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
	parsedOk       bool
	senderContract coretypes.Hname
	// ID of the target smart contract
	targetContract coretypes.Hname
	// entry point code
	entryPoint coretypes.Hname
	// request arguments, not decoded yet wrt blobRefs
	args requestargs.RequestArgs
}

func NewRequestMetadata() *RequestMetadata {
	return &RequestMetadata{
		parsedOk: true,
		args:     requestargs.RequestArgs(dict.New()),
	}
}

func RequestMetadataFromBytes(data []byte) *RequestMetadata {
	ret := NewRequestMetadata()
	err := ret.Read(bytes.NewReader(data))
	ret.parsedOk = err == nil
	return ret
}

func (p *RequestMetadata) WithSender(s coretypes.Hname) *RequestMetadata {
	p.senderContract = s
	return p
}

func (p *RequestMetadata) WithTarget(t coretypes.Hname) *RequestMetadata {
	p.targetContract = t
	return p
}

func (p *RequestMetadata) WithEntryPoint(ep coretypes.Hname) *RequestMetadata {
	p.entryPoint = ep
	return p
}

func (p *RequestMetadata) WithArgs(args requestargs.RequestArgs) *RequestMetadata {
	p.args = args.Clone()
	return p
}

func (p *RequestMetadata) Clone() *RequestMetadata {
	ret := *p
	ret.args = p.args.Clone()
	return &ret
}

func (p *RequestMetadata) ParsedOk() bool {
	return p.parsedOk
}

func (p *RequestMetadata) SenderContract() coretypes.Hname {
	if !p.parsedOk {
		return 0
	}
	return p.senderContract
}

func (p *RequestMetadata) TargetContract() coretypes.Hname {
	if !p.parsedOk {
		return 0
	}
	return p.targetContract
}

func (p *RequestMetadata) EntryPoint() coretypes.Hname {
	if !p.parsedOk {
		return 0
	}
	return p.entryPoint
}

func (p *RequestMetadata) Args() requestargs.RequestArgs {
	if !p.parsedOk {
		return requestargs.RequestArgs(dict.New())
	}
	return p.args
}

func (p *RequestMetadata) Bytes() []byte {
	var buf bytes.Buffer
	_ = p.Write(&buf)
	return buf.Bytes()
}

func (p *RequestMetadata) Write(w io.Writer) error {
	if err := p.senderContract.Write(w); err != nil {
		return err
	}
	if err := p.targetContract.Write(w); err != nil {
		return err
	}
	if err := p.entryPoint.Write(w); err != nil {
		return err
	}
	if err := p.args.Write(w); err != nil {
		return err
	}
	return nil
}

func (p *RequestMetadata) Read(r io.Reader) error {
	if err := p.senderContract.Read(r); err != nil {
		return err
	}
	if err := p.targetContract.Read(r); err != nil {
		return err
	}
	if err := p.entryPoint.Read(r); err != nil {
		return err
	}
	if err := p.args.Read(r); err != nil {
		return err
	}
	return nil
}

// endregion
