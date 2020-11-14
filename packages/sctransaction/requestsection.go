package sctransaction

import (
	"fmt"
	"io"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
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
	// - if state block present, it is index of the sending contracts
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
	// input arguments in the form of variable/value pairs
	args dict.Dict
}

type RequestRef struct {
	Tx    *Transaction
	Index uint16
}

// RequestSection creates new request block
func NewRequestSection(senderContractHname coretypes.Hname, targetContract coretypes.ContractID, entryPointCode coretypes.Hname) *RequestSection {
	return &RequestSection{
		senderContractHname: senderContractHname,
		targetContractID:    targetContract,
		entryPoint:          entryPointCode,
		args:                dict.NewDict(),
	}
}

// NewRequestSectionByWallet same as NewRequestSection but assumes senderAddress index is 0
func NewRequestSectionByWallet(targetContract coretypes.ContractID, entryPointCode coretypes.Hname) *RequestSection {
	return NewRequestSection(0, targetContract, entryPointCode)
}

func (req *RequestSection) String() string {
	return fmt.Sprintf("[[senderAddress: %s, target: %s, entry point: '%s', args: %s]]",
		req.senderContractHname.String(), req.targetContractID.String(), req.entryPoint.String(), req.args.String())
}

func (req *RequestSection) Clone() *RequestSection {
	if req == nil {
		return nil
	}
	ret := NewRequestSection(req.senderContractHname, req.targetContractID, req.entryPoint)
	ret.args = req.args.Clone()
	return ret
}

func (req *RequestSection) SenderContractHname() coretypes.Hname {
	return req.senderContractHname
}

func (req *RequestSection) Target() coretypes.ContractID {
	return req.targetContractID
}

func (req *RequestSection) SetArgs(args dict.Dict) {
	if args != nil {
		req.args = args.Clone()
	}
}

func (req *RequestSection) Args() codec.ImmutableCodec {
	return codec.NewCodec(req.args)
}

func (req *RequestSection) EntryPointCode() coretypes.Hname {
	return req.entryPoint
}

func (req *RequestSection) Timelock() uint32 {
	return req.timelock
}

func (req *RequestSection) WithTimelock(tl uint32) *RequestSection {
	req.timelock = tl
	return req
}

func (req *RequestSection) WithTimelockUntil(deadline time.Time) *RequestSection {
	return req.WithTimelock(uint32(deadline.Unix()))
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
	req.args = dict.NewDict()
	if err := req.args.Read(r); err != nil {
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

// request block is authorised if the containing transaction's inputs contain owner's address
func (ref *RequestRef) IsAuthorised(ownerAddr *address.Address) bool {
	// would be better to have something like tx.IsSignedBy(addr)

	if !ref.Tx.Transaction.SignaturesValid() {
		return false // not needed, just in case
	}
	auth := false
	ref.Tx.Transaction.Inputs().ForEach(func(oid valuetransaction.OutputID) bool {
		if oid.Address() == *ownerAddr {
			auth = true
			return false
		}
		return true
	})
	return auth
}

func (ref *RequestRef) SenderContractHname() coretypes.Hname {
	return ref.RequestSection().senderContractHname
}

func (ref *RequestRef) SenderAddress() *address.Address {
	return ref.Tx.MustProperties().SenderAddress()
}

func (ref *RequestRef) SenderContractID() (ret coretypes.ContractID, err error) {
	if _, ok := ref.Tx.State(); !ok {
		err = fmt.Errorf("request not sent by the smart contract: %s", ref.RequestID().String())
		return
	}
	ret = coretypes.NewContractID((coretypes.ChainID)(*ref.SenderAddress()), ref.SenderContractHname())
	return
}

func (ref *RequestRef) SenderAgentID() coretypes.AgentID {
	if contractID, err := ref.SenderContractID(); err == nil {
		return coretypes.NewAgentIDFromContractID(contractID)
	}
	return coretypes.NewAgentIDFromAddress(*ref.SenderAddress())
}
