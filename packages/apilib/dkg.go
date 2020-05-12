package apilib

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/plugins/webapi/dkgapi"
	"github.com/pkg/errors"
	"math/rand"
)

func GenerateNewDistributedKeySet(nodes []*registry.PortAddr, n, t uint16) (*address.Address, error) {
	if len(nodes) != int(n) {
		return nil, errors.New("wrong params")
	}
	if err := tcrypto.ValidateDKSParams(t, n, 0); err != nil {
		return nil, err
	}
	// temporary numeric id during DKG
	params := dkgapi.NewDKSRequest{
		TmpId: rand.Int(),
		N:     n,
		T:     t,
	}
	// generate new key shares
	// results in the matrix
	priSharesMatrix := make([][]string, params.N)
	for i, pa := range nodes {
		par := params
		par.Index = uint16(i)
		resp, err := callNewKey(pa.Addr, pa.Port, par)
		if err != nil {
			return nil, err
		}
		if len(resp.PriShares) != int(params.N) {
			return nil, errors.New("len(resp.PriShares) != int(params.N)")
		}
		priSharesMatrix[i] = resp.PriShares
	}

	// aggregate private shares
	pubShares := make([]string, params.N)
	priSharesCol := make([]string, params.N)
	for col, pa := range nodes {
		for row := range nodes {
			priSharesCol[row] = priSharesMatrix[row][col]
		}
		resp, err := callAggregate(pa.Addr, pa.Port, dkgapi.AggregateDKSRequest{
			TmpId:     params.TmpId,
			Index:     uint16(col),
			PriShares: priSharesCol,
		})
		if err != nil {
			return nil, err
		}
		pubShares[col] = resp.PubShare
	}

	// commit keys
	var addrRet *address.Address
	for _, pa := range nodes {
		addr, err := callCommit(pa.Addr, pa.Port, dkgapi.CommitDKSRequest{
			TmpId:     params.TmpId,
			PubShares: pubShares,
		})
		if err != nil {
			return nil, err
		}
		if addrRet != nil && !bytes.Equal(addrRet.Bytes(), addr.Bytes()) {
			return nil, errors.New("key commit returned different addresses from different nodes")
		}
		if addr.Version() != address.VersionBLS {
			return nil, errors.New("key commit returned non-BLS address")
		}
		addrRet = addr
	}
	return addrRet, nil
}

// retrieves public info about key with specific address
func GetPublicKeyInfo(nodes []*registry.PortAddr, address *address.Address) []*dkgapi.GetPubKeyInfoResponse {
	params := dkgapi.GetPubKeyInfoRequest{
		Address: address.String(),
	}
	ret := make([]*dkgapi.GetPubKeyInfoResponse, len(nodes))
	for i, pa := range nodes {
		ret[i] = callGetPubKeyInfo(pa.Addr, pa.Port, params)
	}
	return ret
}
