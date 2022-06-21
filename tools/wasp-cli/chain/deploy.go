// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"os"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
	"github.com/spf13/cobra"
)

func deployCmd() *cobra.Command {
	var (
		committee   []int
		quorum      int
		description string
		evmParams   evmDeployParams
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new chain",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			alias := GetChainAlias()

			if committee == nil {
				committee = []int{0, 1, 2, 3}
			}

			committeePubKeys := make([]string, 0)
			for _, api := range config.CommitteeAPI(committee) {
				peerInfo, err := client.NewWaspClient(api).GetPeeringSelf()
				log.Check(err)
				committeePubKeys = append(committeePubKeys, peerInfo.PubKey)
			}

			chainid, _, err := apilib.DeployChainWithDKG(apilib.CreateChainParams{
				Layer1Client:      config.L1Client(),
				CommitteeAPIHosts: config.CommitteeAPI(committee),
				CommitteePubKeys:  committeePubKeys,
				N:                 uint16(len(committee)),
				T:                 uint16(quorum),
				OriginatorKeyPair: wallet.Load().KeyPair,
				Description:       description,
				Textout:           os.Stdout,
				InitParams: dict.Dict{
					root.ParamEVM(evm.FieldChainID):         codec.EncodeUint16(evmParams.ChainID),
					root.ParamEVM(evm.FieldGenesisAlloc):    evmtypes.EncodeGenesisAlloc(evmParams.getGenesis(nil)),
					root.ParamEVM(evm.FieldBlockGasLimit):   codec.EncodeUint64(evmParams.BlockGasLimit),
					root.ParamEVM(evm.FieldBlockKeepAmount): codec.EncodeInt32(evmParams.BlockKeepAmount),
					root.ParamEVM(evm.FieldGasRatio):        codec.EncodeRatio32(evmParams.GasRatio),
				},
			})
			log.Check(err)

			AddChainAlias(alias, chainid.String())
		},
	}

	cmd.Flags().IntSliceVarP(&committee, "committee", "", nil, "peers acting as committee nodes  (default: 0,1,2,3)")
	cmd.Flags().IntVarP(&quorum, "quorum", "", 3, "quorum")
	cmd.Flags().StringVarP(&description, "description", "", "", "description")
	evmParams.initFlags(cmd)
	return cmd
}
