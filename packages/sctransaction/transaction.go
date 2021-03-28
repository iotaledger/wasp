package sctransaction

import (
	"bytes"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"golang.org/x/crypto/blake2b"
	"io"
	"time"
)

// region ParsedTransaction //////////////////////////////////////////////////////////////////

// ParsedTransaction is a wrapper of ledgerstate.Transaction. It provides additional validation
// and methods for ISCP. Represents a set of parsed outputs with target of specific chainID
type ParsedTransaction struct {
	*ledgerstate.Transaction
	receivingChainID coretypes.ChainID
	senderAddr       ledgerstate.Address
	chainOutput      *ledgerstate.AliasOutput
	stateHash        hashing.HashValue
	requests         []*RequestOnLedger
}

// Parse analyzes value transaction and parses its data
func Parse(tx *ledgerstate.Transaction, sender ledgerstate.Address, receivingChainID coretypes.ChainID) *ParsedTransaction {
	ret := &ParsedTransaction{
		Transaction:      tx,
		receivingChainID: receivingChainID,
		senderAddr:       sender,
		requests:         make([]*RequestOnLedger, 0),
	}
	for _, out := range tx.Essence().Outputs() {
		if !out.Address().Equals(receivingChainID.AsAddress()) {
			continue
		}
		switch o := out.(type) {
		case *ledgerstate.ExtendedLockedOutput:
			ret.requests = append(ret.requests, RequestOnLedgerFromOutput(o, sender))
		case *ledgerstate.AliasOutput:
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
func (tx *ParsedTransaction) ChainOutput() *ledgerstate.AliasOutput {
	return tx.chainOutput
}

func (tx *ParsedTransaction) SenderAddress() ledgerstate.Address {
	return tx.senderAddr
}

func (tx *ParsedTransaction) Requests() []*RequestOnLedger {
	return tx.requests
}

// endregion /////////////////////////////////////////////////////////////////

// region RequestOnLedger //////////////////////////////////////////////////////////////////

type RequestOnLedger struct {
	outputObj       *ledgerstate.ExtendedLockedOutput
	senderAddress   ledgerstate.Address
	minted          map[ledgerstate.Color]uint64
	requestMetadata RequestMetadata
	solidArgs       dict.Dict
}

// implements coretypes.Request interface
var _ coretypes.Request = &RequestOnLedger{}

// RequestDataFromOutput
func RequestOnLedgerFromOutput(output *ledgerstate.ExtendedLockedOutput, senderAddr ledgerstate.Address, minted ...map[ledgerstate.Color]uint64) *RequestOnLedger {
	ret := &RequestOnLedger{outputObj: output, senderAddress: senderAddr}
	ret.requestMetadata = *RequestMetadataFromBytes(output.GetPayload())
	ret.minted = make(map[ledgerstate.Color]uint64, 0)
	if len(minted) > 0 {
		for k, v := range minted[0] {
			ret.minted[k] = v
		}
	}
	return ret
}

// RequestsOnLedgerFromTransaction creates RequestOnLedger object from transaction and output index
func RequestsOnLedgerFromTransaction(tx *ledgerstate.Transaction, targetAddr ledgerstate.Address) ([]*RequestOnLedger, error) {
	senderAddr, err := utxoutil.GetSingleSender(tx)
	if err != nil {
		return nil, err
	}
	mintedAmounts := utxoutil.GetMintedAmounts(tx)
	ret := make([]*RequestOnLedger, 0)
	for _, o := range tx.Essence().Outputs() {
		if out, ok := o.(*ledgerstate.ExtendedLockedOutput); ok {
			if out.Address().Equals(targetAddr) {
				ret = append(ret, RequestOnLedgerFromOutput(out, senderAddr, mintedAmounts))
			}
		}
	}
	return ret, nil
}

func (req *RequestOnLedger) ID() coretypes.RequestID {
	return coretypes.RequestID(req.Output().ID())
}

func (req *RequestOnLedger) IsFeePrepaid() bool {
	return false
}

func (req *RequestOnLedger) Output() ledgerstate.Output {
	return req.output()
}

func (req *RequestOnLedger) output() *ledgerstate.ExtendedLockedOutput {
	return req.outputObj
}

func (req *RequestOnLedger) TimeLock() time.Time {
	return req.outputObj.TimeLock()
}

func (req *RequestOnLedger) SenderAddress() ledgerstate.Address {
	return req.senderAddress
}

func (req *RequestOnLedger) SenderAccount() *coretypes.AgentID {
	return coretypes.NewAgentID(req.senderAddress, req.requestMetadata.SenderContract())
}

func (req *RequestOnLedger) SetMetadata(d *RequestMetadata) {
	req.requestMetadata = *d.Clone()
}

func (req *RequestOnLedger) GetMetadata() *RequestMetadata {
	return &req.requestMetadata
}

// Target returns target contract and target entry point
func (req *RequestOnLedger) Target() (coretypes.Hname, coretypes.Hname) {
	return req.requestMetadata.targetContract, req.requestMetadata.entryPoint
}

func (req *RequestOnLedger) MintColor() ledgerstate.Color {
	return blake2b.Sum256(req.Output().ID().Bytes())
}

func (req *RequestOnLedger) MintedAmounts() map[ledgerstate.Color]uint64 {
	return req.minted
}

// Args returns solid args if decoded already or nil otherwise
func (req *RequestOnLedger) Params() dict.Dict {
	return req.solidArgs
}

func (req *RequestOnLedger) Tokens() *ledgerstate.ColoredBalances {
	return req.outputObj.Balances()
}

// SolidifyArgs return true if solidified successfully
func (req *RequestOnLedger) SolidifyArgs(reg coretypes.BlobCache) (bool, error) {
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

func (req *RequestOnLedger) Short() string {
	return req.outputObj.ID().Base58()[:6] + ".."
}

// endregion /////////////////////////////////////////////////////////////////

// region RequestMetadata  ///////////////////////////////////////////////////////

// RequestMetadata represents content of the data payload of the output
type RequestMetadata struct {
	err            error
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
		args: requestargs.RequestArgs(dict.New()),
	}
}

func RequestMetadataFromBytes(data []byte) *RequestMetadata {
	ret := NewRequestMetadata()
	ret.err = ret.Read(bytes.NewReader(data))
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
	return p.err == nil
}

func (p *RequestMetadata) ParsedError() error {
	return p.err
}

func (p *RequestMetadata) SenderContract() coretypes.Hname {
	if !p.ParsedOk() {
		return 0
	}
	return p.senderContract
}

func (p *RequestMetadata) TargetContract() coretypes.Hname {
	if !p.ParsedOk() {
		return 0
	}
	return p.targetContract
}

func (p *RequestMetadata) EntryPoint() coretypes.Hname {
	if !p.ParsedOk() {
		return 0
	}
	return p.entryPoint
}

func (p *RequestMetadata) Args() requestargs.RequestArgs {
	if !p.ParsedOk() {
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
