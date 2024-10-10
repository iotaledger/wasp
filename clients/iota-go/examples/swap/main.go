package main

import (
	"context"
	_ "embed"
	"fmt"

	pkg2 "github.com/iotaledger/wasp/clients/iota-go/examples/swap/pkg"
	"github.com/iotaledger/wasp/clients/iota-go/move"
	suiclient2 "github.com/iotaledger/wasp/clients/iota-go/suiclient"
	"github.com/iotaledger/wasp/clients/iota-go/suiconn"
	"github.com/iotaledger/wasp/clients/iota-go/suitest"
)

//go:generate sh -c "cd ../swap && sui move build --dump-bytecode-as-base64 > bytecode.json"
//go:embed swap/bytecode.json
var swapBytecodeJSON []byte

func main() {
	suiClient := suiclient2.NewHTTP(suiconn.LocalnetEndpointURL)
	signer := suitest.MakeSignerWithFunds(0, suiconn.LocalnetFaucetURL)
	swapper := suitest.MakeSignerWithFunds(1, suiconn.LocalnetFaucetURL)

	fmt.Println("signer: ", signer.Address())
	fmt.Println("swapper: ", swapper.Address())

	swapPackageID := pkg2.Publish(suiClient, signer, move.DecodePackageBytecode(swapBytecodeJSON))
	testcoinID, _ := pkg2.PublishMintTestcoin(suiClient, signer)
	testcoinCoinType := fmt.Sprintf("%s::testcoin::TESTCOIN", testcoinID.String())

	fmt.Println("swapPackageID: ", swapPackageID)
	fmt.Println("testcoinCoinType: ", testcoinCoinType)

	testcoinCoins, err := suiClient.GetCoins(
		context.Background(),
		suiclient2.GetCoinsRequest{
			Owner:    signer.Address(),
			CoinType: &testcoinCoinType,
		},
	)
	if err != nil {
		panic(err)
	}

	signerSuiCoinPage, err := suiClient.GetCoins(
		context.Background(),
		suiclient2.GetCoinsRequest{Owner: signer.Address()},
	)
	if err != nil {
		panic(err)
	}

	poolObjectID := pkg2.CreatePool(
		suiClient,
		signer,
		swapPackageID,
		testcoinID,
		testcoinCoins.Data[0],
		signerSuiCoinPage.Data,
	)

	swapperSuiCoinPage1, err := suiClient.GetAllCoins(
		context.Background(),
		suiclient2.GetAllCoinsRequest{Owner: swapper.Address()},
	)
	if err != nil {
		panic(err)
	}

	pkg2.SwapSui(suiClient, swapper, swapPackageID, testcoinID, poolObjectID, swapperSuiCoinPage1.Data)

	swapperSuiCoinPage2, err := suiClient.GetAllCoins(
		context.Background(),
		suiclient2.GetAllCoinsRequest{Owner: swapper.Address()},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("swapper now has")
	for _, coin := range swapperSuiCoinPage2.Data {
		fmt.Printf("object: %s in type: %s\n", coin.CoinObjectID, coin.CoinType)
	}
}
