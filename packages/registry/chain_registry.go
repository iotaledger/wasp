// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package registry provides functionality for managing chain records, consensus state,
// and distributed key shares in the system.
package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/runtime/event"
	"github.com/iotaledger/hive.go/runtime/ioutils"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/onchangemap"
	"github.com/iotaledger/wasp/packages/util"
)

// ChainRecord represents chain the node is participating in.
type ChainRecord struct {
	id          isc.ChainID
	Active      bool
	AccessNodes []*cryptolib.PublicKey
}

func NewChainRecord(chainID isc.ChainID, active bool, accessNodes []*cryptolib.PublicKey) *ChainRecord {
	return &ChainRecord{
		id:          chainID,
		Active:      active,
		AccessNodes: accessNodes,
	}
}

func (r *ChainRecord) ID() isc.ChainID {
	return r.id
}

func (r *ChainRecord) ChainID() isc.ChainID {
	return r.id
}

func (r *ChainRecord) Clone() onchangemap.Item[isc.ChainIDKey, isc.ChainID] {
	return &ChainRecord{
		id:          r.id,
		Active:      r.Active,
		AccessNodes: util.CloneSlice(r.AccessNodes),
	}
}

func (r *ChainRecord) AddAccessNode(pubKey *cryptolib.PublicKey) (modified bool) {
	if lo.ContainsBy(r.AccessNodes, func(p *cryptolib.PublicKey) bool {
		return p.Equals(pubKey)
	}) {
		// node already exists
		return false
	}
	r.AccessNodes = append(r.AccessNodes, pubKey)
	return true
}

func (r *ChainRecord) RemoveAccessNode(pubKey *cryptolib.PublicKey) (modified bool) {
	newAccessNodes := []*cryptolib.PublicKey{}
	for _, p := range r.AccessNodes {
		if p.Equals(pubKey) {
			modified = true
		} else {
			newAccessNodes = append(newAccessNodes, p)
		}
	}
	r.AccessNodes = newAccessNodes
	return modified
}

type jsonChainRecord struct {
	ChainID     string   `json:"chainID"`
	Active      bool     `json:"active"`
	AccessNodes []string `json:"accessNodes"`
}

type jsonChainRecords struct {
	ChainRecords []*ChainRecord `json:"chainRecords"`
}

func (r *ChainRecord) MarshalJSON() ([]byte, error) {
	accessNodesPubKeysHex := make([]string, 0)
	for _, accessNodePubKey := range r.AccessNodes {
		accessNodesPubKeysHex = append(accessNodesPubKeysHex, accessNodePubKey.String())
	}

	return json.Marshal(&jsonChainRecord{
		ChainID:     r.ID().String(),
		Active:      r.Active,
		AccessNodes: accessNodesPubKeysHex,
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

	address, err := cryptolib.NewAddressFromHexString(j.ChainID)
	if err != nil {
		return err
	}

	accessNodesPubKeys := make([]*cryptolib.PublicKey, len(j.AccessNodes))
	for i, accessNodePubKeyHex := range j.AccessNodes {
		accessNodePubKey, err := cryptolib.PublicKeyFromString(accessNodePubKeyHex)
		if err != nil {
			return err
		}

		accessNodesPubKeys[i] = accessNodePubKey
	}

	*r = ChainRecord{
		id:          isc.ChainIDFromAddress(address),
		Active:      j.Active,
		AccessNodes: accessNodesPubKeys,
	}
	return nil
}

// ChainRecordModifiedEvent contains the updated information of a chain record.
type ChainRecordModifiedEvent struct {
	ChainRecord *ChainRecord
}

// ChainRecordRegistryEvents contain all events of the ChainRecordRegistry.
type ChainRecordRegistryEvents struct {
	// A ChainRecordModified event is triggered, when a chain record was modified.
	ChainRecordModified *event.Event1[*ChainRecordModifiedEvent]
}

type ChainRecordRegistryImpl struct {
	onChangeMap *onchangemap.OnChangeMap[isc.ChainIDKey, isc.ChainID, *ChainRecord]

	events *ChainRecordRegistryEvents

	filePath string
}

var _ ChainRecordRegistryProvider = &ChainRecordRegistryImpl{}

// NewChainRecordRegistryImpl creates new instance of the chain registry implementation.
func NewChainRecordRegistryImpl(filePath string) (*ChainRecordRegistryImpl, error) {
	// create the target directory during initialization
	if err := util.CreateDirectoryForFilePath(filePath, 0o770); err != nil {
		return nil, err
	}

	registry := &ChainRecordRegistryImpl{
		filePath: filePath,
		events: &ChainRecordRegistryEvents{
			ChainRecordModified: event.New1[*ChainRecordModifiedEvent](),
		},
	}

	registry.onChangeMap = onchangemap.NewOnChangeMap(
		onchangemap.WithChangedCallback[isc.ChainIDKey, isc.ChainID](registry.writeChainRecordsJSON),
	)

	// load chain records on startup
	if err := registry.loadChainRecordsJSON(); err != nil {
		return nil, fmt.Errorf("unable to read chain records configuration (%s): %w", filePath, err)
	}

	registry.onChangeMap.CallbacksEnabled(true)

	return registry, nil
}

func (p *ChainRecordRegistryImpl) loadChainRecordsJSON() error {
	if p.filePath == "" {
		// do not load entries if no path is given
		return nil
	}

	tmpChainRecords := &jsonChainRecords{}
	if err := ioutils.ReadJSONFromFile(p.filePath, tmpChainRecords); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to unmarshal json file: %w", err)
	}

	for i := range tmpChainRecords.ChainRecords {
		if err := p.AddChainRecord(tmpChainRecords.ChainRecords[i]); err != nil {
			return fmt.Errorf("unable to add ChainRecord to registry: %w", err)
		}
	}

	return nil
}

func (p *ChainRecordRegistryImpl) writeChainRecordsJSON(chainRecords []*ChainRecord) error {
	if p.filePath == "" {
		// do not store entries if no path is given
		return nil
	}

	if err := util.CreateDirectoryForFilePath(p.filePath, 0o770); err != nil {
		return err
	}

	if err := ioutils.WriteJSONToFile(p.filePath, &jsonChainRecords{ChainRecords: chainRecords}, 0o600); err != nil {
		return fmt.Errorf("unable to marshal json file: %w", err)
	}

	return nil
}

func (p *ChainRecordRegistryImpl) Events() *ChainRecordRegistryEvents {
	return p.events
}

func (p *ChainRecordRegistryImpl) ChainRecord(chainID isc.ChainID) (*ChainRecord, error) {
	chainRecord, err := p.onChangeMap.Get(chainID)
	if err != nil {
		// chain record doesn't exist
		return nil, nil
	}

	return chainRecord, nil
}

func (p *ChainRecordRegistryImpl) ChainRecords() ([]*ChainRecord, error) {
	return lo.Values(p.onChangeMap.All()), nil
}

func (p *ChainRecordRegistryImpl) ForEachActiveChainRecord(consumer func(*ChainRecord) bool) error {
	chainRecords, err := p.ChainRecords()
	if err != nil {
		return err
	}

	for _, chainRecord := range chainRecords {
		if !chainRecord.Active {
			continue
		}

		if !consumer(chainRecord) {
			return nil
		}
	}

	return nil
}

func (p *ChainRecordRegistryImpl) AddChainRecord(chainRecord *ChainRecord) error {
	chains := len(p.onChangeMap.All())
	// Only allow a single chain
	if chains != 0 {
		return fmt.Errorf("too many active chain records")
	}

	return p.onChangeMap.Add(chainRecord)
}

func (p *ChainRecordRegistryImpl) DeleteChainRecord(chainID isc.ChainID) error {
	return p.onChangeMap.Delete(chainID)
}

// UpdateChainRecord modifies a ChainRecord in the Registry.
func (p *ChainRecordRegistryImpl) UpdateChainRecord(chainID isc.ChainID, callback func(*ChainRecord) bool) (*ChainRecord, error) {
	var modified bool
	callbackHook := func(chainRecord *ChainRecord) bool {
		modified = callback(chainRecord)
		return modified
	}

	chainRecord, err := p.onChangeMap.Modify(chainID, callbackHook)
	if err != nil {
		return chainRecord, err
	}

	if modified {
		p.events.ChainRecordModified.Trigger(&ChainRecordModifiedEvent{
			ChainRecord: chainRecord,
		})
	}

	return chainRecord, nil
}

func (p *ChainRecordRegistryImpl) ActivateChainRecord(chainID isc.ChainID) (*ChainRecord, error) {
	return p.UpdateChainRecord(chainID, func(r *ChainRecord) bool {
		if r.Active {
			// chain was already active
			return false
		}
		r.Active = true
		return true
	})
}

func (p *ChainRecordRegistryImpl) DeactivateChainRecord(chainID isc.ChainID) (*ChainRecord, error) {
	return p.UpdateChainRecord(chainID, func(r *ChainRecord) bool {
		if !r.Active {
			// chain was already disabled
			return false
		}
		r.Active = false
		return true
	})
}
