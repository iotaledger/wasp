package disrec

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/parameters/parameterstest"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/chain"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/wallet/providers"
)

/*
	This test exists to generate unsigned transactions to test the recovery. The transaction creates a new AssetsBag.
	In a real recovery case, we need to know the Committee Address and the GasCoinID used by the Committee.

	In this testing case, the test expects to be run against the Alphanet, which gets reset from time to time.
	This means that the GasCoinID can suddenly vanish.
	As we only need the GasCoin for sending the transaction, the test will make a new one.

	After the transaction has been generated, store the hex string into a file, then you can execute the recovery by:
	`./wasp-cli disrec sign_post testtx.hex "0x4c7fb31a460907210c3b7cbaa50cf9faa23f60cbfbe5f26efd27809265458894" isc-private/tools/wasp-cli/disrec/test_committee_keys/ 	wss://api.iota-rebased-alphanet.iota.cafe`
														^ <-- The Committee Address								     ^ <--- The path to the extracted committee keys       		^ <-- The Websocket address to L1.
*/

func TestDepositFundsToGasCoin(t *testing.T) {
	t.Skip("you only want to call this test manually")

	client := cliclients.L1Client()
	committeeAddress := lo.Must(cryptolib.AddressFromHex("0x6e6d126fc61cbf50672f1738580c7b275e7c4727912842d71ee33e195f9879fe"))
	gasCoinID := lo.Must(iotago.ObjectIDFromHex("0x9e274660552ed50402c8015c5388478415cde8a06d114af48fd2e3ec365c562d"))

	gasCoin, err := client.GetObject(context.Background(), iotaclient.GetObjectRequest{ObjectID: gasCoinID})
	require.NoError(t, err)

	gasCoinRef := gasCoin.Data.Ref()

	kp := cryptolib.NewKeyPair()
	wallet := providers.NewUnsafeInMemoryTestingSeed(kp, 0)
	require.NoError(t, client.RequestFunds(context.Background(), *kp.Address()))
	require.NoError(t, client.RequestFunds(context.Background(), *kp.Address()))

	baseCoin := coin.BaseTokenType.String()
	coins, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{
		CoinType: &baseCoin,
		Owner:    kp.Address().AsIotaAddress(),
	})
	require.NoError(t, err)

	res, err := client.TransferIota(context.Background(), iotaclient.TransferIotaRequest{
		Signer:    kp.Address().AsIotaAddress(),
		GasBudget: iotajsonrpc.NewBigIntInt64(iotaclient.DefaultGasBudget),
		Recipient: committeeAddress.AsIotaAddress(),
		ObjectID:  coins.Data[1].CoinObjectID,
	})
	require.NoError(t, err)

	response, err := client.SignAndExecuteTransaction(context.Background(), &iotaclient.SignAndExecuteTransactionRequest{
		Signer:      cryptolib.SignerToIotaSigner(wallet),
		TxDataBytes: res.TxBytes,
		Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
			ShowObjectChanges: true,
			ShowEffects:       true,
		},
	})
	require.NoError(t, err)

	res2, err := client.TransferIota(context.Background(), iotaclient.TransferIotaRequest{
		Signer:    kp.Address().AsIotaAddress(),
		GasBudget: iotajsonrpc.NewBigIntInt64(iotaclient.DefaultGasBudget),
		Recipient: committeeAddress.AsIotaAddress(),
		ObjectID:  coins.Data[0].CoinObjectID,
	})
	require.NoError(t, err)

	response2, err := client.SignAndExecuteTransaction(context.Background(), &iotaclient.SignAndExecuteTransactionRequest{
		Signer:      cryptolib.SignerToIotaSigner(wallet),
		TxDataBytes: res2.TxBytes,
		Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
			ShowObjectChanges: true,
			ShowEffects:       true,
		},
	})
	require.NoError(t, err)

	fmt.Println(response)

	selectedCoinToFillUpGasCoin, err := response.GetMutatedCoinByType("iota", "IOTA")
	require.NoError(t, err)

	selectedCoinToPayForGas, err := response2.GetMutatedCoinByType("iota", "IOTA")
	require.NoError(t, err)

	ptb := iotago.NewProgrammableTransactionBuilder()

	_ = ptb.Command(
		iotago.Command{
			MergeCoins: &iotago.ProgrammableMergeCoins{
				Destination: ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: &gasCoinRef}),
				Sources:     []iotago.Argument{ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: selectedCoinToFillUpGasCoin})},
			},
		},
	)

	txData := iotago.NewProgrammable(
		committeeAddress.AsIotaAddress(),
		ptb.Finish(),
		[]*iotago.ObjectRef{selectedCoinToPayForGas},
		iotaclient.DefaultGasBudget,
		parameterstest.L1Mock.Protocol.ReferenceGasPrice.Uint64(),
	)

	txnBytes, err := bcs.Marshal(&txData)
	require.NoError(t, err)

	fmt.Println(hexutil.Encode(txnBytes))
}

func TestCreateTX(t *testing.T) {
	t.Skipf("you only want to call this test manually")

	// It really is the Committee Address (Not the Anchor Object ID aka ChainID as this transaction will be signed by the committee)
	// This specific address is the product of the committee keys in `test_committee_keys`.
	committeeAddress := lo.Must(cryptolib.AddressFromHex("0x4c7fb31a460907210c3b7cbaa50cf9faa23f60cbfbe5f26efd27809265458894"))

	client := cliclients.L1Client()
	kp := cryptolib.NewKeyPair()
	wallet := providers.NewUnsafeInMemoryTestingSeed(kp, 0)
	require.NoError(t, client.RequestFunds(context.Background(), *kp.Address()))
	packageID := lo.Must(client.DeployISCContracts(context.Background(), cryptolib.SignerToIotaSigner(kp)))

	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb = iscmoveclient.PTBAssetsBagNewAndTransfer(ptb, packageID, committeeAddress)

	t.Log("Creating new coin and transfer it to the Committee address")

	l1Params := lo.Must(parameters.FetchLatest(context.Background(), client.IotaClient()))
	newGasCoinAddress := lo.Must(chain.CreateAndSendGasCoin(context.Background(), client, wallet, committeeAddress.AsIotaAddress(), l1Params))

	gasCoin := lo.Must(client.GetObject(context.Background(), iotaclient.GetObjectRequest{
		ObjectID: &newGasCoinAddress,
	})).Data.Ref()

	t.Logf("Gas coin ref: %v\n", gasCoin)

	tx := iotago.NewProgrammable(committeeAddress.AsIotaAddress(), ptb.Finish(), []*iotago.ObjectRef{&gasCoin}, 9999999, 1000)
	txnBytes := lo.Must(bcs.Marshal(&tx))

	t.Logf("Test Transaction hex:\n%s\n", hexutil.Encode(txnBytes))
}
