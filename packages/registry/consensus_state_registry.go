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
	"strings"

	"github.com/iotaledger/hive.go/runtime/ioutils"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/chain/cmtlog"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/onchangemap"
	"github.com/iotaledger/wasp/v2/packages/util"
)

const comparableChainCommitteeIDKeyLength = isc.ChainIDLength + iotago.AddressLen

type comparableChainCommitteeIDKey [comparableChainCommitteeIDKeyLength]byte

type comparableChainCommitteeID struct {
	key     comparableChainCommitteeIDKey
	chainID isc.ChainID
	address *cryptolib.Address
}

func newComparableChainCommitteeID(chainID isc.ChainID, address *cryptolib.Address) *comparableChainCommitteeID {
	addressBytes := address.Bytes()

	key := comparableChainCommitteeIDKey{}
	copy(key[:isc.ChainIDLength], chainID[:])
	copy(key[isc.ChainIDLength:], addressBytes)

	return &comparableChainCommitteeID{
		key:     key,
		chainID: chainID,
		address: address,
	}
}

func (c *comparableChainCommitteeID) Key() comparableChainCommitteeIDKey {
	return c.key
}

func (c *comparableChainCommitteeID) String() string {
	return fmt.Sprintf("%s-%s", c.chainID, c.address.String())
}

type consensusState struct {
	identifier *comparableChainCommitteeID

	LogIndex cmtlog.LogIndex
}

func (c *consensusState) ID() *comparableChainCommitteeID {
	return c.identifier
}

func (c *consensusState) Clone() onchangemap.Item[comparableChainCommitteeIDKey, *comparableChainCommitteeID] {
	return &consensusState{
		identifier: c.identifier,
		LogIndex:   c.LogIndex,
	}
}

func (c *consensusState) ChainID() isc.ChainID {
	return c.identifier.chainID
}

func (c *consensusState) Address() *cryptolib.Address {
	return c.identifier.address
}

type jsonConsensusState struct {
	ChainID          string `json:"chainId"`
	CommitteeAddress string `json:"committeeAddress"`
	LogIndex         uint32 `json:"logIndex"`
}

func (c *consensusState) MarshalJSON() ([]byte, error) {
	chainID := c.identifier.chainID.AsAddress().String()

	return json.Marshal(&jsonConsensusState{
		ChainID:          chainID,
		CommitteeAddress: c.identifier.address.String(),
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

	committeeAddress, err := cryptolib.NewAddressFromHexString(j.CommitteeAddress)
	if err != nil {
		return err
	}

	*c = consensusState{
		identifier: newComparableChainCommitteeID(chainID, committeeAddress),
		LogIndex:   cmtlog.LogIndex(j.LogIndex),
	}

	return nil
}

type ConsensusStateRegistry struct {
	onChangeMap *onchangemap.OnChangeMap[comparableChainCommitteeIDKey, *comparableChainCommitteeID, *consensusState]

	folderPath string
}

var _ cmtlog.ConsensusStateRegistry = &ConsensusStateRegistry{}

// NewConsensusStateRegistry creates new instance of the consensus state registry implementation.
func NewConsensusStateRegistry(folderPath string) (*ConsensusStateRegistry, error) {
	// create the target directory during initialization
	if err := ioutils.CreateDirectory(folderPath, 0o770); err != nil {
		return nil, err
	}

	registry := &ConsensusStateRegistry{
		folderPath: folderPath,
	}

	registry.onChangeMap = onchangemap.NewOnChangeMap(
		onchangemap.WithItemAddedCallback[comparableChainCommitteeIDKey, *comparableChainCommitteeID](registry.writeConsensusStateJSON),
		onchangemap.WithItemModifiedCallback[comparableChainCommitteeIDKey, *comparableChainCommitteeID](registry.writeConsensusStateJSON),
		onchangemap.WithItemDeletedCallback[comparableChainCommitteeIDKey, *comparableChainCommitteeID](registry.deleteConsensusStateJSON),
	)

	// load chain records on startup
	if err := registry.loadConsensusStateJSONsFromFolder(); err != nil {
		return nil, fmt.Errorf("unable to read chain records configuration (%s): %w", folderPath, err)
	}

	registry.onChangeMap.CallbacksEnabled(true)

	return registry, nil
}

func (p *ConsensusStateRegistry) loadConsensusStateJSONsFromFolder() error {
	if p.folderPath == "" {
		// do not load entries if no path is given
		return nil
	}

	checkSubFolderFiles := func(chainID isc.ChainID, subFolderPath string, subFolderFile fs.DirEntry) error {
		if subFolderFile.IsDir() {
			// ignore folders
			return nil
		}

		committeeAddressHex := strings.ReplaceAll(subFolderFile.Name(), ".json", "")
		committeeAddress, err := cryptolib.NewAddressFromHexString(committeeAddressHex)
		if err != nil {
			return fmt.Errorf("unable to parse committee hex address (%s), error: %w", committeeAddressHex, err)
		}

		consensusStateFilePath := path.Join(subFolderPath, subFolderFile.Name())

		state := &consensusState{
			identifier: newComparableChainCommitteeID(chainID, committeeAddress),
			LogIndex:   0,
		}

		if err := ioutils.ReadJSONFromFile(consensusStateFilePath, state); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("unable to unmarshal json file (%s): %w", consensusStateFilePath, err)
		}

		if state.identifier.chainID != chainID {
			return errors.New("unable to add consensus state to registry: chainID in the file not equal to chainID in folder name")
		}

		if !state.identifier.address.Equals(committeeAddress) {
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

		chainAddress, err := cryptolib.NewAddressFromHexString(rootFolderFile.Name())
		if err != nil {
			return fmt.Errorf("unable to parse consensus state hex address (%s), error: %w", rootFolderFile.Name(), err)
		}

		chainID := isc.ChainIDFromAddress(chainAddress)

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
	chainAddressHex := state.ChainID().AsAddress().String()
	committeeAddressHex := state.Address().String()

	return path.Join(p.folderPath, chainAddressHex, fmt.Sprintf("%s.json", committeeAddressHex))
}

func (p *ConsensusStateRegistry) writeConsensusStateJSON(state *consensusState) error {
	if p.folderPath == "" {
		// do not store entries if no path is given
		return nil
	}

	filePath := p.getConsensusStateFilePath(state)

	if err := util.CreateDirectoryForFilePath(filePath, 0o770); err != nil {
		return err
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

// Get retrieves the consensus state for the given key. Can return cmtLog.ErrCmtLogStateNotFound.
func (p *ConsensusStateRegistry) Get(chainID isc.ChainID, committeeAddress *cryptolib.Address) (*cmtlog.State, error) {
	state, err := p.onChangeMap.Get(newComparableChainCommitteeID(chainID, committeeAddress))
	if err != nil {
		return nil, cmtlog.ErrCmtLogStateNotFound
	}

	return &cmtlog.State{
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

func (p *ConsensusStateRegistry) Set(chainID isc.ChainID, committeeAddress *cryptolib.Address, state *cmtlog.State) error {
	return p.add(&consensusState{
		identifier: newComparableChainCommitteeID(chainID, committeeAddress),
		LogIndex:   state.LogIndex,
	})
}
