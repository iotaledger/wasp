// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
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
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
)

func getAllWaspNodes() []int {
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
		committee        []int
		quorum           int
		description      string
		evmParams        evmDeployParams
		govControllerStr string
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new chain",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			l1Client := config.L1Client()

			if ok, _ := isEnoughQuorum(len(committee), quorum); !ok {
				log.Fatal("quorum needs to be bigger than 1/3 of committee size")
			}

			committeePubKeys := make([]string, 0)
			for _, api := range config.CommitteeAPI(committee) {
				peerInfo, err := config.WaspClient(api).GetPeeringSelf()
				log.Check(err)
				committeePubKeys = append(committeePubKeys, peerInfo.PubKey)
			}

			chainid, _, err := apilib.DeployChainWithDKG(apilib.CreateChainParams{
				AuthenticationToken:  config.GetToken(),
				Layer1Client:         l1Client,
				CommitteeAPIHosts:    config.CommitteeAPI(committee),
				CommitteePubKeys:     committeePubKeys,
				N:                    uint16(len(committee)),
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

			var alias string
			if len(args) == 0 {
				alias = GetChainAlias()
			} else {
				alias = args[0]
			}
			AddChainAlias(alias, chainid.String())
		},
	}

	cmd.Flags().IntSliceVarP(&committee, "committee", "", []int{0}, "peers acting as committee nodes (ex: 0,1,2,3) (default: all nodes)")
	cmd.Flags().IntVarP(&quorum, "quorum", "", 1, "quorum (default: 3/4s of the number of committee nodes)")
	cmd.Flags().StringVarP(&description, "description", "", "", "description")
	cmd.Flags().StringVarP(&govControllerStr, "gov-controller", "", "", "governance controller address")

	evmParams.initFlags(cmd)
	return cmd
}
