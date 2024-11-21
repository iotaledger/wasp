package tests

import (
	"context"
	"crypto/ecdsa"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/contracts/inccounter"
)

func (e *ChainEnv) expectCounter(counter int64) {
	c := e.getNativeContractCounter()
	require.EqualValues(e.t, counter, c)
}

func (e *ChainEnv) getNativeContractCounter() int64 {
	return e.getCounterForNode(0)
}

func (e *ChainEnv) getCounterForNode(nodeIndex int) int64 {
	result, _, err := e.Chain.Cluster.WaspClient(nodeIndex).ChainsApi.
		CallView(context.Background(), e.Chain.ChainID.String()).
		ContractCallViewRequest(apiclient.ContractCallViewRequest{
			ContractHName: inccounter.Contract.Hname().String(),
			FunctionName:  inccounter.ViewGetCounter.Name,
		}).Execute()
	require.NoError(e.t, err)

	decodedDict, err := apiextensions.APIResultToCallArgs(result)
	require.NoError(e.t, err)

	counter, err := inccounter.ViewGetCounter.DecodeOutput(decodedDict)
	require.NoError(e.t, err)

	return counter
}

func (e *ChainEnv) waitUntilCounterEquals(expected int64, duration time.Duration) {
	timeout := time.After(duration)
	var c int64
	allNodesEqualFun := func() bool {
		for _, node := range e.Chain.AllPeers {
			c = e.getCounterForNode(node)
			if c != expected {
				return false
			}
		}
		return true
	}
	for {
		select {
		case <-timeout:
			e.t.Errorf("timeout waiting for inccounter, current: %d, expected: %d", c, expected)
			e.t.Fatal()
		default:
			if allNodesEqualFun() {
				return // success
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func newEthereumAccount() (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	return key, crypto.PubkeyToAddress(key.PublicKey)
}
