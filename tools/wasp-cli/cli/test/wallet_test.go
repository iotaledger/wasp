package test

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/require"
	"github.com/tyler-smith/go-bip39"

	iotago "github.com/iotaledger/iota.go/v3"
	wasp_wallet_sdk "github.com/iotaledger/wasp-wallet-sdk"
	"github.com/iotaledger/wasp-wallet-sdk/types"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

func TestMnemonic(t *testing.T) {
	seedBytes, _ := iotago.DecodeHex("0xbc278147b72c6af948eced45252c496901e194c9610bfbffea639e18769c7715")
	seed := cryptolib.SeedFromBytes(seedBytes)
	kp := cryptolib.KeyPairFromSeed(seed)
	address := kp.Address().Bech32("rms")
	require.Equal(t, address, "rms1qzy0uqyzcm6asngsjxwc76nuar479uukvxa4dapzaz8fx5e3234wxw5mlmz")
	fmt.Println(address)

	mnemonic, err := bip39.NewMnemonic(seed[:])
	require.NoError(t, err)

	cwd, err := os.Getwd()
	require.NoError(t, err)

	sdk, err := wasp_wallet_sdk.NewIotaSDK(path.Join(cwd, "../../../../libiota_sdk_native.so"))
	require.NoError(t, err)

	manager, err := wasp_wallet_sdk.NewStrongholdSecretManager(sdk, memguard.NewEnclaveRandom(32), "./test.snapshot")
	defer os.Remove("./test.snapshot")

	require.NoError(t, err)

	mnemonicBytes := []byte(mnemonic)
	success, err := manager.StoreMnemonic(memguard.NewEnclave(mnemonicBytes))
	require.NoError(t, err)
	require.True(t, success)

	strongholdAddress, err := manager.GenerateEd25519Address(0, 0, "rms", types.CoinTypeSMR, nil)
	require.NoError(t, err)
	fmt.Println(strongholdAddress)
}
