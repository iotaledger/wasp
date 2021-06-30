// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/client/goshimmer"
	"github.com/iotaledger/wasp/contracts/native/evmchain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/chainid"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/evm/evmcli"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/mr-tron/base58"
	"github.com/spf13/cobra"
)

type Params struct {
	waspHost         string
	goshimmerHost    string
	chainIDBase58    string
	privateKeyBase58 string
	deploy           evmcli.DeployParams
}

var (
	params        Params
	jsonRPCServer evmcli.JSONRPCServer
)

func main() {
	rootCmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "evmproxy",
		Short: "evmproxy is a command-line tool to manage Ethereum smart contracts running on ISCP",
		Long: `evmproxy is a command-line tool to manage Ethereum smart contracts running on ISCP.

To use evmproxy you need a previously deployed ISCP chain, e.g. via the wasp-cli tool. With evmproxy you can:

- deploy a new Ethereum blockchain on top of the ISCP chain
- run a JSON-RPC service to interact with your Ethereum blockchain
`,
		Run: func(cmd *cobra.Command, args []string) {
			log.Check(cmd.Help())
		},
	}
	rootCmd.PersistentFlags().StringVarP(&params.goshimmerHost, "goshimmer", "g", "127.0.0.1:8080", "Goshimmer API location")
	rootCmd.PersistentFlags().StringVarP(&params.waspHost, "wasp", "w", "127.0.0.1:9090", "Wasp API location")
	rootCmd.PersistentFlags().StringVarP(&params.chainIDBase58, "chainid", "c", "", "Chain ID (base58)")
	rootCmd.PersistentFlags().StringVarP(&params.privateKeyBase58, "private-key", "k", "", "Private key for signing IOTA transactions (base58)")

	log.Init(rootCmd)

	deployCmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "deploy",
		Short: "Deploy the evmchain ISCP contract (which creates a new Ethereum blockchain)",
		Long: fmt.Sprintf(`Deploy the evmchain ISCP contract (which creates a new Ethereum blockchain).

The default genesis allocation is: %s:%d
                                   private key: %s`,
			evmtest.FaucetAddress,
			evmtest.FaucetSupply,
			hex.EncodeToString(crypto.FromECDSA(evmtest.FaucetKey)),
		),
		Run: deploy,
	}
	params.deploy.InitFlags(deployCmd)
	rootCmd.AddCommand(deployCmd)

	jsonRPCCmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "jsonrpc",
		Short: "Start a JSON-RPC service to interact with an Ethereum blockchain running on ISCP",
		Long: `Start a JSON-RPC service to interact with an Ethereum blockchain running on ISCP.

By default the server has no unlocked accounts. To send transactions, either:

- use eth_sendRawTransaction
- configure an unlocked account with --account, and use eth_sendTransaction`,
		Run: startJSONRPC,
	}
	jsonRPCServer.InitFlags(jsonRPCCmd)
	rootCmd.AddCommand(jsonRPCCmd)

	log.Check(rootCmd.Execute())
}

func deploy(cmd *cobra.Command, args []string) {
	tx, err := chainClient().Post1Request(
		root.Interface.Hname(),
		coretypes.Hn(root.FuncDeployContract),
		chainclient.PostRequestParams{
			Args: requestargs.New().AddEncodeSimpleMany(codec.MakeDict(map[string]interface{}{
				root.ParamName:             "evmchain",
				root.ParamDescription:      "EVM chain",
				root.ParamProgramHash:      evmchain.Interface.ProgramHash,
				evmchain.FieldGenesisAlloc: evmchain.EncodeGenesisAlloc(params.deploy.GetGenesis()),
				evmchain.FieldGasPerIota:   params.deploy.GasPerIOTA,
			})),
		},
	)
	log.Check(err)
	log.Printf("Posted evmchain contract deployment request. Transaction ID: %s\n", tx.ID().Base58())
}

func chainClient() *chainclient.Client {
	return chainclient.New(goshimmerClient(), waspClient(), chainID(), signer())
}

func goshimmerClient() *goshimmer.Client {
	return goshimmer.NewClient(params.goshimmerHost, -1)
}

func waspClient() *client.WaspClient {
	return client.NewWaspClient(params.waspHost)
}

func chainID() chainid.ChainID {
	if params.chainIDBase58 == "" {
		log.Fatalf("--chainid is mandatory")
	}
	chainID, err := chainid.ChainIDFromBase58(params.chainIDBase58)
	log.Check(err)
	return *chainID
}

func signer() *ed25519.KeyPair {
	if params.privateKeyBase58 == "" {
		log.Fatalf("--private-key is mandatory")
	}
	b, err := base58.Decode(params.privateKeyBase58)
	log.Check(err)
	priv, err, _ := ed25519.PrivateKeyFromBytes(b)
	log.Check(err)
	return &ed25519.KeyPair{
		PrivateKey: priv,
		PublicKey:  priv.Public(),
	}
}

func startJSONRPC(cmd *cobra.Command, args []string) {
	backend := jsonrpc.NewWaspClientBackend(chainClient())
	jsonRPCServer.ServeJSONRPC(backend)
}
