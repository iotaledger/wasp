package apilib

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/plugins/webapi/dkgapi"
)

// GetPublicKeyInfoMulti retrieves public info about key with specific address from multiple hosts
func GetPublicKeyInfoMulti(nodes []string, addr *address.Address) []*dkgapi.GetPubKeyInfoResponse {
	ret := make([]*dkgapi.GetPubKeyInfoResponse, len(nodes))
	for i, host := range nodes {
		ret[i] = GetPublicKeyInfo(host, addr)
	}
	return ret
}

// GetPublicKeyInfo retrieves public info about key with specific address from host
func GetPublicKeyInfo(node string, addr *address.Address) *dkgapi.GetPubKeyInfoResponse {
	return callGetPubKeyInfo(node, dkgapi.GetPubKeyInfoRequest{
		Address: addr.String(),
	})
}
