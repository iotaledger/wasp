// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/iotaledger/hive.go/core/generics/onchangemap"
	"github.com/iotaledger/hive.go/core/ioutils"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util"
)

type DKSharesRegistry struct {
	onChangeMap *onchangemap.OnChangeMap[string, *util.ComparableAddress, tcrypto.DKShare]

	folderPath string
}

var _ DKShareRegistryProvider = &DKSharesRegistry{}

// NewDKSharesRegistry creates new instance of the DKShare registry implementation.
func NewDKSharesRegistry(folderPath string, nodePrivKey *cryptolib.PrivateKey) (*DKSharesRegistry, error) {
	registry := &DKSharesRegistry{
		folderPath: folderPath,
	}

	registry.onChangeMap = onchangemap.NewOnChangeMap(
		onchangemap.WithItemAddedCallback[string, *util.ComparableAddress](registry.writeDKShareJSONToFolder),
		onchangemap.WithItemModifiedCallback[string, *util.ComparableAddress](registry.writeDKShareJSONToFolder),
		onchangemap.WithItemDeletedCallback[string, *util.ComparableAddress](registry.deleteDKShareJSON),
	)

	// load DKShares on startup
	if err := registry.loadDKSharesJSONFromFolder(nodePrivKey); err != nil {
		return nil, fmt.Errorf("unable to read DKShares configuration (%s): %s", folderPath, err)
	}

	registry.onChangeMap.CallbacksEnabled(true)

	return registry, nil
}

func (p *DKSharesRegistry) loadDKSharesJSONFromFolder(nodePrivKey *cryptolib.PrivateKey) error {
	if p.folderPath == "" {
		// do not load entries if no path is given
		return nil
	}

	// regex example: atoi1qqqrqtn44e0563utwau9aaygt824qznjkhvr6836eratglg3cp2n6ydplqx.json
	filesRegex := regexp.MustCompile(`([a-z]{1,4}1[a-z0-9]{59}).json`)

	files, err := os.ReadDir(p.folderPath)
	if err != nil {
		if os.IsNotExist(err) {
			// if the folder doesn't exist, there are no entries yet.
			return nil
		}
		return fmt.Errorf("unable to read dkShares directory (%s), error: %w", p.folderPath, err)
	}

	// loop over all matching files
	for _, file := range files {
		if file.IsDir() {
			// ignore folders
			return nil
		}

		if !filesRegex.MatchString(file.Name()) {
			// ignore unknown files
			return nil
		}

		sharedAddressBech32 := filesRegex.FindStringSubmatch(file.Name())[1]
		_, sharedAddress, err := iotago.ParseBech32(sharedAddressBech32)
		if err != nil {
			return fmt.Errorf("unable to parse shared bech32 address (%s), error: %w", sharedAddressBech32, err)
		}

		dkShareFilePath := path.Join(p.folderPath, file.Name())
		dkShare := tcrypto.NewEmptyDKShare(nodePrivKey, tcrypto.DefaultEd25519Suite(), tcrypto.DefaultBLSSuite())
		if err := ioutils.ReadJSONFromFile(dkShareFilePath, dkShare); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("unable to unmarshal json file (%s): %w", dkShareFilePath, err)
		}

		if !dkShare.GetAddress().Equal(sharedAddress) {
			return errors.New("unable to add DKShare to registry: sharedAddress in the file not equal to sharedAddress in folder name")
		}

		if err := p.SaveDKShare(dkShare); err != nil {
			return fmt.Errorf("unable to add DKShare to registry: %w", err)
		}
	}

	return nil
}

func (p *DKSharesRegistry) getDKShareFilePath(dkShare tcrypto.DKShare) string {
	sharedAddressBech32 := dkShare.GetAddress().Bech32(parameters.L1().Protocol.Bech32HRP)

	return path.Join(p.folderPath, fmt.Sprintf("%s.json", sharedAddressBech32))
}

func (p *DKSharesRegistry) writeDKShareJSONToFolder(dkShare tcrypto.DKShare) error {
	if p.folderPath == "" {
		// do not store entries if no path is given
		return nil
	}

	filePath := p.getDKShareFilePath(dkShare)

	if err := os.MkdirAll(path.Dir(filePath), 0o770); err != nil {
		return fmt.Errorf("unable to create folder \"%s\": %w", path.Dir(filePath), err)
	}

	if err := ioutils.WriteJSONToFile(filePath, dkShare, 0o600); err != nil {
		return fmt.Errorf("unable to marshal json file: %w", err)
	}

	return nil
}

func (p *DKSharesRegistry) deleteDKShareJSON(dkShare tcrypto.DKShare) error {
	if p.folderPath == "" {
		// do not delete entries if no path is given
		return nil
	}

	filePath := p.getDKShareFilePath(dkShare)

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

func (p *DKSharesRegistry) SaveDKShare(dkShare tcrypto.DKShare) error {
	return p.onChangeMap.Add(dkShare)
}

func (p *DKSharesRegistry) LoadDKShare(sharedAddress iotago.Address) (tcrypto.DKShare, error) {
	dkShare, err := p.onChangeMap.Get(util.NewComparableAddress(sharedAddress))
	if err != nil {
		return dkShare, tcrypto.ErrDKShareNotFound
	}
	return dkShare, nil
}
