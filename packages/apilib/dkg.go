package apilib

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/multicall"
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
	params := client.NewDKSRequest{
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
		priShares, err := client.NewWaspClient(host).NewDKShare(par)
		if err != nil {
			return nil, err
		}
		if len(priShares) != int(params.N) {
			return nil, errors.New("apilib: len(priShares) != int(params.N)")
		}
		priSharesMatrix[i] = priShares
	}

	// aggregate private shares
	pubShares := make([]string, params.N)
	priSharesCol := make([]string, params.N)
	for col, host := range nodes {
		for row := range nodes {
			priSharesCol[row] = priSharesMatrix[row][col]
		}
		pubShare, err := client.NewWaspClient(host).AggregateDKShare(client.AggregateDKSRequest{
			TmpId:     params.TmpId,
			Index:     uint16(col),
			PriShares: priSharesCol,
		})
		if err != nil {
			return nil, err
		}
		pubShares[col] = pubShare
	}

	// commit keys
	var addrRet *address.Address
	for _, host := range nodes {
		addr, err := client.NewWaspClient(host).CommitDKShare(client.CommitDKSRequest{
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
			priShares, err := client.NewWaspClient(h).NewDKShare(client.NewDKSRequest{
				TmpId: tmpId,
				N:     n,
				T:     t,
				Index: uint16(idx),
			})
			if err != nil {
				return err
			}
			if len(priShares) != int(n) {
				return errors.New("inconsistency: len(priShares) != int(params.N)")
			}
			priSharesMatrix[idx] = priShares
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
			pubShare, err := client.NewWaspClient(h).AggregateDKShare(client.AggregateDKSRequest{
				TmpId:     tmpId,
				Index:     uint16(c),
				PriShares: priSharesCol,
			})
			if err != nil {
				return err
			}
			pubShares[c] = pubShare
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
			addr, err := client.NewWaspClient(h).CommitDKShare(client.CommitDKSRequest{
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
