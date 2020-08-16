package apilib

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/multicall"
	"math/rand"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/plugins/webapi/dkgapi"
	"github.com/pkg/errors"
)

// GenerateNewDistributedKeySetOld calls nodes one after the other to produce distributed key set
func GenerateNewDistributedKeySetOld(nodes []string, n, t uint16) (*address.Address, error) {
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
	for i, host := range nodes {
		par := params
		par.Index = uint16(i)
		resp, err := callNewKey(host, par)
		if err != nil {
			return nil, err
		}
		if len(resp.PriShares) != int(params.N) {
			return nil, errors.New("apilib: len(resp.PriShares) != int(params.N)")
		}
		priSharesMatrix[i] = resp.PriShares
	}

	// aggregate private shares
	pubShares := make([]string, params.N)
	priSharesCol := make([]string, params.N)
	for col, host := range nodes {
		for row := range nodes {
			priSharesCol[row] = priSharesMatrix[row][col]
		}
		resp, err := callAggregate(host, dkgapi.AggregateDKSRequest{
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
	for _, host := range nodes {
		addr, err := callCommit(host, dkgapi.CommitDKSRequest{
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

// GenerateNewDistributedKeySet calls nodes in parallel to produce distributed key set
func GenerateNewDistributedKeySet(hosts []string, n, t uint16) (*address.Address, error) {
	if len(hosts) != int(n) {
		return nil, errors.New("wrong params")
	}
	if err := tcrypto.ValidateDKSParams(t, n, 0); err != nil {
		return nil, err
	}

	if util.ContainsDuplicates(hosts) {
		return nil, fmt.Errorf("duplicate hosts")
	}
	// temporary numeric id during DKG
	// generate new key shares
	// results in the matrix
	tmpId := rand.Int()

	funs := make([]func() error, len(hosts))
	priSharesMatrix := make([][]string, n)
	for i, host := range hosts {
		h := host
		idx := i
		funs[i] = func() error {
			resp, err := callNewKey(h, dkgapi.NewDKSRequest{
				TmpId: tmpId,
				N:     n,
				T:     t,
				Index: uint16(idx),
			})
			if err != nil {
				return err
			}
			if len(resp.PriShares) != int(n) {
				return errors.New("inconsistency: len(resp.PriShares) != int(params.N)")
			}
			priSharesMatrix[idx] = resp.PriShares
			return nil
		}
	}
	succ, errs := multicall.MultiCall(funs, 2*time.Second)
	if !succ {
		return nil, multicall.WrapErrors(errs)
	}

	// aggregate private shares
	pubShares := make([]string, n)

	for col, host := range hosts {
		priSharesCol := make([]string, n)
		for row := range hosts {
			priSharesCol[row] = priSharesMatrix[row][col]
		}
		h := host
		c := col
		funs[col] = func() error {
			resp, err := callAggregate(h, dkgapi.AggregateDKSRequest{
				TmpId:     tmpId,
				Index:     uint16(c),
				PriShares: priSharesCol,
			})
			if err != nil {
				return err
			}
			pubShares[c] = resp.PubShare
			return nil
		}
	}
	succ, errs = multicall.MultiCall(funs, 2*time.Second)
	if !succ {
		return nil, multicall.WrapErrors(errs)
	}

	// commit keys
	var addrRet *address.Address
	for i, host := range hosts {
		h := host
		funs[i] = func() error {
			addr, err := callCommit(h, dkgapi.CommitDKSRequest{
				TmpId:     tmpId,
				PubShares: pubShares,
			})
			if err != nil {
				return err
			}
			if addrRet != nil && *addrRet != *addr {
				return errors.New("key commit returned different addresses from different nodes")
			}
			if addr.Version() != address.VersionBLS {
				return errors.New("key commit returned non-BLS address")
			}
			addrRet = addr
			return nil
		}
	}
	succ, errs = multicall.MultiCall(funs, 2*time.Second)
	if !succ {
		return nil, multicall.WrapErrors(errs)
	}
	return addrRet, nil
}

// retrieves public info about key with specific address
func GetPublicKeyInfo(nodes []string, address *address.Address) []*dkgapi.GetPubKeyInfoResponse {
	params := dkgapi.GetPubKeyInfoRequest{
		Address: address.String(),
	}
	ret := make([]*dkgapi.GetPubKeyInfoResponse, len(nodes))
	for i, host := range nodes {
		ret[i] = callGetPubKeyInfo(host, params)
	}
	return ret
}

func ExportDKShare(node string, address *address.Address) (string, error) {
	return callExportDKShare(node, dkgapi.ExportDKShareRequest{
		Address: address.String(),
	})
}

func ImportDKShare(node string, base58blob string) error {
	return callImportDKShare(node, dkgapi.ImportDKShareRequest{
		Blob: base58blob,
	})
}
