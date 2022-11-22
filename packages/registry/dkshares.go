// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/iotaledger/hive.go/core/generics/onchangemap"
	"github.com/iotaledger/hive.go/core/ioutils"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util"
)

var ErrDKShareNotFound = errors.New("dkShare not found")

type DKSharesRegistry struct {
	storeOnChangeMap *onchangemap.OnChangeMap[string, *util.ComparableAddress, tcrypto.DKShare]
}

var _ DKShareRegistryProvider = &DKSharesRegistry{}

// NewDKSharesRegistry creates new instance of the DKShare registry implementation.
func NewDKSharesRegistry(storeCallback func(dkShares []tcrypto.DKShare) error) *DKSharesRegistry {
	return &DKSharesRegistry{
		storeOnChangeMap: onchangemap.NewOnChangeMap[string, *util.ComparableAddress](storeCallback),
	}
}

func (p *DKSharesRegistry) EnableStoreOnChange() {
	p.storeOnChangeMap.CallbackEnabled(true)
}

func (p *DKSharesRegistry) SaveDKShare(dkShare tcrypto.DKShare) error {
	return p.storeOnChangeMap.Add(dkShare)
}

func (p *DKSharesRegistry) LoadDKShare(sharedAddress iotago.Address) (tcrypto.DKShare, error) {
	dkShare, err := p.storeOnChangeMap.Get(util.NewComparableAddress(sharedAddress))
	if err != nil {
		return dkShare, ErrDKShareNotFound
	}
	return dkShare, nil
}

type jsonDKShares struct {
	DKShares []tcrypto.DKShare `json:"dkShares"`
}

type jsonDKSharesRaw struct {
	DKShares []*json.RawMessage `json:"dkShares"`
}

func LoadDKSharesJSONFromFile(filePath string, dkSharesRegistry *DKSharesRegistry, nodePrivKey *cryptolib.PrivateKey) error {
	tmpDKSharesRaw := &jsonDKSharesRaw{}
	if err := ioutils.ReadJSONFromFile(filePath, tmpDKSharesRaw); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to unmarshal json file: %w", err)
	}

	for _, dkShareRaw := range tmpDKSharesRaw.DKShares {
		dkShareBytes, err := dkShareRaw.MarshalJSON()
		if err != nil {
			return fmt.Errorf("unable to unmarshal DKShare: %w", err)
		}

		dkShare := tcrypto.NewEmptyDKShare(nodePrivKey, tcrypto.DefaultEd25519Suite(), tcrypto.DefaultBLSSuite())
		if err := dkShare.UnmarshalJSON(dkShareBytes); err != nil {
			return fmt.Errorf("unable to unmarshal DKShare: %w", err)
		}

		if err := dkSharesRegistry.SaveDKShare(dkShare); err != nil {
			return fmt.Errorf("unable to add DKShare to registry: %w", err)
		}
	}

	return nil
}

func WriteDKSharesJSONToFile(filePath string, dkShares []tcrypto.DKShare) error {
	if err := os.MkdirAll(path.Dir(filePath), 0o770); err != nil {
		return fmt.Errorf("unable to create folder \"%s\": %w", path.Dir(filePath), err)
	}

	if err := ioutils.WriteJSONToFile(filePath, &jsonDKShares{DKShares: dkShares}, 0o600); err != nil {
		return fmt.Errorf("unable to marshal json file: %w", err)
	}

	return nil
}
