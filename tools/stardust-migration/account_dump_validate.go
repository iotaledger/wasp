package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/maps"

	"github.com/iotaledger/iota.go/v3/bech32"
	"github.com/iotaledger/wasp/packages/webapi/controllers/chain/models"
	"github.com/iotaledger/wasp/tools/stardust-migration/webapi-validation/base"
)

/**
{
  "0xFffF309103Ad05aA9819A0c31f28F3e4F041D602@iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5": {
    "baseTokens": "0",
    "nativeTokens": [],
    "nfts": []
  },
  "iota1qqsrjywudtz5t4at7tjaguh3s8nnzp89vak9fvzjnpasdasdasd": {
    "baseTokens": "216533",
    "nativeTokens": [
      {
        "id": "0x083c55be0f034673cef16a7553f42a7928d998ccc1e970968ea0965608de2c6a440100000000",
        "amount": "100000"
      }
    ],
    "nfts": [
      "0x0a687167ca954498390cb23a78a3d422d8123123123"
    ]
  },
}
*/

type StardustNativeToken struct {
	ID     string `json:"id"`
	Amount string `json:"amount"`
}

type StardustAccount struct {
	BaseTokens   string                `json:"baseTokens"`
	NativeTokens []StardustNativeToken `json:"nativeTokens"`
	NFTs         []string              `json:"nfts"`
}

type StardustAccountDump = map[string]StardustAccount
type RebasedAccountDump = models.DumpAccountsResponse

var t = &base.MockT{}

func validateAccountDumps(c *cli.Context) error {
	stardustEndPoint := c.Args().Get(0)
	rebasedEndpoint := c.Args().Get(1)

	stardustDump, err := os.ReadFile(stardustEndPoint)
	if err != nil {
		return err
	}

	rebasedAccountDump, err := os.ReadFile(rebasedEndpoint)
	if err != nil {
		return err
	}

	stardustPreConversion := StardustAccountDump{}
	err = json.Unmarshal(stardustDump, &stardustPreConversion)
	if err != nil {
		return err
	}

	stardust := StardustAccountDump{}
	for k, v := range stardustPreConversion {
		// Convert addresses to the new format

		if strings.Contains(k, "@") {
			address := strings.Split(k, "@")[0]

			if !strings.HasPrefix(k, "0x") {
				address = "0x" + address
			}

			stardust[address] = v
		} else {
			_, bytes, err := bech32.Decode(k)
			if err != nil {
				return fmt.Errorf("Failed to decode bech32: %w", err)
			}

			newAddress := hexutil.Encode(bytes[1:])
			stardust[newAddress] = v
		}
	}

	rebased := RebasedAccountDump{}
	err = json.Unmarshal(rebasedAccountDump, &rebased)
	if err != nil {
		return err
	}

	validateAccountIntegrity(stardust, rebased)
	validateAccountAssetsEqual(stardust, rebased)

	return nil
}

// validateAccountIntegrity validates that both dumps have the same length and the same account keys.
func validateAccountIntegrity(stardust StardustAccountDump, rebased RebasedAccountDump) {
	require.Equal(t, len(rebased.Accounts), len(stardust))

	rebasedKeys := maps.Keys(rebased.Accounts)
	stardustKeys := maps.Keys(stardust)

	sort.Strings(rebasedKeys)
	sort.Strings(stardustKeys)

	require.Equal(t, rebasedKeys, stardustKeys)
}

func validateAccountAssetsEqual(stardust StardustAccountDump, rebased RebasedAccountDump) {
	validatedAccounts := 0

	for address, stardustAssets := range stardust {
		// This is a past crosschain request exception
		// The legacy migration chain has a 180i deposit on our chain which will be consumed in to our chains account.
		if address == "0x05204969" {
			continue
		}

		rebasedAssets := rebased.Accounts[address]

		// Again, here we need to adjust for the decimal change = 6 => 9 decimals

		stardustBaseToken, err := strconv.ParseUint(stardustAssets.BaseTokens, 10, 64)
		require.NoError(t, err)

		fmt.Printf("%s: stardust:%d, rebased:%d \n", address, stardustBaseToken, rebasedAssets.Coins.BaseTokens().Uint64())

		// The account dumper only dumps the balances as uint64. Internally we use BigInts, which can cause a mismatch in the conversion validation.
		// The stardust balance could be reported as 0, even though it's something like 0.001. The conversion would make this appear greater than 0. Causing validation to fail

		if stardustBaseToken > 10 {
			require.Equal(t, len(stardustAssets.BaseTokens)+3, len(rebasedAssets.Coins.BaseTokens().String()), "Mismatch between baseTokens for address: %s => Stardust: %d, Rebased: %d", address, stardustBaseToken, rebasedAssets.Coins.BaseTokens().Uint64())
			require.True(t, strings.HasPrefix(rebasedAssets.Coins.BaseTokens().String(), stardustAssets.BaseTokens))
			validatedAccounts++
		} else {
			//	require.Equal(t, len(stardustAssets.BaseTokens), len(rebasedAssets.Coins.BaseTokens().String()))
		}

	}

	fmt.Printf("Validated accounts: %d from %d\n", validatedAccounts, len(rebased.Accounts))

}
