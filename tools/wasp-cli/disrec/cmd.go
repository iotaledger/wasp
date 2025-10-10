// Package disrec implements disaster recovery functionality for the wasp-cli tool,
// allowing users to recover from various failure scenarios.
package disrec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common/hexutil"

	hivep2p "github.com/iotaledger/hive.go/crypto/p2p"
	hivelog "github.com/iotaledger/hive.go/log"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/testutil/testpeers"
)

func initDisrecCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disrec <command>",
		Short: "Disaster recovery utils.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
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
		RunE: runSignAndPost,
	}
	return cmd
}

// runSignAndPost executes the sign_post command logic.
func runSignAndPost(cmd *cobra.Command, args []string) error {
	// Read and decode the serialized TX Data.
	txBytes, err := readAndDecodeTxBytes(args[0])
	if err != nil {
		return err
	}

	// Parse the committee address.
	committeeAddressStr := args[1]
	committeeAddress, err := cryptolib.AddressFromHex(committeeAddressStr)
	if err != nil {
		return fmt.Errorf("invalid committee address '%s': %w", committeeAddressStr, err)
	}

	// Build signer from committee keys.
	nodeIDs, peerIdentities, dkRegistries, err := collectDKComponents(args[2], committeeAddressStr)
	if err != nil {
		return err
	}
	log := hivelog.NewLogger(hivelog.WithName("disrec"))
	signer := testpeers.NewTestDistributedSignatureSigner(committeeAddress, dkRegistries, nodeIDs, peerIdentities, log)

	// Sign and Post the TX to the L1.
	iotaL1ClientURL := args[3]
	ctx := context.Background()
	httpClient := iscmoveclient.NewHTTPClient(iotaL1ClientURL, "", iotaclient.WaitForEffectsEnabled)
	res, execErr := httpClient.SignAndExecuteTransaction(ctx, &iotaclient.SignAndExecuteTransactionRequest{
		TxDataBytes: txBytes,
		Signer:      cryptolib.SignerToIotaSigner(signer),
		Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
			ShowEffects:        true,
			ShowObjectChanges:  true,
			ShowBalanceChanges: true,
			ShowEvents:         true,
		},
	})
	if execErr != nil {
		return fmt.Errorf("error executing tx: %w, res: %v", execErr, res)
	}
	if !res.Effects.Data.IsSuccess() {
		return fmt.Errorf("error executing tx: %s, digest: %s", res.Effects.Data.V1.Status.Error, res.Digest)
	}

	log.LogInfof("Transaction posted! Digest: %s\n", res.Digest)
	log.LogInfo("Transaction data:")

	if objChanges, err := json.MarshalIndent(res.ObjectChanges, "\t", " "); err == nil {
		log.LogInfof("Object Changes:\n%v\n", string(objChanges))
	}
	if effects, err := json.MarshalIndent(res.Effects, "\t", " "); err == nil {
		log.LogInfof("Effects:\n%v\n", string(effects))
	}
	return nil
}

// readAndDecodeTxBytes reads the file and hex-decodes its contents.
func readAndDecodeTxBytes(txBytesFile string) ([]byte, error) {
	txBytesRaw, err := os.ReadFile(txBytesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read tx bytes file '%s': %w", txBytesFile, err)
	}
	// To make handling unsigned tx easier, a hex encoded value might make sense,
	// in case we want to print the data instead of exporting a file directly.
	txBytes, err := hexutil.Decode(string(txBytesRaw))
	if err != nil {
		return nil, fmt.Errorf("failed to decode tx bytes: %w", err)
	}
	return txBytes, nil
}

// collectDKComponents scans the committee keys directory and builds the DK registries and identities.
func collectDKComponents(committeeKeysDir, committeeAddressStr string) ([]gpa.NodeID, []*cryptolib.KeyPair, []registry.DKShareRegistryProvider, error) {
	stat, err := os.Stat(committeeKeysDir)
	if err != nil || !stat.IsDir() {
		return nil, nil, nil, fmt.Errorf("committee keys must be a directory: %s", committeeKeysDir)
	}
	entries, err := os.ReadDir(committeeKeysDir)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read committee keys directory '%s': %w", committeeKeysDir, err)
	}
	var nodeIDs []gpa.NodeID
	var peerIdentities []*cryptolib.KeyPair
	var dkRegistries []registry.DKShareRegistryProvider
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		nodeID, keyPair, reg, ok := loadNodeComponents(filepath.Join(committeeKeysDir, entry.Name()), committeeAddressStr)
		if !ok {
			continue
		}
		nodeIDs = append(nodeIDs, nodeID)
		peerIdentities = append(peerIdentities, keyPair)
		if reg != nil {
			dkRegistries = append(dkRegistries, reg)
		}
	}
	return nodeIDs, peerIdentities, dkRegistries, nil
}

// loadNodeComponents tries to load identity and dkshare registry from a node subdirectory.
// It returns ok=false if the directory doesn't contain the expected files or keys couldn't be loaded.
func loadNodeComponents(nodeDir, committeeAddressStr string) (gpa.NodeID, *cryptolib.KeyPair, registry.DKShareRegistryProvider, bool) {
	identityPath := filepath.Join(nodeDir, "identity", "identity.key")
	if st, err := os.Stat(identityPath); err != nil || st.IsDir() {
		return gpa.NodeID{}, nil, nil, false
	}
	// dkshare for this committee
	dkSharesDir := filepath.Join(nodeDir, "dkshares")
	dkSharePath := filepath.Join(dkSharesDir, committeeAddressStr+".json")
	if st, err := os.Stat(dkSharePath); err != nil || st.IsDir() {
		return gpa.NodeID{}, nil, nil, false
	}
	// load identity private key
	privKeyRaw, newlyCreated, err := hivep2p.LoadOrCreateIdentityPrivateKey(identityPath, "")
	if err != nil || newlyCreated {
		return gpa.NodeID{}, nil, nil, false
	}
	privKeyBytes, err := privKeyRaw.Raw()
	if err != nil {
		return gpa.NodeID{}, nil, nil, false
	}
	privKey, err := cryptolib.PrivateKeyFromBytes(privKeyBytes)
	if err != nil {
		return gpa.NodeID{}, nil, nil, false
	}
	keyPair := cryptolib.KeyPairFromPrivateKey(privKey)
	nodeID := gpa.NodeIDFromPublicKey(keyPair.GetPublicKey())
	reg, err := registry.NewDKSharesRegistry(dkSharesDir, privKey)
	if err != nil {
		reg = nil
	}
	return nodeID, keyPair, reg, true
}

func Init(rootCmd *cobra.Command) {
	disrecCmd := initDisrecCmd()
	disrecCmd.AddCommand(initSignAndPostCmd())
	rootCmd.AddCommand(disrecCmd)
}
