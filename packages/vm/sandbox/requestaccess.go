package sandbox

import (
	"bytes"
	"sort"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/table"
)

// access to the request block
type requestWrapper struct {
	ref *sctransaction.RequestRef
}

func (r *requestWrapper) ID() sctransaction.RequestId {
	return *r.ref.RequestId()
}

func (r *requestWrapper) Code() sctransaction.RequestCode {
	return r.ref.RequestBlock().RequestCode()
}

func (r *requestWrapper) Args() table.RCodec {
	return r.ref.RequestBlock().Args()
}

func (r *requestWrapper) IsAuthorisedByAddress(addr *address.Address) bool {
	found := false
	r.ref.Tx.Inputs().ForEachAddress(func(currentAddress address.Address) bool {
		if currentAddress == *addr {
			found = true
			return false
		}
		return true
	})
	return found
}

// addresses of request transaction inputs
func (r *requestWrapper) Senders() []address.Address {
	ret := make([]address.Address, 0)
	r.ref.Tx.Inputs().ForEachAddress(func(currentAddress address.Address) bool {
		ret = append(ret, currentAddress)
		return true
	})
	// sort to be deterministic
	sort.Slice(ret, func(i, j int) bool {
		return bytes.Compare(ret[i][:], ret[j][:]) < 0
	})

	return ret
}
