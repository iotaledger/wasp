package main

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/iotaledger/wasp/sui-go/examples/swap/pkg"
	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/move"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui/conn"
	"github.com/iotaledger/wasp/sui-go/sui_signer"
)

//go:generate sh -c "cd ../swap && sui move build --dump-bytecode-as-base64 > bytecode.json"
//go:embed swap/bytecode.json
var swapBytecodeJSON []byte

func main() {
	suiClient, signer := sui.NewSuiClient(conn.LocalnetEndpointUrl).WithSignerAndFund(sui_signer.TEST_SEED, 0)
	_, swapper := sui.NewSuiClient(conn.LocalnetEndpointUrl).WithSignerAndFund(sui_signer.TEST_SEED, 1)
	fmt.Println("signer: ", signer.Address())
	fmt.Println("swapper: ", swapper.Address())

	swapPackageID := pkg.Publish(suiClient, signer, move.DecodePackageBytecode(swapBytecodeJSON))
	testcoinID, _ := pkg.PublishMintTestcoin(suiClient, signer)
	testcoinCoinType := fmt.Sprintf("%s::testcoin::TESTCOIN", testcoinID.String())

	fmt.Println("swapPackageID: ", swapPackageID)
	fmt.Println("testcoinCoinType: ", testcoinCoinType)

	testcoinCoins, err := suiClient.GetCoins(
		context.Background(),
		&models.GetCoinsRequest{
			Owner:    signer.Address(),
			CoinType: &testcoinCoinType,
		},
	)
	if err != nil {
		panic(err)
	}

	signerSuiCoinPage, err := suiClient.GetCoins(
		context.Background(),
		&models.GetCoinsRequest{Owner: signer.Address()},
	)
	if err != nil {
		panic(err)
	}

	poolObjectID := pkg.CreatePool(
		suiClient,
		signer,
		swapPackageID,
		testcoinID,
		testcoinCoins.Data[0],
		signerSuiCoinPage.Data,
	)

	swapperSuiCoinPage1, err := suiClient.GetAllCoins(
		context.Background(),
		&models.GetAllCoinsRequest{Owner: swapper.Address()},
	)
	if err != nil {
		panic(err)
	}

	pkg.SwapSui(suiClient, swapper, swapPackageID, testcoinID, poolObjectID, swapperSuiCoinPage1.Data)

	swapperSuiCoinPage2, err := suiClient.GetAllCoins(
		context.Background(),
		&models.GetAllCoinsRequest{Owner: swapper.Address()},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("swapper now has")
	for _, coin := range swapperSuiCoinPage2.Data {
		fmt.Printf("object: %s in type: %s\n", coin.CoinObjectID, coin.CoinType)
	}
}
