package sctransaction

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"io"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
)

// FIXME timelock uint32 ref Year 2038 problem https://en.wikipedia.org/wiki/Year_2038_problem
// signed int32 can store values uo to 03:14:07 UTC on 19 January 2038
// But if we use uint32 we should extend the range twice, something like until 2092. Not a problem ???
// other wise we can use values in 'timelock' field of the request block
// counting from 2020.01.01 Then it would extend until 2140 or so

type RequestBlock struct {
	// sender contract index
	// - if state block present, it is index of the sending contracts
	// - if state block is absent, it is uninterpreted (it means requests are sent by the wallet)
	senderContractIndex coretypes.Uint16
	// ID of the target smart contract
	targetContractID coretypes.ContractID
	// entry point code
	entryPoint coretypes.EntryPointCode
	// timelock in Unix seconds.
	// Request will only be processed when time reaches
	// specified moment. It is guaranteed that timestamp of the state transaction which
	// settles the request is greater or equal to the request timelock.
	// 0 timelock naturally means it has no effect
	timelock uint32
	// input arguments in the form of variable/value pairs
	args kv.Map
}

type RequestRef struct {
	Tx    *Transaction
	Index coretypes.Uint16
}

// RequestBlock creates new request block
func NewRequestBlock(senderContractIndex coretypes.Uint16, targetContract coretypes.ContractID, entryPointCode coretypes.EntryPointCode) *RequestBlock {
	return &RequestBlock{
		senderContractIndex: senderContractIndex,
		targetContractID:    targetContract,
		entryPoint:          entryPointCode,
		args:                kv.NewMap(),
	}
}

// NewRequestBlockByWallet same as NewRequestBlock but assumes sender index is 0
func NewRequestBlockByWallet(targetContract coretypes.ContractID, entryPointCode coretypes.EntryPointCode) *RequestBlock {
	return NewRequestBlock(0, targetContract, entryPointCode)
}

func (req *RequestBlock) Clone() *RequestBlock {
	if req == nil {
		return nil
	}
	ret := NewRequestBlock(req.senderContractIndex, req.targetContractID, req.entryPoint)
	ret.args = req.args.Clone()
	return ret
}

func (req *RequestBlock) SenderContractIndex() uint16 {
	return (uint16)(req.senderContractIndex)
}

func (req *RequestBlock) Target() coretypes.ContractID {
	return req.targetContractID
}

func (req *RequestBlock) SetArgs(args kv.Map) {
	if args != nil {
		req.args = args.Clone()
	}
}

func (req *RequestBlock) Args() kv.ImmutableCodec {
	return req.args.Codec()
}

func (req *RequestBlock) EntryPointCode() coretypes.EntryPointCode {
	return req.entryPoint
}

func (req *RequestBlock) Timelock() uint32 {
	return req.timelock
}

func (req *RequestBlock) WithTimelock(tl uint32) *RequestBlock {
	req.timelock = tl
	return req
}

func (req *RequestBlock) WithTimelockUntil(deadline time.Time) *RequestBlock {
	return req.WithTimelock(uint32(deadline.Unix()))
}

// encoding

func (req *RequestBlock) Write(w io.Writer) error {
	if err := req.senderContractIndex.Write(w); err != nil {
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

func (req *RequestBlock) Read(r io.Reader) error {
	if err := req.senderContractIndex.Read(r); err != nil {
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
	req.args = kv.NewMap()
	if err := req.args.Read(r); err != nil {
		return err
	}
	return nil
}

// request ref

func (ref *RequestRef) RequestBlock() *RequestBlock {
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

func (ref *RequestRef) SenderAddress() *address.Address {
	return ref.Tx.MustProperties().Sender()
}

func (ref *RequestRef) SenderContractID() (ret coretypes.ContractID, err error) {
	if _, ok := ref.Tx.State(); !ok {
		err = fmt.Errorf("request not sent by the smart contract: %s", ref.RequestID().String())
		return
	}
	ret = coretypes.NewContractID((coretypes.ChainID)(*ref.SenderAddress()), ref.Index)
	return
}
