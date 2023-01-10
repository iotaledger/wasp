// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"

	"github.com/iotaledger/hive.go/core/generics/onchangemap"
	"github.com/iotaledger/hive.go/core/ioutils"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cmtLog"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
)

const comparableConsensusIDKeyLength = isc.ChainIDLength + iotago.Ed25519AddressBytesLength

type comparableConsensusIDKey [comparableConsensusIDKeyLength]byte

type comparableConsensusID struct {
	key     comparableConsensusIDKey
	chainID isc.ChainID
	address iotago.Address
}

func newComparableConsensusID(chainID isc.ChainID, address iotago.Address) *comparableConsensusID {
	addressBytes := isc.BytesFromAddress(address)

	key := comparableConsensusIDKey{}
	copy(key[:isc.ChainIDLength], chainID[:])
	copy(key[isc.ChainIDLength:], addressBytes)

	return &comparableConsensusID{
		key:     key,
		chainID: chainID,
		address: address,
	}
}

func (c *comparableConsensusID) Key() comparableConsensusIDKey {
	return c.key
}

func (c *comparableConsensusID) String() string {
	return fmt.Sprintf("%s-%s", c.chainID, c.address.Bech32(parameters.L1().Protocol.Bech32HRP))
}

type consensusState struct {
	identifier *comparableConsensusID

	LogIndex cmtLog.LogIndex
}

func (c *consensusState) ID() *comparableConsensusID {
	return c.identifier
}

func (c *consensusState) Clone() onchangemap.Item[comparableConsensusIDKey, *comparableConsensusID] {
	return &consensusState{
		identifier: c.identifier,
		LogIndex:   c.LogIndex,
	}
}

func (c *consensusState) ChainID() isc.ChainID {
	return c.identifier.chainID
}

func (c *consensusState) Address() iotago.Address {
	return c.identifier.address
}

type jsonConsensusState struct {
	ChainID          string           `json:"chainId"`
	CommitteeAddress *json.RawMessage `json:"committeeAddress"`
	LogIndex         uint32           `json:"logIndex"`
}

func (c *consensusState) MarshalJSON() ([]byte, error) {
	chainIDBech32 := c.identifier.chainID.AsAddress().Bech32(parameters.L1().Protocol.Bech32HRP)

	jAddressRaw, err := iotago.AddressToJSONRawMsg(c.identifier.address)
	if err != nil {
		return nil, err
	}

	return json.Marshal(&jsonConsensusState{
		ChainID:          chainIDBech32,
		CommitteeAddress: jAddressRaw,
		LogIndex:         c.LogIndex.AsUint32(),
	})
}

func (c *consensusState) UnmarshalJSON(bytes []byte) error {
	j := &jsonConsensusState{}
	if err := json.Unmarshal(bytes, j); err != nil {
		return err
	}

	chainID, err := isc.ChainIDFromString(j.ChainID)
	if err != nil {
		return err
	}

	committeeAddress, err := iotago.AddressFromJSONRawMsg(j.CommitteeAddress)
	if err != nil {
		return err
	}

	*c = consensusState{
		identifier: newComparableConsensusID(chainID, committeeAddress),
		LogIndex:   cmtLog.LogIndex(j.LogIndex),
	}

	return nil
}

type ConsensusStateRegistry struct {
	onChangeMap *onchangemap.OnChangeMap[comparableConsensusIDKey, *comparableConsensusID, *consensusState]

	folderPath string
}

var _ cmtLog.ConsensusStateRegistry = &ConsensusStateRegistry{}

// NewConsensusStateRegistry creates new instance of the consensus state registry implementation.
func NewConsensusStateRegistry(folderPath string) (*ConsensusStateRegistry, error) {
	registry := &ConsensusStateRegistry{
		folderPath: folderPath,
	}

	registry.onChangeMap = onchangemap.NewOnChangeMap(
		onchangemap.WithItemAddedCallback[comparableConsensusIDKey, *comparableConsensusID](registry.writeConsensusStateJSON),
		onchangemap.WithItemModifiedCallback[comparableConsensusIDKey, *comparableConsensusID](registry.writeConsensusStateJSON),
		onchangemap.WithItemDeletedCallback[comparableConsensusIDKey, *comparableConsensusID](registry.deleteConsensusStateJSON),
	)

	// load chain records on startup
	if err := registry.loadConsensusStateJSONsFromFolder(); err != nil {
		return nil, fmt.Errorf("unable to read chain records configuration (%s): %w", folderPath, err)
	}

	registry.onChangeMap.CallbacksEnabled(true)

	return registry, nil
}

//nolint:gocyclo,funlen
func (p *ConsensusStateRegistry) loadConsensusStateJSONsFromFolder() error {
	if p.folderPath == "" {
		// do not load entries if no path is given
		return nil
	}

	// regex example: atoi1qqqrqtn44e0563utwau9aaygt824qznjkhvr6836eratglg3cp2n6ydplqx
	foldersRegex := regexp.MustCompile(`[a-z]{1,4}1[a-z0-9]{59}`)

	// regex example: atoi1qqqrqtn44e0563utwau9aaygt824qznjkhvr6836eratglg3cp2n6ydplqx.json
	filesRegex := regexp.MustCompile(`([a-z]{1,4}1[a-z0-9]{59}).json`)

	checkSubFolderFiles := func(chainID isc.ChainID, subFolderPath string, subFolderFile fs.DirEntry) error {
		if subFolderFile.IsDir() {
			// ignore folders
			return nil
		}

		if !filesRegex.MatchString(subFolderFile.Name()) {
			// ignore unknown files
			return nil
		}

		committeeAddressBech32 := filesRegex.FindStringSubmatch(subFolderFile.Name())[1]
		_, committeeAddress, err := iotago.ParseBech32(committeeAddressBech32)
		if err != nil {
			return fmt.Errorf("unable to parse committee bech32 address (%s), error: %w", committeeAddressBech32, err)
		}

		consensusStateFilePath := path.Join(subFolderPath, subFolderFile.Name())

		state := &consensusState{
			identifier: newComparableConsensusID(chainID, committeeAddress),
			LogIndex:   0,
		}

		if err := ioutils.ReadJSONFromFile(consensusStateFilePath, state); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("unable to unmarshal json file (%s): %w", consensusStateFilePath, err)
		}

		if state.identifier.chainID != chainID {
			return errors.New("unable to add consensus state to registry: chainID in the file not equal to chainID in folder name")
		}

		if !state.identifier.address.Equal(committeeAddress) {
			return errors.New("unable to add consensus state to registry: committeeAddress in the file not equal to committeeAddress in folder name")
		}

		if err := p.add(state); err != nil {
			return fmt.Errorf("unable to add consensus state to registry: %w", err)
		}

		return nil
	}

	rootFolderFiles, err := os.ReadDir(p.folderPath)
	if err != nil {
		if os.IsNotExist(err) {
			// if the folder doesn't exist, there are no entries yet.
			return nil
		}
		return fmt.Errorf("unable to read consensus state directory (%s), error: %w", p.folderPath, err)
	}

	// loop over all matching folders
	for _, rootFolderFile := range rootFolderFiles {
		if !rootFolderFile.IsDir() {
			// ignore files
			continue
		}

		if !foldersRegex.MatchString(rootFolderFile.Name()) {
			// ignore unknown folders
			continue
		}

		chainAddressBech32 := foldersRegex.FindStringSubmatch(rootFolderFile.Name())[0]
		_, chainAddress, err := iotago.ParseBech32(chainAddressBech32)
		if err != nil {
			return fmt.Errorf("unable to parse consensus state bech32 address (%s), error: %w", chainAddressBech32, err)
		}

		if chainAddress.Type() != iotago.AddressAlias {
			return fmt.Errorf("chainID bech32 address is not an alias address (%s), error: %w", chainAddressBech32, err)
		}

		chainID := isc.ChainIDFromAddress(chainAddress.(*iotago.AliasAddress))

		subFolderPath := path.Join(p.folderPath, rootFolderFile.Name())

		subFolderFiles, err := os.ReadDir(subFolderPath)
		if err != nil {
			return fmt.Errorf("unable to read consensus state directory (%s), error: %w", subFolderPath, err)
		}

		// loop over all matching files
		for _, subFolderFile := range subFolderFiles {
			if err := checkSubFolderFiles(chainID, subFolderPath, subFolderFile); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *ConsensusStateRegistry) getConsensusStateFilePath(state *consensusState) string {
	chainAddressBech32 := state.ChainID().AsAddress().Bech32(parameters.L1().Protocol.Bech32HRP)
	committeeAddressBech32 := state.Address().Bech32(parameters.L1().Protocol.Bech32HRP)

	return path.Join(p.folderPath, chainAddressBech32, fmt.Sprintf("%s.json", committeeAddressBech32))
}

func (p *ConsensusStateRegistry) writeConsensusStateJSON(state *consensusState) error {
	if p.folderPath == "" {
		// do not store entries if no path is given
		return nil
	}

	filePath := p.getConsensusStateFilePath(state)

	if err := os.MkdirAll(path.Dir(filePath), 0o770); err != nil {
		return fmt.Errorf("unable to create folder \"%s\": %w", path.Dir(filePath), err)
	}

	if err := ioutils.WriteJSONToFile(filePath, state, 0o600); err != nil {
		return fmt.Errorf("unable to marshal json file: %w", err)
	}

	return nil
}

func (p *ConsensusStateRegistry) deleteConsensusStateJSON(state *consensusState) error {
	if p.folderPath == "" {
		// do not delete entries if no path is given
		return nil
	}

	filePath := p.getConsensusStateFilePath(state)

	exists, isDir, err := ioutils.PathExists(filePath)
	if err != nil {
		return fmt.Errorf("delete consensus state file failed (%s): %w", filePath, err)
	}
	if !exists {
		// files doesn't exist
		return nil
	}
	if isDir {
		return fmt.Errorf("delete consensus state file failed: given path is a directory instead of a file %s", filePath)
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("delete consensus state file failed (%s): %w", filePath, err)
	}

	return nil
}

// Can return cmtLog.ErrCmtLogStateNotFound.
func (p *ConsensusStateRegistry) Get(chainID isc.ChainID, committeeAddress iotago.Address) (*cmtLog.State, error) {
	state, err := p.onChangeMap.Get(newComparableConsensusID(chainID, committeeAddress))
	if err != nil {
		return nil, cmtLog.ErrCmtLogStateNotFound
	}

	return &cmtLog.State{
		LogIndex: state.LogIndex,
	}, nil
}

func (p *ConsensusStateRegistry) add(state *consensusState) error {
	if err := p.onChangeMap.Add(state); err != nil {
		// already exists, modify the existing
		if _, err := p.onChangeMap.Modify(state.ID(), func(entry *consensusState) bool {
			if entry.LogIndex == state.LogIndex {
				return false
			}

			*entry = *state
			return true
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *ConsensusStateRegistry) Set(chainID isc.ChainID, committeeAddress iotago.Address, state *cmtLog.State) error {
	return p.add(&consensusState{
		identifier: newComparableConsensusID(chainID, committeeAddress),
		LogIndex:   state.LogIndex,
	})
}
