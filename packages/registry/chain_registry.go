// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/iotaledger/hive.go/core/generics/lo"
	"github.com/iotaledger/hive.go/core/generics/onchangemap"
	"github.com/iotaledger/hive.go/core/ioutils"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
)

type ComparableChainID struct {
	chainID isc.ChainID
	address iotago.Address
}

func NewComparableChainID(chainID isc.ChainID) *ComparableChainID {
	return &ComparableChainID{
		chainID: chainID,
		address: chainID.AsAddress(),
	}
}

func (c *ComparableChainID) Clone() *ComparableChainID {
	chainID := isc.ChainID{}
	copy(chainID[:], c.chainID[:])

	return &ComparableChainID{
		chainID: chainID,
		address: chainID.AsAddress(),
	}
}

func (c *ComparableChainID) ChainID() isc.ChainID {
	return c.chainID
}

func (c *ComparableChainID) Key() string {
	return c.address.Key()
}

func (c *ComparableChainID) String() string {
	return c.address.Bech32(parameters.L1().Protocol.Bech32HRP)
}

// ChainRecord represents chain the node is participating in.
type ChainRecord struct {
	id     *ComparableChainID
	Active bool
}

func NewChainRecord(chainID isc.ChainID, active bool) *ChainRecord {
	return &ChainRecord{
		id:     NewComparableChainID(chainID),
		Active: active,
	}
}

func (r *ChainRecord) ID() *ComparableChainID {
	return r.id
}

func (r *ChainRecord) ChainID() isc.ChainID {
	return r.id.ChainID()
}

func (r *ChainRecord) Clone() onchangemap.Item[string, *ComparableChainID] {
	return &ChainRecord{
		id:     r.ID().Clone(),
		Active: r.Active,
	}
}

type jsonChainRecord struct {
	ChainID string `json:"chainID"`
	Active  bool   `json:"active"`
}

func (r *ChainRecord) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonChainRecord{
		ChainID: r.ID().String(),
		Active:  r.Active,
	})
}

func (r *ChainRecord) UnmarshalJSON(bytes []byte) error {
	j := &jsonChainRecord{}
	if err := json.Unmarshal(bytes, j); err != nil {
		return err
	}

	if j.ChainID == "" {
		return errors.New("missing chainID")
	}

	_, address, err := iotago.ParseBech32(j.ChainID)
	if err != nil {
		return err
	}

	if address.Type() != iotago.AddressAlias {
		return errors.New("chainID is not an alias address")
	}

	aliasAddress, ok := address.(*iotago.AliasAddress)
	if !ok {
		return errors.New("chainID is not an alias address")
	}

	*r = *NewChainRecord(isc.ChainID(aliasAddress.AliasID()), j.Active)

	return nil
}

type ChainRecordRegistry struct {
	storeOnChangeMap *onchangemap.OnChangeMap[string, *ComparableChainID, *ChainRecord]
}

var _ ChainRecordRegistryProvider = &ChainRecordRegistry{}

// NewChainRecordRegistry creates new instance of the chain registry implementation.
func NewChainRecordRegistry(storeCallback func(chainRecords []*ChainRecord) error) *ChainRecordRegistry {
	return &ChainRecordRegistry{
		storeOnChangeMap: onchangemap.NewOnChangeMap[string, *ComparableChainID](storeCallback),
	}
}

func (p *ChainRecordRegistry) EnableStoreOnChange() {
	p.storeOnChangeMap.CallbackEnabled(true)
}

func (p *ChainRecordRegistry) ChainRecord(chainID isc.ChainID) (*ChainRecord, error) {
	addr := NewComparableChainID(chainID)

	chainRecord, err := p.storeOnChangeMap.Get(addr)
	if err != nil {
		// chain record doesn't exist
		return nil, nil
	}

	return chainRecord, nil
}

func (p *ChainRecordRegistry) ChainRecords() ([]*ChainRecord, error) {
	return lo.Values(p.storeOnChangeMap.All()), nil
}

func (p *ChainRecordRegistry) AddChainRecord(chainRecord *ChainRecord) error {
	return p.storeOnChangeMap.Add(chainRecord)
}

func (p *ChainRecordRegistry) DeleteChainRecord(chainID isc.ChainID) error {
	addr := NewComparableChainID(chainID)
	return p.storeOnChangeMap.Delete(addr)
}

// UpdateChainRecord modifies a ChainRecord in the Registry.
func (p *ChainRecordRegistry) UpdateChainRecord(chainID isc.ChainID, callback func(*ChainRecord) bool) (*ChainRecord, error) {
	addr := NewComparableChainID(chainID)
	return p.storeOnChangeMap.Modify(addr, callback)
}

func (p *ChainRecordRegistry) ActivateChainRecord(chainID isc.ChainID) (*ChainRecord, error) {
	return p.UpdateChainRecord(chainID, func(r *ChainRecord) bool {
		if r.Active {
			// chain was already active
			return false
		}
		r.Active = true
		return true
	})
}

func (p *ChainRecordRegistry) DeactivateChainRecord(chainID isc.ChainID) (*ChainRecord, error) {
	return p.UpdateChainRecord(chainID, func(r *ChainRecord) bool {
		if !r.Active {
			// chain was already disabled
			return false
		}
		r.Active = false
		return true
	})
}

type jsonChainRecords struct {
	ChainRecords []*ChainRecord `json:"chainRecords"`
}

func LoadChainRecordsJSONFromFile(filePath string, chainRecordsRegistry *ChainRecordRegistry) error {
	tmpChainRecords := &jsonChainRecords{}
	if err := ioutils.ReadJSONFromFile(filePath, tmpChainRecords); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to unmarshal json file: %w", err)
	}

	for i := range tmpChainRecords.ChainRecords {
		if err := chainRecordsRegistry.AddChainRecord(tmpChainRecords.ChainRecords[i]); err != nil {
			return fmt.Errorf("unable to add ChainRecord to registry: %w", err)
		}
	}

	return nil
}

func WriteChainRecordsJSONToFile(filePath string, chainRecords []*ChainRecord) error {
	if err := os.MkdirAll(path.Dir(filePath), 0o770); err != nil {
		return fmt.Errorf("unable to create folder \"%s\": %w", path.Dir(filePath), err)
	}

	if err := ioutils.WriteJSONToFile(filePath, &jsonChainRecords{ChainRecords: chainRecords}, 0o600); err != nil {
		return fmt.Errorf("unable to marshal json file: %w", err)
	}

	return nil
}
