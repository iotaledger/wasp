package suisigner_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/sui"
	"github.com/iotaledger/wasp/clients/iota-go/suisigner"
)

var testMnemonic = "ordinary cry margin host traffic bulb start zone mimic wage fossil eight diagram clay say remove add atom"

func TestNewSigner(t *testing.T) {
	testMnemonic := "ordinary cry margin host traffic bulb start zone mimic wage fossil eight diagram clay say remove add atom"
	testIotaAddress := sui.MustAddressFromHex("0x786dff8a4ee13d45b502c8f22f398e3517e6ec78aa4ae564c348acb07fad7f50")
	testEd25519Address := sui.MustAddressFromHex("0xe54d993cf56be93ba0764c7ee2c817085b70f0e6d3ad1a71c3335ee3529b4a48")
	signer, err := suisigner.NewSignerWithMnemonic(testMnemonic, suisigner.KeySchemeFlagIotaEd25519)
	require.NoError(t, err)
	require.Equal(t, testIotaAddress, signer.Address())
	signer, err = suisigner.NewSignerWithMnemonic(testMnemonic, suisigner.KeySchemeFlagEd25519)
	require.NoError(t, err)
	require.Equal(t, testEd25519Address, signer.Address())
}

func TestSignatureMarshalUnmarshal(t *testing.T) {
	signer, err := suisigner.NewSignerWithMnemonic(testMnemonic, suisigner.KeySchemeFlagDefault)
	require.NoError(t, err)

	msg := "I want to have some bubble tea"
	msgBytes := []byte(msg)

	signature1, err := signer.SignTransactionBlock(msgBytes, suisigner.DefaultIntent())
	require.NoError(t, err)

	marshaledData, err := json.Marshal(signature1)
	require.NoError(t, err)

	var signature2 *suisigner.Signature
	err = json.Unmarshal(marshaledData, &signature2)
	require.NoError(t, err)

	require.Equal(t, signature1, signature2)
}

func ExampleSigner() {
	// Create a suisigner.InMemorySigner with mnemonic
	mnemonic := "ordinary cry margin host traffic bulb start zone mimic wage fossil eight diagram clay say remove add atom"
	signer1, _ := suisigner.NewSignerWithMnemonic(mnemonic, suisigner.KeySchemeFlagDefault)
	fmt.Printf("address   : %v\n", signer1.Address())

	// Create suisigner.InMemorySigner with private key
	privKey, _ := hex.DecodeString("4ec5a9eefc0bb86027a6f3ba718793c813505acc25ed09447caf6a069accdd4b")
	signer2 := suisigner.NewSigner(privKey, suisigner.KeySchemeFlagDefault)

	// Get private key, public key, address
	fmt.Printf("privateKey: %x\n", signer2.PrivateKey()[:32])
	fmt.Printf("publicKey : %x\n", signer2.PublicKey())
	fmt.Printf("address   : %v\n", signer2.Address())

	// Output:
	// address   : 0x786dff8a4ee13d45b502c8f22f398e3517e6ec78aa4ae564c348acb07fad7f50
	// privateKey: 4ec5a9eefc0bb86027a6f3ba718793c813505acc25ed09447caf6a069accdd4b
	// publicKey : 9342fa65507f5cf61f1b8fb3b94a5aa80fa9b2e2c68963e30d68a2660a50c57e
	// address   : 0x07e542f628f0e48950578aaff3e0c0566b6dccfc7cc248d9941308b47e934e6a
}
