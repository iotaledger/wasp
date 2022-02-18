// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/utxodb"
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
	db := utxodb.New(utxodb.DefaultInitParams(seedBytes))
	_, address := db.NewKeyPairByIndex(index)
	buf := iscp.BytesFromAddress(address)
	return Address(base58.Encode(buf))
}

func SeedToAgentID(mySeed string, index uint64) AgentID {
	seedBytes, err := base58.Decode(mySeed)
	if err != nil {
		panic(err)
	}
	db := utxodb.New(utxodb.DefaultInitParams(seedBytes))
	_, address := db.NewKeyPairByIndex(index)
	agentID := iscp.NewAgentID(address, 0)
	return AgentID(base58.Encode(agentID.Bytes()))
}

func SeedToKeyPair(mySeed string, index uint64) *cryptolib.KeyPair {
	seedBytes, err := base58.Decode(mySeed)
	if err != nil {
		panic(err)
	}
	db := utxodb.New(utxodb.DefaultInitParams(seedBytes))
	pair, _ := db.NewKeyPairByIndex(index)
	return pair
}
