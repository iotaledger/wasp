package disrec

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/tools/wasp-cli/chain"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/providers"
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
