// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/iotaledger/hive.go/runtime/ioutils"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/onchangemap"
	"github.com/iotaledger/wasp/v2/packages/tcrypto"
	"github.com/iotaledger/wasp/v2/packages/util"
)

type DistKeyPartsRegistry struct {
	onChangeMap *onchangemap.OnChangeMap[cryptolib.AddressKey, *util.ComparableAddress, tcrypto.DistibutedKeyPart]

	folderPath string
}

var _ DistKeyPartRegistryProvider = &DistKeyPartsRegistry{}

// NewDistKeyPartsRegistry creates new instance of the DistKeyPart registry implementation.
func NewDistKeyPartsRegistry(folderPath string, nodePrivKey *cryptolib.PrivateKey) (*DistKeyPartsRegistry, error) {
	// create the target directory during initialization
	if err := ioutils.CreateDirectory(folderPath, 0o770); err != nil {
		return nil, err
	}

	registry := &DistKeyPartsRegistry{
		folderPath: folderPath,
	}

	registry.onChangeMap = onchangemap.NewOnChangeMap(
		onchangemap.WithItemAddedCallback[cryptolib.AddressKey, *util.ComparableAddress](registry.writeDistKeyPartJSONToFolder),
		onchangemap.WithItemModifiedCallback[cryptolib.AddressKey, *util.ComparableAddress](registry.writeDistKeyPartJSONToFolder),
		onchangemap.WithItemDeletedCallback[cryptolib.AddressKey, *util.ComparableAddress](registry.deleteDistKeyPartJSON),
	)

	// load DistKeyParts on startup
	if err := registry.loadDistKeyPartsJSONFromFolder(nodePrivKey); err != nil {
		return nil, fmt.Errorf("unable to read DistKeyParts configuration (%s): %w", folderPath, err)
	}

	registry.onChangeMap.CallbacksEnabled(true)

	return registry, nil
}

func (p *DistKeyPartsRegistry) loadDistKeyPartsJSONFromFolder(nodePrivKey *cryptolib.PrivateKey) error {
	if p.folderPath == "" {
		// do not load entries if no path is given
		return nil
	}

	files, err := os.ReadDir(p.folderPath)
	if err != nil {
		if os.IsNotExist(err) {
			// if the folder doesn't exist, there are no entries yet.
			return nil
		}
		return fmt.Errorf("unable to read distKeyParts directory (%s), error: %w", p.folderPath, err)
	}

	// loop over all matching files
	for _, file := range files {
		if file.IsDir() {
			// ignore folders
			return nil
		}

		if !bytes.HasSuffix([]byte(file.Name()), []byte(".json")) {
			// ignore unknown files
			return nil
		}

		sharedAddressHex := strings.ReplaceAll(file.Name(), ".json", "")
		sharedAddress, err := cryptolib.NewAddressFromHexString(sharedAddressHex)
		if err != nil {
			return fmt.Errorf("unable to parse shared hex address (%s), error: %w", sharedAddressHex, err)
		}

		distKeyPartFilePath := path.Join(p.folderPath, file.Name())
		distKeyPart := tcrypto.NewEmptyDistKeyPart(nodePrivKey, tcrypto.DefaultEd25519Suite(), tcrypto.DefaultBLSSuite())
		if err := ioutils.ReadJSONFromFile(distKeyPartFilePath, distKeyPart); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("unable to unmarshal json file (%s): %w", distKeyPartFilePath, err)
		}

		if !distKeyPart.GetAddress().Equals(sharedAddress) {
			return errors.New("unable to add DistKeyPart to registry: sharedAddress in the file not equal to sharedAddress in folder name")
		}

		if err := p.SaveDistKeyPart(distKeyPart); err != nil {
			return fmt.Errorf("unable to add DistKeyPart to registry: %w", err)
		}
	}

	return nil
}

func (p *DistKeyPartsRegistry) getDistKeyPartFilePath(distKeyPart tcrypto.DistibutedKeyPart) string {
	sharedAddressHex := distKeyPart.GetAddress().String()

	return path.Join(p.folderPath, fmt.Sprintf("%s.json", sharedAddressHex))
}

func (p *DistKeyPartsRegistry) writeDistKeyPartJSONToFolder(distKeyPart tcrypto.DistibutedKeyPart) error {
	if p.folderPath == "" {
		// do not store entries if no path is given
		return nil
	}

	filePath := p.getDistKeyPartFilePath(distKeyPart)
	if err := util.CreateDirectoryForFilePath(filePath, 0o770); err != nil {
		return err
	}

	if err := ioutils.WriteJSONToFile(filePath, distKeyPart, 0o600); err != nil {
		return fmt.Errorf("unable to marshal json file: %w", err)
	}

	return nil
}

func (p *DistKeyPartsRegistry) deleteDistKeyPartJSON(distKeyPart tcrypto.DistibutedKeyPart) error {
	if p.folderPath == "" {
		// do not delete entries if no path is given
		return nil
	}

	filePath := p.getDistKeyPartFilePath(distKeyPart)

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

func (p *DistKeyPartsRegistry) SaveDistKeyPart(distKeyPart tcrypto.DistibutedKeyPart) error {
	return p.onChangeMap.Add(distKeyPart)
}

func (p *DistKeyPartsRegistry) LoadDistKeyPart(sharedAddress *cryptolib.Address) (tcrypto.DistibutedKeyPart, error) {
	distKeyPart, err := p.onChangeMap.Get(util.NewComparableAddress(sharedAddress))
	if err != nil {
		return distKeyPart, tcrypto.ErrDistKeyPartNotFound
	}
	return distKeyPart, nil
}
