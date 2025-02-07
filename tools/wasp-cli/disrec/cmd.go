package disrec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/iotaledger/hive.go/app/configuration"
	appLogger "github.com/iotaledger/hive.go/app/logger"
	hivep2p "github.com/iotaledger/hive.go/crypto/p2p"
	"github.com/iotaledger/hive.go/logger"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initDisrecCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disrec <command>",
		Short: "Disaster recovery utils.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			log.Check(cmd.Help())
		},
	}
}

func initSignAndPostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign_post <tx_bytes_file> <committee_address> <committee_keys_dir> <iota_ws_url>",
		Short: "Read unsigned TX byted from the file, sign it using the committee partial keys and send to the L1 network.",
		Long: `
			We assume contents of <tx_bytes_file> contains bytes corresponding to the\n
			serialized *iotago.TransactionData.\n
			\n
			The <committee_keys_dir> should contain directories from distinct nodes,\n
			each containing the waspdb contents, namely:\n
			  - <committee_keys_dir>/<any>/identity/identity.key\n
			  - <committee_keys_dir>/<any>/dkshares/0x...hex....json\n
			`,
		Args: cobra.ExactArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			//
			// Read the serialized TX Data.
			txBytesFile := args[0]
			txBytesRaw := lo.Must(os.ReadFile(txBytesFile))

			// To make handling unsigned tx easier, a hex encoded value might make sense,
			// in case we want to print the data instead of exporting a file directly.
			txBytes := lo.Must(hexutil.Decode(string(txBytesRaw)))

			//
			// Parse the committee address.
			committeeAddressStr := args[1]
			committeeAddress := lo.Must(cryptolib.AddressFromHex(committeeAddressStr))
			//
			// Read the node keys and construct the DK Registries and the signer.
			committeeKeysDir := args[2]
			if !lo.Must(os.Stat(committeeKeysDir)).IsDir() {
				panic("committee keys must be a directory")
			}

			var nodeIDs []gpa.NodeID
			var peerIdentities []*cryptolib.KeyPair
			var dkRegistries []registry.DKShareRegistryProvider

			for _, entry := range lo.Must(os.ReadDir(committeeKeysDir)) {
				if !entry.IsDir() {
					continue
				}
				identityPath := filepath.Join(committeeKeysDir, entry.Name(), "identity", "identity.key")
				if lo.Must(os.Stat(identityPath)).IsDir() {
					continue
				}

				dkSharesDir := filepath.Join(committeeKeysDir, entry.Name(), "dkshares")
				dkSharePath := filepath.Join(dkSharesDir, committeeAddressStr+".json")
				if lo.Must(os.Stat(dkSharePath)).IsDir() {
					continue
				}

				privKeyRaw, newlyCreated, err := hivep2p.LoadOrCreateIdentityPrivateKey(identityPath, "")
				if err != nil || newlyCreated {
					continue
				}

				privKey := lo.Must(cryptolib.PrivateKeyFromBytes(lo.Must(privKeyRaw.Raw())))
				keyPair := cryptolib.KeyPairFromPrivateKey(privKey)
				nodeID := gpa.NodeIDFromPublicKey(keyPair.GetPublicKey())

				nodeIDs = append(nodeIDs, nodeID)
				peerIdentities = append(peerIdentities, keyPair)
				dkRegistries = append(dkRegistries, lo.Must(registry.NewDKSharesRegistry(dkSharesDir, privKey)))
			}
			_ = appLogger.InitGlobalLogger(configuration.New())
			log := logger.NewLogger("disrec")
			signer := testpeers.NewTestDSSSigner(committeeAddress, dkRegistries, nodeIDs, peerIdentities, log)

			//
			// Sign and Post the TX to the L1.
			iotaWsUrl := args[3]
			ctx := context.Background()
			wsClient := lo.Must(iscmoveclient.NewWebsocketClient(ctx, iotaWsUrl, "", iotaclient.WaitForEffectsEnabled, log))
			res, err := wsClient.SignAndExecuteTransaction(ctx, &iotaclient.SignAndExecuteTransactionRequest{
				TxDataBytes: txBytes,
				Signer:      cryptolib.SignerToIotaSigner(signer),
				Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
					ShowEffects:        true,
					ShowObjectChanges:  true,
					ShowBalanceChanges: true,
					ShowEvents:         true,
				},
			})
			if err != nil {
				panic(fmt.Errorf("error executing tx: %s Res: %v", err, res))
			}
			if !res.Effects.Data.IsSuccess() {
				panic(fmt.Errorf("error executing tx: %s Digest: %s", res.Effects.Data.V1.Status.Error, res.Digest))
			}

			log.Infof("Transaction posted! Digest: %s\n", res.Digest)
			log.Info("Transaction data:")

			log.Infof("Object Changes:\n%v\n", string(lo.Must(json.MarshalIndent(res.ObjectChanges, "\t", " "))))
			log.Infof("Effects:\n%v\n", string(lo.Must(json.MarshalIndent(res.Effects, "\t", " "))))
		},
	}
	return cmd
}

func Init(rootCmd *cobra.Command) {
	disrecCmd := initDisrecCmd()
	disrecCmd.AddCommand(initSignAndPostCmd())
	rootCmd.AddCommand(disrecCmd)
}
