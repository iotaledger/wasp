package main

import (
	"context"
	"fmt"
	"os"

	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
)

const (
	APIAddress = "https://api.testnet.shimmer.network"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: ask-anchor <chainID>\n")
		os.Exit(1)
	}

	chainID, err := isc.ChainIDFromString(os.Args[1])
	mustNoErr(err)

	indexerClient, err := nodeclient.New(APIAddress).Indexer(context.Background())
	mustNoErr(err)
	stateOutputID, stateOutput, err := indexerClient.Alias(context.Background(), *chainID.AsAliasID())
	mustNoErr(err)

	fmt.Printf("outputID: %v\n", stateOutputID.ToHex())
	fmt.Printf("stateIndex: %v\n", stateOutput.StateIndex)
	fmt.Printf("amount: %v\n", stateOutput.Deposit())
	fmt.Printf("foundryCounter: %v\n", stateOutput.FoundryCounter)
	l1Commitment, err := state.L1CommitmentFromBytes(stateOutput.StateMetadata)
	mustNoErr(err)
	fmt.Printf("L1Commitment:\n     state commitment: %s\n     block hash:       %s\n",
		l1Commitment.TrieRoot(), l1Commitment.BlockHash())
}

func mustNoErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
