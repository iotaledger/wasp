// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscmagic

import (
	_ "embed"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// If you change any of the .sol files, you must recompile.  You will need
// the `solc` binary installed in your system. Then, simply run `go generate`
// in this directory.

//go:generate sh -c "solc --abi --overwrite @iscmagic=`realpath .` ISCSandbox.sol -o ."
//go:generate sh -c "solc --abi --overwrite @iscmagic=`realpath .` ISCUtil.sol -o ."
//go:generate sh -c "solc --abi --overwrite @iscmagic=`realpath .` ISCAccounts.sol -o ."
//go:generate sh -c "solc --abi --overwrite @iscmagic=`realpath .` ISCPrivileged.sol -o ."
var (
	//go:embed ISCSandbox.abi
	SandboxABI string

	//go:embed ISCUtil.abi
	UtilABI string

	//go:embed ISCAccounts.abi
	AccountsABI string

	//go:embed ISCPrivileged.abi
	PrivilegedABI string
)

//go:generate sh -c "solc --abi --bin-runtime --overwrite @iscmagic=`realpath .` ERC20BaseTokens.sol -o ."
var (
	//go:embed ERC20BaseTokens.abi
	ERC20BaseTokensABI string

	//go:embed ERC20BaseTokens.bin-runtime
	erc20BaseRuntimeBytecodeHex    string
	ERC20BaseTokensRuntimeBytecode = common.FromHex(strings.TrimSpace(erc20BaseRuntimeBytecodeHex))

	ERC20BaseTokensAddress = packMagicAddress(addressKindERC20BaseTokens, nil)
)

//go:generate sh -c "solc --abi --bin-runtime --storage-layout --overwrite @iscmagic=`realpath .` ERC20NativeTokens.sol -o ."
var (
	//go:embed ERC20NativeTokens.abi
	ERC20NativeTokensABI string

	//go:embed ERC20NativeTokens.bin-runtime
	erc20NativeTokensRuntimeBytecodeHex string
	ERC20NativeTokensRuntimeBytecode    = common.FromHex(strings.TrimSpace(erc20NativeTokensRuntimeBytecodeHex))
)

//go:generate sh -c "solc --abi --bin-runtime --storage-layout --overwrite @iscmagic=`realpath .` ERC20ExternalNativeTokens.sol -o ."
var (
	//go:embed ERC20ExternalNativeTokens.abi
	ERC20ExternalNativeTokensABI string

	//go:embed ERC20ExternalNativeTokens.bin-runtime
	erc20ExternalNativeTokensRuntimeBytecodeHex string
	ERC20ExternalNativeTokensRuntimeBytecode    = common.FromHex(strings.TrimSpace(erc20ExternalNativeTokensRuntimeBytecodeHex))
)

//go:generate sh -c "solc --abi --bin-runtime --overwrite @iscmagic=`realpath .` ERC721NFTs.sol -o ."
var (
	//go:embed ERC721NFTs.abi
	ERC721NFTsABI string
	//go:embed ERC721NFTs.bin-runtime
	erc721NFTsBytecodeHex     string
	ERC721NFTsRuntimeBytecode = common.FromHex(strings.TrimSpace(erc721NFTsBytecodeHex))

	ERC721NFTsAddress = packMagicAddress(addressKindERC721NFTs, nil)
)

//go:generate sh -c "solc --abi --storage-layout --bin-runtime --overwrite @iscmagic=`realpath .` ERC721NFTCollection.sol -o ."
var (
	//go:embed ERC721NFTCollection.abi
	ERC721NFTCollectionABI string
	//go:embed ERC721NFTCollection.bin-runtime
	erc721NFTCollectionBytecodeHex     string
	ERC721NFTCollectionRuntimeBytecode = common.FromHex(strings.TrimSpace(erc721NFTCollectionBytecodeHex))
)
