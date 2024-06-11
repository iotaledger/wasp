package wallets

// TODO: remove it
/*import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp-wallet-sdk/types"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

func SDKED25519SignatureToIOTAGo(responseSignature *types.Ed25519Signature) (*cryptolib.Signature, error) {
	signatureBytes, err := cryptolib.DecodeHex(responseSignature.Signature)
	if err != nil {
		return nil, err
	}

	publicKeyBytes, err := cryptolib.DecodeHex(responseSignature.PublicKey)
	if err != nil {
		return nil, err
	}

	signature := cryptolib.NewSignature()
	copy(signature.Signature[:], signatureBytes)
	copy(signature.PublicKey[:], publicKeyBytes)

	return &signature, nil
}*/
