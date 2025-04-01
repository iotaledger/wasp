// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/migration"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/wallets"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	cliutil "github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func PrintIotaBlock(block iotajsonrpc.IotaProgrammableTransactionBlock) {
	fmt.Println("IotaProgrammableTransactionBlock:")

	// Print Inputs
	fmt.Println("Inputs:")
	for i, input := range block.Inputs {
		fmt.Printf("%d) ", i)
		prettyPrint(input, 1)
		fmt.Println()
	}

	// Print Commands
	fmt.Println("Commands:")
	for i, command := range block.Commands {
		fmt.Printf("%d) ", i)
		prettyPrint(command, 1)
		fmt.Println()
	}
}

func prettyPrint(v interface{}, indent int) {
	padding := strings.Repeat("  ", indent)

	switch val := v.(type) {
	case map[string]interface{}:
		fmt.Println()
		for key, value := range val {
			fmt.Printf("%s%s: ", padding, key)
			prettyPrint(value, indent+1)
		}
	case []interface{}:
		if len(val) == 0 {
			fmt.Println("[]")
			return
		}
		fmt.Println()
		for _, item := range val {
			fmt.Printf("%s- ", padding)
			prettyPrint(item, indent+1)
		}
	case map[interface{}]interface{}:
		fmt.Println()
		for key, value := range val {
			fmt.Printf("%s%v: ", padding, key)
			prettyPrint(value, indent+1)
		}
	default:
		fmt.Printf("%v\n", val)
	}
}

func getFirstCoin(ctx context.Context, kp wallets.Wallet, client clients.L1Client, excludedCoin *iotago.ObjectID) *iotago.ObjectRef {
	coinType := iotajsonrpc.IotaCoinType.String()
	coins := lo.Must(client.GetCoins(ctx, iotaclient.GetCoinsRequest{
		CoinType: &coinType,
		Owner:    kp.Address().AsIotaAddress(),
	}))

	// This is only used for the GasCoin in the Prepare->Run transition.
	// We don't want to accidentally use the destined GasCoin for the Committee to pay for transactions.
	if excludedCoin != nil {
		var selectedCoin *iotago.ObjectRef

		lo.Filter(coins.Data, func(item *iotajsonrpc.Coin, index int) bool {
			if bytes.Equal(item.CoinObjectID.Bytes(), excludedCoin.Bytes()) {
				return true
			}

			selectedCoin = item.Ref()
			return false
		})

		return selectedCoin
	}

	return coins.Data[0].Ref()
}

/*
*
The prepare function is responsible for
* running the L1 migration
* creating the initial empty Anchor object
* initialize the chain state
* return a configuration object to the L2 migration tool
*/
func migrationPrepare(ctx context.Context, node string, packageID iotago.PackageID, kp wallets.Wallet, l1Client clients.L1Client, peers []string, quorum int) {
	cwd, err := os.Getwd()
	log.Check(err)

	dkgCommitteeAddress := doDKG(ctx, node, peers, quorum)

	fmt.Println("Creating Anchor")
	anchor, err := l1Client.L2().StartNewChain(ctx, &iscmoveclient.StartNewChainRequest{
		PackageID:         packageID,
		GasPayments:       []*iotago.ObjectRef{getFirstCoin(ctx, kp, l1Client, nil)},
		ChainOwnerAddress: kp.Address(),
		Signer:            kp,
		GasPrice:          iotaclient.DefaultGasPrice,
		GasBudget:         iotaclient.DefaultGasBudget,
		StateMetadata:     make([]byte, 0),
		InitCoinRef:       nil,
	})
	log.Check(err)

	prepareConfiguration := &migration.PrepareConfiguration{
		DKGCommitteeAddress: dkgCommitteeAddress,
		AnchorID:            anchor.ObjectID,
		ChainOwner:          kp.Address(),
		PackageID:           packageID,
		L1ApiUrl:            config.L1APIAddress(),
	}

	fmt.Println("Preparation finished:")
	serializeConfig := lo.Must(json.MarshalIndent(prepareConfiguration, "", "  "))
	fmt.Println(string(serializeConfig))
	log.Check(os.WriteFile("migration_preparation.json", serializeConfig, 0644))

	fmt.Printf("The 'migration_preperation.json' file has been written to your current cwd (%s)", cwd)
}

func migrationRun(ctx context.Context, node string, chainName string, packageID iotago.PackageID, kp wallets.Wallet, l1Client clients.L1Client) {
	configBytes := lo.Must(os.ReadFile("migration_preparation.json"))
	var prepareConfig migration.PrepareConfiguration
	log.Check(json.Unmarshal(configBytes, &prepareConfig))

	resultBytes := lo.Must(os.ReadFile("migration_result.json"))
	var migrationResult migration.MigrationResult
	log.Check(json.Unmarshal(resultBytes, &migrationResult))

	anchor, err := l1Client.GetObject(ctx, iotaclient.GetObjectRequest{
		ObjectID: prepareConfig.AnchorID,
	})
	log.Check(err)

	anchorRef := anchor.Data.Ref()

	l1Params, err := parameters.FetchLatest(ctx, l1Client.IotaClient())
	log.Check(err)

	fmt.Println("Creating GasCoin")
	gasCoin, err := cliutil.CreateAndSendGasCoin(ctx, l1Client, kp, kp.Address().AsIotaAddress(), l1Params)

	stateMetadata, err := transaction.StateMetadataFromBytes(hexutil.MustDecode(migrationResult.StateMetadataHex))
	log.Check(err)
	stateMetadata.GasCoinObjectID = &gasCoin

	// Update the StateAnchor with the latest migrated block info
	success, err := l1Client.L2().UpdateAnchorStateMetadata(ctx, &iscmoveclient.UpdateAnchorStateMetadataRequest{
		AnchorRef:     &anchorRef,
		StateMetadata: stateMetadata.Bytes(),
		StateIndex:    migrationResult.StateIndex,
		PackageID:     packageID,
		Signer:        kp,
		GasBudget:     iotaclient.DefaultGasBudget,
		GasPrice:      iotaclient.DefaultGasPrice,
		GasPayments: []*iotago.ObjectRef{
			getFirstCoin(ctx, kp, l1Client, &gasCoin),
		},
	})

	if !success || err != nil {
		log.Printf("FAILED TO UPDATE STATE! result:%v, \nerr:%v", success, err)
		panic("omg")
	}

	// Transfer the GasCoin to the Committee
	transferGasCoin, err := l1Client.TransferObject(ctx, iotaclient.TransferObjectRequest{
		Signer:    kp.Address().AsIotaAddress(),
		ObjectID:  &gasCoin,
		Recipient: prepareConfig.DKGCommitteeAddress.AsIotaAddress(),
		GasBudget: iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		Gas:       getFirstCoin(ctx, kp, l1Client, &gasCoin).ObjectID,
	})
	if err != nil {
		log.Printf("FAILED TO CONSTRUCT -TRANSFER ANCHOR-: err:%v", err)
		panic("omg")
	}

	result, err := l1Client.SignAndExecuteTransaction(ctx, &iotaclient.SignAndExecuteTransactionRequest{
		Signer:      cryptolib.SignerToIotaSigner(kp),
		TxDataBytes: transferGasCoin.TxBytes,
		Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
			ShowObjectChanges: true,
		},
	})
	if err != nil {
		log.Printf("FAILED TO EXECUTE -TRANSFER GASCOIN-: err:%v", err)
		panic("omg")
	}

	// Transfer the Anchor to the Committee
	transferAnchor, err := l1Client.TransferObject(ctx, iotaclient.TransferObjectRequest{
		Signer:    kp.Address().AsIotaAddress(),
		ObjectID:  anchorRef.ObjectID,
		Recipient: prepareConfig.DKGCommitteeAddress.AsIotaAddress(),
		GasBudget: iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		Gas:       getFirstCoin(ctx, kp, l1Client, &gasCoin).ObjectID,
	})
	if err != nil {
		log.Printf("FAILED TO CONSTRUCT -TRANSFER ANCHOR-: err:%v", err)
		panic("omg")
	}

	result, err = l1Client.SignAndExecuteTransaction(ctx, &iotaclient.SignAndExecuteTransactionRequest{
		Signer:      cryptolib.SignerToIotaSigner(kp),
		TxDataBytes: transferAnchor.TxBytes,
		Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
			ShowObjectChanges: true,
		},
	})
	if err != nil {
		log.Printf("FAILED TO EXECUTE -TRANSFER ANCHOR-: err:%v", err)
		panic("omg")
	}

	log.Printf("\n%v\n %v\n", transferAnchor, result)

	chainID := isc.ChainIDFromObjectID(*prepareConfig.AnchorID)
	config.AddChain(chainName, chainID.String())
	activateChain(ctx, node, chainName, chainID)
}

func initMigrateDeployPrepareCmd() *cobra.Command {
	var (
		node             string
		peers            []string
		quorum           int
		evmChainID       uint16
		blockKeepAmount  int32
		govControllerStr string
		chainName        string
	)

	cmd := &cobra.Command{
		Use:   "migrate-chain --chain=<name> 1) prepare, 2) run",
		Short: "Migrates a Stardust chain to Rebased",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chainName = defaultChainFallback(chainName)
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()

			if !util.IsSlug(chainName) {
				log.Fatalf("invalid chain name: %s, must be in slug format, only lowercase and hyphens, example: foo-bar", chainName)
			}

			l1Client := cliclients.L1Client()
			kp := wallet.Load()
			packageID := config.GetPackageID()

			if args[0] == "prepare" {
				migrationPrepare(ctx, node, packageID, kp, l1Client, peers, quorum)
				return
			}

			if args[0] == "run" {
				migrationRun(ctx, node, chainName, packageID, kp, l1Client)
				return
			}
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	waspcmd.WithPeersFlag(cmd, &peers)
	cmd.Flags().Uint16VarP(&evmChainID, "evm-chainid", "", evm.DefaultChainID, "ChainID")
	cmd.Flags().Int32VarP(&blockKeepAmount, "block-keep-amount", "", governance.DefaultBlockKeepAmount, "Amount of blocks to keep in the blocklog (-1 to keep all blocks)")
	cmd.Flags().StringVar(&chainName, "chain", "", "name of the chain")
	log.Check(cmd.MarkFlagRequired("chain"))
	cmd.Flags().IntVar(&quorum, "quorum", 0, "quorum (default: 3/4s of the number of committee nodes)")
	cmd.Flags().StringVar(&govControllerStr, "gov-controller", "", "governance controller address")
	return cmd
}
