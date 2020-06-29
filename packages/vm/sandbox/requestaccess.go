package sandbox

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"sort"
)

// access to the request block

func (r *requestWrapper) ID() sctransaction.RequestId {
	return *r.ref.RequestId()
}

func (r *requestWrapper) Code() sctransaction.RequestCode {
	return r.ref.RequestBlock().RequestCode()
}

func (r *requestWrapper) GetInt64(name string) (int64, bool) {
	return r.ref.RequestBlock().Args().MustGetInt64(name)
}

func (r *requestWrapper) GetString(name string) (string, bool) {
	return r.ref.RequestBlock().Args().GetString(name)
}

func (r *requestWrapper) GetAddressValue(name string) (address.Address, bool) {
	return r.ref.RequestBlock().Args().MustGetAddress(name)
}

func (r *requestWrapper) GetHashValue(name string) (hashing.HashValue, bool) {
	return r.ref.RequestBlock().Args().MustGetHashValue(name)
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
