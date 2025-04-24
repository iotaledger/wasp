package iotasigner_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	testcommon "github.com/iotaledger/wasp/clients/iota-go/test_common"
)

func TestNewSigner(t *testing.T) {
	signer, err := iotasigner.NewSignerWithMnemonic(testcommon.TestMnemonic, iotasigner.KeySchemeFlagDefault)
	require.NoError(t, err)
	require.Equal(t, iotago.MustAddressFromHex(testcommon.TestAddress), signer.Address())
}

func TestSignatureMarshalUnmarshal(t *testing.T) {
	signer, err := iotasigner.NewSignerWithMnemonic(testcommon.TestMnemonic, iotasigner.KeySchemeFlagDefault)
	require.NoError(t, err)

	msg := "I want to have some bubble tea"
	msgBytes := []byte(msg)

	signature1, err := signer.SignTransactionBlock(msgBytes, iotasigner.DefaultIntent())
	require.NoError(t, err)

	marshaledData, err := json.Marshal(signature1)
	require.NoError(t, err)

	var signature2 *iotasigner.Signature
	err = json.Unmarshal(marshaledData, &signature2)
	require.NoError(t, err)

	require.Equal(t, signature1, signature2)
}

func ExampleSigner() {
	// Create a iotasigner.InMemorySigner with mnemonic
	signer1, _ := iotasigner.NewSignerWithMnemonic(testcommon.TestMnemonic, iotasigner.KeySchemeFlagDefault)
	fmt.Printf("address   : %v\n", signer1.Address())

	// Create iotasigner.InMemorySigner with private key
	privKey, _ := hex.DecodeString("4ec5a9eefc0bb86027a6f3ba718793c813505acc25ed09447caf6a069accdd4b")
	signer2 := iotasigner.NewSigner(privKey, iotasigner.KeySchemeFlagDefault)

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
