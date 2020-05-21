package sctransaction

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/variables"
	"io"
)

const RequestIdSize = hashing.HashSize + 2

type RequestId [RequestIdSize]byte

type RequestBlock struct {
	address address.Address
	// small variable state with variable/value pairs
	vars variables.Variables
}

// RequestBlock

func NewRequestBlock(addr address.Address) *RequestBlock {
	return &RequestBlock{
		address: addr,
		vars:    variables.New(nil),
	}
}

func (req *RequestBlock) Address() address.Address {
	return req.address
}

func (req *RequestBlock) Variables() variables.Variables {
	return req.vars
}

// encoding
// important: each block starts with 65 bytes of scid

func (req *RequestBlock) Write(w io.Writer) error {
	if _, err := w.Write(req.address.Bytes()); err != nil {
		return err
	}
	if err := req.vars.Write(w); err != nil {
		return err
	}
	return nil
}

func (req *RequestBlock) Read(r io.Reader) error {
	if err := util.ReadAddress(r, &req.address); err != nil {
		return fmt.Errorf("error while reading address: %v", err)
	}
	req.vars = variables.New(nil)
	if err := req.vars.Read(r); err != nil {
		return err
	}
	return nil
}

func NewRequestId(txid valuetransaction.ID, index uint16) (ret RequestId) {
	copy(ret[:valuetransaction.IDLength], txid.Bytes())
	copy(ret[valuetransaction.IDLength:], util.Uint16To2Bytes(index)[:])
	return
}

func NewRandomRequestId(index uint16) (ret RequestId) {
	copy(ret[:valuetransaction.IDLength], hashing.RandomHash(nil).Bytes())
	copy(ret[valuetransaction.IDLength:], util.Uint16To2Bytes(index)[:])
	return
}

func (rid *RequestId) Bytes() []byte {
	return rid[:]
}

func (rid *RequestId) TransactionId() *valuetransaction.ID {
	var ret valuetransaction.ID
	copy(ret[:], rid[:valuetransaction.IDLength])
	return &ret
}

func (rid *RequestId) Index() uint16 {
	return util.Uint16From2Bytes(rid[valuetransaction.IDLength:])
}

func (rid *RequestId) Write(w io.Writer) error {
	_, err := w.Write(rid.Bytes())
	return err
}

func (rid *RequestId) Read(r io.Reader) error {
	n, err := r.Read(rid[:])
	if err != nil {
		return err
	}
	if n != RequestIdSize {
		return errors.New("not enough data for RequestId")
	}
	return nil
}

func (rid *RequestId) String() string {
	return fmt.Sprintf("[%d]%s", rid.Index(), rid.TransactionId().String())
}

func (rid *RequestId) Short() string {
	return rid.String()[:8] + ".."
}
