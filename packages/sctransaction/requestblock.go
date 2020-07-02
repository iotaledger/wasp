package sctransaction

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/table"
	"github.com/iotaledger/wasp/packages/util"
)

const RequestIdSize = hashing.HashSize + 2

type RequestId [RequestIdSize]byte

type RequestBlock struct {
	address address.Address
	// request code
	reqCode RequestCode
	// small variable state with variable/value pairs
	args table.MemTable
}

type RequestRef struct {
	Tx    *Transaction
	Index uint16
}

// RequestBlock

func NewRequestBlock(addr address.Address, reqCode RequestCode) *RequestBlock {
	return &RequestBlock{
		address: addr,
		reqCode: reqCode,
		args:    table.NewMemTable(),
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

func (req *RequestBlock) SetArgs(args table.MemTable) {
	req.args = args.Clone()
}

func (req *RequestBlock) Args() table.RCodec {
	return req.args.Codec()
}

func (req *RequestBlock) RequestCode() RequestCode {
	return req.reqCode
}

func (req *RequestBlock) String(reqId *RequestId) string {
	return fmt.Sprintf("Request: %s to: %s, code: %s\n%s",
		reqId.Short(), req.Address().String(), req.reqCode.String(), req.args.String())
}

func NewRequestIdFromString(reqIdStr string) (ret RequestId, err error) {
	splitStr := strings.Split(reqIdStr, "]")
	if len(splitStr) != 2 {
		err = fmt.Errorf("wrong request id string")
		return
	}
	indexStr := splitStr[0][1:]
	indexInt, err := strconv.Atoi(indexStr)
	if err != nil {
		err = fmt.Errorf("wrong request id string")
		return
	}
	index := uint16(indexInt)
	txid, err := valuetransaction.IDFromBase58(splitStr[1])
	if err != nil {
		return
	}
	ret = NewRequestId(txid, index)
	return
}

// encoding
// important: each block starts with 65 bytes of scid

func (req *RequestBlock) Write(w io.Writer) error {
	if _, err := w.Write(req.address.Bytes()); err != nil {
		return err
	}
	if err := util.WriteUint16(w, uint16(req.reqCode)); err != nil {
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
	var rc uint16
	if err := util.ReadUint16(r, &rc); err != nil {
		return err
	}
	req.reqCode = RequestCode(rc)

	req.args = table.NewMemTable()
	if err := req.args.Read(r); err != nil {
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

// request ref

func (ref *RequestRef) RequestBlock() *RequestBlock {
	return ref.Tx.Requests()[ref.Index]
}

func (ref *RequestRef) RequestId() *RequestId {
	ret := NewRequestId(ref.Tx.ID(), ref.Index)
	return &ret
}

func TakeRequestIds(lst []RequestRef) []RequestId {
	ret := make([]RequestId, len(lst))
	for i := range ret {
		ret[i] = *lst[i].RequestId()
	}
	return ret
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
