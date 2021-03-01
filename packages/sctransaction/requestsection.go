package sctransaction

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"io"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
)

// FIXME timelock uint32 ref Year 2038 problem https://en.wikipedia.org/wiki/Year_2038_problem
// signed int32 can store values uo to 03:14:07 UTC on 19 January 2038
// But if we use uint32 we should extend the range twice, something like until 2092. Not a problem ???
// other wise we can use values in 'timelock' field of the request block
// counting from 2020.01.01 Then it would extend until 2140 or so

type RequestSection struct {
	// senderAddress contract index
	// - if state block present, it is hname of the sending contract in the chain of which state transaction it is
	// - if state block is absent, it is uninterpreted (it means requests are sent by the wallet)
	senderContractHname coretypes.Hname
	// ID of the target smart contract
	targetContractID coretypes.ContractID
	// entry point code
	entryPoint coretypes.Hname
	// timelock in Unix seconds.
	// Request will only be processed when time reaches
	// specified moment. It is guaranteed that timestamp of the state transaction which
	// settles the request is greater or equal to the request timelock.
	// 0 timelock naturally means it has no effect
	timelock uint32
	// request arguments, not decoded yet wrt blobRefs
	args requestargs.RequestArgs
	// decoded args, if not nil. If nil, it means it wasn't
	// successfully decoded yet and can't be used in the batch for calculations in VM
	solidArgs dict.Dict
	// references corresponding output with all attached tokens.
	output uint16
}

type RequestRef struct {
	Tx    *TransactionEssence
	Index uint16
}

// RequestSection creates new request block
func NewRequestSection(senderContractHname coretypes.Hname, targetContract coretypes.ContractID, entryPointCode coretypes.Hname) *RequestSection {
	return &RequestSection{
		senderContractHname: senderContractHname,
		targetContractID:    targetContract,
		entryPoint:          entryPointCode,
		args:                requestargs.New(nil),
	}
}

func (req *RequestSection) String() string {
	return fmt.Sprintf("[[sender contract: %s, target: %s, entry point: '%s', args: %s]]",
		req.senderContractHname.String(), req.targetContractID.String(), req.entryPoint.String(), req.args.String())
}

func (req *RequestSection) Clone() *RequestSection {
	if req == nil {
		return nil
	}
	return &RequestSection{
		senderContractHname: req.senderContractHname,
		targetContractID:    req.targetContractID,
		entryPoint:          req.entryPoint,
		timelock:            req.timelock,
		args:                req.args.Clone(),
		solidArgs:           req.solidArgs.Clone(),
		output:              req.output,
	}
}

func (req *RequestSection) SenderContractHname() coretypes.Hname {
	return req.senderContractHname
}

func (req *RequestSection) Target() coretypes.ContractID {
	return req.targetContractID
}

// WithArgs sets encoded args
func (req *RequestSection) WithArgs(args requestargs.RequestArgs) *RequestSection {
	req.args = args
	return req
}

// SolidArgs returns solid args if decoded already or nil otherwise
func (req *RequestSection) SolidArgs() dict.Dict {
	return req.solidArgs
}

// SolidifyArgs return true if solidified successfully
func (req *RequestSection) SolidifyArgs(reg coretypes.BlobCache) (bool, error) {
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

func (req *RequestSection) EntryPointCode() coretypes.Hname {
	return req.entryPoint
}

func (req *RequestSection) Timelock() uint32 {
	return req.timelock
}

func (req *RequestSection) OutputIndex() uint16 {
	return req.output
}

func (req *RequestSection) WithTimeLock(tl uint32) *RequestSection {
	req.timelock = tl
	return req
}

func (req *RequestSection) WithOutputIndex(index uint16) *RequestSection {
	req.output = index
	return req
}

func (req *RequestSection) WithTimeLockUntil(deadline time.Time) *RequestSection {
	return req.WithTimeLock(uint32(deadline.Unix()))
}

// encoding

func (req *RequestSection) Write(w io.Writer) error {
	if err := req.senderContractHname.Write(w); err != nil {
		return err
	}
	if err := req.targetContractID.Write(w); err != nil {
		return err
	}
	if err := util.WriteUint32(w, req.timelock); err != nil {
		return err
	}
	if err := req.entryPoint.Write(w); err != nil {
		return err
	}
	if err := req.args.Write(w); err != nil {
		return err
	}
	if err := util.WriteUint16(w, req.output); err != nil {
		return err
	}
	return nil
}

func (req *RequestSection) Read(r io.Reader) error {
	if err := req.senderContractHname.Read(r); err != nil {
		return err
	}
	if err := req.targetContractID.Read(r); err != nil {
		return err
	}
	if err := util.ReadUint32(r, &req.timelock); err != nil {
		return err
	}
	if err := req.entryPoint.Read(r); err != nil {
		return err
	}
	req.args = requestargs.New(nil)
	if err := req.args.Read(r); err != nil {
		return err
	}
	if err := util.ReadUint16(r, &req.output); err != nil {
		return err
	}
	return nil
}

// request ref

func (ref *RequestRef) RequestSection() *RequestSection {
	return ref.Tx.Requests()[ref.Index]
}

func (ref *RequestRef) RequestID() *coretypes.RequestID {
	ret := coretypes.NewRequestID(ref.Tx.ID(), ref.Index)
	return &ret
}

func (ref *RequestRef) SenderContractHname() coretypes.Hname {
	return ref.RequestSection().senderContractHname
}

func (ref *RequestRef) SenderAddress() ledgerstate.Address {
	return ref.Tx.MustProperties().SenderAddress()
}

func (ref *RequestRef) SenderContractID() (ret coretypes.ContractID, err error) {
	if _, ok := ref.Tx.State(); !ok {
		err = fmt.Errorf("request wasn't sent by the smart contract: %s", ref.RequestID().String())
		return
	}
	ret = coretypes.NewContractID(ref.SenderAddress().Array(), ref.SenderContractHname())
	return
}

func (ref *RequestRef) SenderAgentID() coretypes.AgentID {
	if contractID, err := ref.SenderContractID(); err == nil {
		return coretypes.NewAgentIDFromContractID(contractID)
	}
	return coretypes.NewAgentIDFromAddress(*ref.SenderAddress())
}
