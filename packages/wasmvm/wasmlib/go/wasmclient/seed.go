// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"github.com/iotaledger/goshimmer/client/wallet/packages/seed"
	cryptolib "github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/mr-tron/base58"
)

func ChainIsValid(chainID string) bool {
	bytes, err := base58.Decode(chainID)
	return err == nil && len(bytes) == 33
}

func SeedIsValid(mySeed string) bool {
	seedBytes, err := base58.Decode(mySeed)
	return err == nil && len(seedBytes) == 32
}

func SeedToAddress(mySeed string, index uint64) Address {
	seedBytes, err := base58.Decode(mySeed)
	if err != nil {
		panic(err)
	}
	address := seed.NewSeed(seedBytes).Address(index)
	return Address(base58.Encode(address.Address().Bytes()))
}

func SeedToAgentID(mySeed string, index uint64) AgentID {
	seedBytes, err := base58.Decode(mySeed)
	if err != nil {
		panic(err)
	}
	address := seed.NewSeed(seedBytes).Address(index)
	agentID := iscp.NewAgentID(address.Address(), 0)
	return AgentID(base58.Encode(agentID.Bytes()))
}

func SeedToKeyPair(mySeed string, index uint64) *cryptolib.KeyPair {
	seedBytes, err := base58.Decode(mySeed)
	if err != nil {
		panic(err)
	}
	return seed.NewSeed(seedBytes).KeyPair(index)
}
