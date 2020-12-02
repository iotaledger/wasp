package testutil

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/dkg"
	"go.dedis.ch/kyber/v3"
)

// DkgRegistryProvider stands for a mock for dkg.RegistryProvider.
type DkgRegistryProvider struct {
	DB    map[string][]byte
	Group kyber.Group
}

// NewDkgRegistryProvider creates new mocked DKG registry provider.
func NewDkgRegistryProvider(group kyber.Group) *DkgRegistryProvider {
	return &DkgRegistryProvider{
		DB:    map[string][]byte{},
		Group: group,
	}
}

// SaveDKShare implements dkg.RegistryProvider.
func (p *DkgRegistryProvider) SaveDKShare(dkShare *dkg.DKShare) error {
	var err error
	var dkShareBytes []byte
	if dkShareBytes, err = dkShare.Bytes(); err != nil {
		return err
	}
	p.DB[dkShare.ChainID.String()] = dkShareBytes
	return nil
}

// LoadDKShare implements dkg.RegistryProvider.
func (p *DkgRegistryProvider) LoadDKShare(chainID *coret.ChainID) (*dkg.DKShare, error) {
	var dkShareBytes = p.DB[chainID.String()]
	if dkShareBytes == nil {
		return nil, fmt.Errorf("DKShare not found for %v", chainID)
	}
	return dkg.DKShareFromBytes(dkShareBytes, p.Group)
}
