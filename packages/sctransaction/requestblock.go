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
	// address of the target smart contract
	address address.Address
	// request code
	reqCode coretypes.EntryPointCode
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

// RequestBlock

func NewRequestBlock(addr address.Address, reqCode coretypes.EntryPointCode) *RequestBlock {
	return &RequestBlock{
		address: addr,
		reqCode: reqCode,
		args:    kv.NewMap(),
	}
}

func (req *RequestBlock) Clone() *RequestBlock {
	if req == nil {
		return nil
	}
	ret := NewRequestBlock(req.address, req.reqCode)
	ret.args = req.args.Clone()
	return ret
}

func (req *RequestBlock) Address() address.Address {
	return req.address
}

func (req *RequestBlock) SetArgs(args kv.Map) {
	if args != nil {
		req.args = args.Clone()
	}
}

func (req *RequestBlock) Args() kv.RCodec {
	return req.args.Codec()
}

func (req *RequestBlock) EntryPointCode() coretypes.EntryPointCode {
	return req.reqCode
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
	if _, err := w.Write(req.address.Bytes()); err != nil {
		return err
	}
	if err := util.WriteUint32(w, req.timelock); err != nil {
		return err
	}
	if err := req.reqCode.Write(w); err != nil {
		return err
	}
	if err := req.args.Write(w); err != nil {
		return err
	}
	return nil
}

func (req *RequestBlock) Read(r io.Reader) error {
	if err := util.ReadAddress(r, &req.address); err != nil {
		return fmt.Errorf("error while reading address: %v", err)
	}
	if err := util.ReadUint32(r, &req.timelock); err != nil {
		return err
	}
	if err := req.reqCode.Read(r); err != nil {
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

func (ref *RequestRef) Sender() *address.Address {
	return ref.Tx.MustProperties().Sender()
}
