// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscutils

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm/evmtest"
)

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func TestPRNGLibrary(t *testing.T) {
	env := evmtest.InitEVM(t)
	ethKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

	prngTest := env.DeployContract(ethKey, PRNGTestContractABI, PRNGTestContractBytecode)

	// ensure we get a pseudorandom number based on the seed 'test'
	value := new(big.Int)
	prngTest.CallFnExpectEvent(nil, "RandomNumberGenerated", &value, "generateRandomNumber")
	require.EqualValues(t, "58061745822097596726174715997747576974229114500566746577421751887060998178974", value.String())

	// generate another random number and ensure it is different (state was updated)
	value2 := new(big.Int)
	prngTest.CallFnExpectEvent(nil, "RandomNumberGenerated", &value, "generateRandomNumber")
	require.NotEqualValues(t, value2.String(), value.String())

	// test for returned random hash
	var entropy hashing.HashValue
	prngTest.CallFnExpectEvent(nil, "RandomHashGenerated", &entropy, "generateRandomHash")
	require.NotEqualValues(t, hashing.NilHash, entropy)

	// test random number in range generation
	prngTest.CallFnExpectEvent(nil, "RandomNumberGenerated", &value, "generateRandomNumberInRange", big.NewInt(0), big.NewInt(1))
	require.EqualValues(t, "0", value.String())

	prngTest.CallFnExpectEvent(nil, "RandomNumberGenerated", &value, "generateRandomNumberInRange", big.NewInt(1), big.NewInt(2))
	require.EqualValues(t, "1", value.String())

	// generate a number less than max uint64
	prngTest.CallFnExpectEvent(nil, "RandomNumberGenerated", &value, "generateRandomNumberInRange", big.NewInt(0), big.NewInt(math.MaxInt64))
	require.LessOrEqual(t, value.Int64(), int64(math.MaxInt64))

	// generate a random number between int64 max and uint64 max
	prngTest.CallFnExpectEvent(nil, "RandomNumberGenerated", &value, "generateRandomNumberInRange", big.NewInt(math.MaxInt64), new(big.Int).SetUint64(math.MaxUint64))
	require.Less(t, value.Uint64(), uint64(math.MaxUint64))
	require.GreaterOrEqual(t, value.Uint64(), uint64(math.MaxInt64))
}
