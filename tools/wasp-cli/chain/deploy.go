// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"
	"math"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func GetAllWaspNodes() []int {
	ret := []int{}
	for index := range viper.GetStringMap("wasp") {
		i, err := strconv.Atoi(index)
		log.Check(err)
		ret = append(ret, i)
	}
	return ret
}

func defaultQuorum(n int) int {
	quorum := int(math.Ceil(3 * float64(n) / 4))
	if quorum < 1 {
		quorum = 1
	}
	return quorum
}

func isEnoughQuorum(n, t int) (bool, int) {
	maxF := (n - 1) / 3
	return t >= (n - maxF), maxF
}

func controllerAddr(addr string) iotago.Address {
	if addr == "" {
		return wallet.Load().Address()
	}
	prefix, govControllerAddr, err := iotago.ParseBech32(addr)
	log.Check(err)
	if parameters.L1().Protocol.Bech32HRP != prefix {
		log.Fatalf("unexpected prefix. expected: %s, actual: %s", parameters.L1().Protocol.Bech32HRP, prefix)
	}
	return govControllerAddr
}

func initDeployCmd() *cobra.Command {
	var (
		node             []string
		quorum           int
		description      string
		evmParams        evmDeployParams
		govControllerStr string
		chainName        string
	)

	cmd := &cobra.Command{
		Use:   "deploy [<alias>]",
		Short: "Deploy a new chain",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			l1Client := cliclients.L1Client()
			node = waspcmd.DefaultNodesFallback(node)

			if quorum == 0 {
				quorum = defaultQuorum(len(node))
			}

			if ok, _ := isEnoughQuorum(len(node), quorum); !ok {
				log.Fatal("quorum needs to be bigger than 1/3 of committee size")
			}

			committeePubKeys := make([]string, 0)
			for _, apiIndex := range node {
				peerInfo, _, err := cliclients.WaspClient(apiIndex).NodeApi.GetPeeringIdentity(context.Background()).Execute()
				log.Check(err)
				committeePubKeys = append(committeePubKeys, peerInfo.PublicKey)
			}

			chainid, _, err := apilib.DeployChainWithDKG(cliclients.WaspClientForHostName, apilib.CreateChainParams{
				Layer1Client:         l1Client,
				CommitteeAPIHosts:    config.NodeAPIURLs(node),
				CommitteePubKeys:     committeePubKeys,
				N:                    uint16(len(node)),
				T:                    uint16(quorum),
				OriginatorKeyPair:    wallet.Load().KeyPair,
				Description:          description,
				Textout:              os.Stdout,
				GovernanceController: controllerAddr(govControllerStr),
				InitParams: dict.Dict{
					root.ParamEVM(evm.FieldChainID):         codec.EncodeUint16(evmParams.ChainID),
					root.ParamEVM(evm.FieldGenesisAlloc):    evmtypes.EncodeGenesisAlloc(evmParams.getGenesis(nil)),
					root.ParamEVM(evm.FieldBlockGasLimit):   codec.EncodeUint64(evmParams.BlockGasLimit),
					root.ParamEVM(evm.FieldBlockKeepAmount): codec.EncodeInt32(evmParams.BlockKeepAmount),
				},
			})
			log.Check(err)

			config.AddChain(chainName, chainid.String())
		},
	}

	waspcmd.WithWaspNodesFlag(cmd, &node)
	cmd.Flags().StringVar(&chainName, "chain", "", "name of the chain)")
	log.Check(cmd.MarkFlagRequired("chain"))
	cmd.Flags().IntVar(&quorum, "quorum", 0, "quorum (default: 3/4s of the number of committee nodes)")
	cmd.Flags().StringVar(&description, "description", "", "description")
	cmd.Flags().StringVar(&govControllerStr, "gov-controller", "", "governance controller address")

	evmParams.initFlags(cmd)
	return cmd
}
