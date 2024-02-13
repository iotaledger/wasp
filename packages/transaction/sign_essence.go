package transaction

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

// alternateSignEssence is basically a 1:1 copy of iota.go with the difference that you can inject your own AddressSigner.
// This will ignore passed addressKeys and only use the passed AddressSigner.
// This is important for HW-wallets where the private key is unknown.
func alternateSignEssence(essence *iotago.TransactionEssence, inputsCommitment []byte, signer iotago.AddressSigner, addrKeys ...iotago.AddressKeys) ([]iotago.Signature, error) {
	// SignBytes produces signatures signing the essence for every given AddressKeys.
	// The produced signatures are in the same order as the AddressKeys.
	if inputsCommitment == nil || len(inputsCommitment) != iotago.InputsCommitmentLength {
		return nil, iotago.ErrInvalidInputsCommitment
	}

	copy(essence.InputsCommitment[:], inputsCommitment)

	signMsg, err := essence.SigningMessage()
	if err != nil {
		return nil, err
	}

	sigs := make([]iotago.Signature, len(addrKeys))

	if signer == nil {
		signer = iotago.NewInMemoryAddressSigner(addrKeys...)
	}

	for i, v := range addrKeys {
		sig, err := signer.Sign(v.Address, signMsg)
		if err != nil {
			return nil, err
		}
		sigs[i] = sig
	}

	return sigs, nil
}

func SignEssence(essence *iotago.TransactionEssence, inputsCommitment []byte, keyPair cryptolib.VariantKeyPair) ([]iotago.Signature, error) {
	signer := keyPair.AsAddressSigner()
	addressKeys := keyPair.AddressKeysForEd25519Address(keyPair.Address())

	return alternateSignEssence(essence, inputsCommitment, signer, addressKeys)
}
