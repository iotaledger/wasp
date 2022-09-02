package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
)

const (
	defaultBechPrefix = "rms"
	getOutputIdURL    = "https://api.testnet.shimmer.network/api/indexer/v1/outputs/alias/"
	getAliasURL       = "https://api.testnet.shimmer.network/api/core/v2/outputs/"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: ask-anchor <chainID> [<Bech prefix>]\n")
		os.Exit(1)
	}
	aliasID := os.Args[1]
	bechPrefix := defaultBechPrefix
	if len(os.Args) > 2 {
		bechPrefix = os.Args[2]
	}
	parameters.L1ForTesting.Protocol.Bech32HRP = iotago.NetworkPrefix(bechPrefix)
	parameters.InitL1(parameters.L1ForTesting)
	chainID, err := isc.ChainIDFromString(aliasID)
	mustNoErr(err)
	fmt.Printf("chainid = %s\naliasID = %s\n", chainID, chainID.AsAliasID())

	// get output ID for alias ID
	body := mustGetURL(getOutputIdURL, chainID.AsAliasID().String())
	var decoded map[string]any
	err = json.Unmarshal(body, &decoded)
	mustNoErr(err)

	// get alias output by outputID
	var parsed map[string]map[string]any
	body = mustGetURL(getAliasURL, decoded["items"].([]any)[0].(string))
	err = json.Unmarshal(body, &parsed)
	mustNoErr(err)

	fmt.Printf("stateIndex: %v\n", parsed["output"]["stateIndex"])
	fmt.Printf("amount: %v\n", parsed["output"]["amount"])
	fmt.Printf("foundryCounter: %v\n", parsed["output"]["foundryCounter"])
	stateMetadataBin, err := iotago.DecodeHex(parsed["output"]["stateMetadata"].(string))
	mustNoErr(err)
	l1Commitment, err := state.L1CommitmentFromBytes(stateMetadataBin)
	mustNoErr(err)
	fmt.Printf("L1Commitment:\n     state commitment: %s\n     block hash:       %s\n",
		l1Commitment.StateCommitment, l1Commitment.BlockHash)
}

func mustNoErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func mustGetURL(url1, url2 string) []byte {
	resp, err := http.Get(url1 + url2)
	mustNoErr(err)
	defer func() {
		_ = resp.Body.Close()
	}()

	//We Read the response body on the line below.
	body, err := io.ReadAll(resp.Body)
	mustNoErr(err)

	return body
}

//
//func jsonPrettyPrint(in []byte) []byte {
//	var out bytes.Buffer
//	err := json.Indent(&out, in, "", "\t")
//	if err != nil {
//		return in
//	}
//	return out.Bytes()
//}
