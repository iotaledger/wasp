// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscmagic

import (
	"math/big"
)

// LegacyNativeTokenID matches the struct definition in ISCTypesLegacy.sol
type LegacyNativeTokenID struct {
	Data []byte
}

// LegacyNativeToken matches the struct definition in ISCTypesLegacy.sol
type LegacyNativeToken struct {
	ID     LegacyNativeTokenID
	Amount *big.Int
}

// LegacyL1Address matches the struct definition in ISCTypesLegacy.sol
type LegacyL1Address struct {
	Data []byte
}

// LegacyNFTID matches the type definition in ISCTypesLegacy.sol
type LegacyNFTID [32]byte

// LegacyISCAssets matches the struct definition in ISCTypesLegacy.sol
type LegacyISCAssets struct {
	BaseTokens   uint64
	NativeTokens []LegacyNativeToken
	Nfts         []LegacyNFTID
}

type LegacyISCSendMetadata struct {
	TargetContract uint32
	Entrypoint     uint32
	Params         ISCDict
	Allowance      LegacyISCAssets
	GasBudget      uint64
}

type LegacyISCExpiration struct {
	Time          int64
	ReturnAddress LegacyL1Address
}

type LegacyISCSendOptions struct {
	Timelock   int64
	Expiration LegacyISCExpiration
}
