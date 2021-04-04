package sctransaction

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"io"
)

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

