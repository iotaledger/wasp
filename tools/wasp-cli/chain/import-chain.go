package chain

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/database"
	"github.com/iotaledger/wasp/v2/packages/kvstore/rocksdb"
	"github.com/iotaledger/wasp/v2/packages/origin"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/transaction"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

func openChainAndRead(dbPath string) (transaction.StateMetadata, uint32, error) {
	dbConn := lo.Must(rocksdb.OpenDBReadOnly(dbPath))
	db := database.New(dbPath, rocksdb.New(dbConn), hivedb.EngineRocksDB, false, func() bool {
		return false
	})

	rebasedDBStore, err := state.NewStoreReadonly(db.KVStore())
	if err != nil {
		return transaction.StateMetadata{}, 0, fmt.Errorf("failed to open read only chain db: %w", err)
	}

	latestBlock := lo.Must(rebasedDBStore.LatestBlock())
	latestState := lo.Must(rebasedDBStore.LatestState())

	governanceReader := governance.NewStateReaderFromChainState(latestState)

	evmContractPart := evm.ContractPartitionR(latestState)
	evmChainID := emulator.GetChainIDFromBlockChainDBState(emulator.BlockchainDBSubrealmR(
		evm.EmulatorStateSubrealmR(evmContractPart),
	))

	initParams := &origin.InitParams{
		DeployTestContracts: true,
		EVMChainID:          evmChainID,
		BlockKeepAmount:     governanceReader.GetBlockKeepAmount(),
		ChainAdmin:          governanceReader.GetChainAdmin(),
	}

	anchorStateMetadata := transaction.StateMetadata{
		L1Commitment:  latestBlock.L1Commitment(),
		SchemaVersion: latestState.SchemaVersion(),
		InitDeposit:   0,
		InitParams:    initParams.Encode(),
		PublicURL:     governanceReader.GetPublicURL(),
		GasFeePolicy:  governanceReader.GetGasFeePolicy(),
	}

	return anchorStateMetadata, latestBlock.StateIndex(), nil
}

func initImportCmd() *cobra.Command {
	var (
		node            string
		peers           []string
		quorum          int
		chainName       string
		iscPackageIDStr string
	)

	cmd := &cobra.Command{
		Use:   "import <path_to_wasp_db_chain> --chain=<name>",
		Short: "Helps importing an existing wasp chain. Call 'chain import --help for further information'",
		Long: "This command helps importing an existing wasp chain db. Eg. the IOTA EVM Mainnet DB.\n" +
			"It reads the metadata of the wasp chain from a local directory, then recreates the Anchors state metadata, creates a GasCoin and Anchor with equal contents. Then it deploys an new chain on a local wasp instance.\n" +
			"After the deployment succeeded, you will need to either link or move the wasp chain files into 'waspdb/chains/data/<chainID>' and call 'chain activate'",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chainName = defaultChainFallback(chainName)
			kp := wallet.Load()

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()

			var iscPackageID = &iotago.PackageID{}
			var err error

			if iscPackageIDStr != "" {
				iscPackageID, err = iotago.PackageIDFromHex(iscPackageIDStr)
				log.Check(err)
			} else {
				log.Printf("Deploying Move contract...\n")
				l1Client := cliclients.L1Client()
				*iscPackageID, err = l1Client.DeployISCContracts(ctx, cryptolib.SignerToIotaSigner(kp))
				log.Check(err)
			}

			result, err := initializeDeploymentWithGasCoin(ctx, kp, node, chainName, peers, quorum)
			if err != nil {
				log.Fatal(err)
			}

			dbPath := args[0]
			anchorStateMetadata, blockIndex, err := openChainAndRead(dbPath)
			log.Check(err)
			anchorStateMetadata.GasCoinObjectID = &result.gasCoinObject

			anchor, err := cliclients.L2Client().StartNewChain(ctx, &iscmoveclient.StartNewChainRequest{
				PackageID:     *iscPackageID,
				AnchorOwner:   kp.Address(),
				Signer:        kp,
				GasPrice:      iotaclient.DefaultGasPrice,
				GasBudget:     iotaclient.DefaultGasBudget,
				StateMetadata: make([]byte, 0),
				InitCoinRef:   nil,
			})
			log.Check(err)

			_, err = cliclients.L2Client().UpdateAnchorStateMetadata(ctx, &iscmoveclient.UpdateAnchorStateMetadataRequest{
				StateIndex:    blockIndex,
				StateMetadata: anchorStateMetadata.Bytes(),
				Signer:        kp,
				GasPrice:      iotaclient.DefaultGasPrice,
				GasBudget:     iotaclient.DefaultGasBudget,
				PackageID:     *iscPackageID,
				AnchorRef:     &anchor.ObjectRef,
			})
			log.Check(err)

			transferAnchor, err := cliclients.L1Client().TransferObject(ctx, iotaclient.TransferObjectRequest{
				Signer:    kp.Address().AsIotaAddress(),
				ObjectID:  anchor.ObjectID,
				Recipient: result.committeeAddress.AsIotaAddress(),
				GasBudget: iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
			})
			if err != nil {
				log.Printf("FAILED TO CONSTRUCT -TRANSFER ANCHOR-: err:%v", err)
				panic("omg")
			}

			_, err = cliclients.L1Client().SignAndExecuteTransaction(ctx, &iotaclient.SignAndExecuteTransactionRequest{
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

			config.AddChain(chainName, anchor.ObjectID.String())

			fmt.Printf("\nChain has been deployed.\nID: %s\nStateMetadata: %v\n", anchor.ObjectID.String(), anchorStateMetadata)
			fmt.Printf("Create the following path: './waspdb/chains/data/%s' and move or link the chain files into it.\n", anchor.ObjectID.String())
			fmt.Println("Then call `chain activate` to finalize the deployment.")
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	waspcmd.WithPeersFlag(cmd, &peers)
	cmd.Flags().StringVar(&chainName, "chain", "", "name of the chain")
	log.Check(cmd.MarkFlagRequired("chain"))
	cmd.Flags().StringVar(&iscPackageIDStr, "package-id", "", "ISC L1 package ID. If not set, new package will be deployed (see `deploy-move-contract` command)")
	cmd.Flags().IntVar(&quorum, "quorum", 0, "quorum (default: 3/4s of the number of committee nodes)")

	return cmd
}
