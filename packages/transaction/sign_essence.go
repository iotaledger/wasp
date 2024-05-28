package transaction

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

// alternateSignEssence is basically a 1:1 copy of iota.go with the difference that you can inject your own AddressSigner.
// This will ignore passed addressKeys and only use the passed AddressSigner.
// This is important for HW-wallets where the private key is unknown.
func alternateSignEssence(essence *iotago.TransactionEssence, inputsCommitment []byte, signers ...cryptolib.Signer) ([]*cryptolib.Signature, error) {
	// SignBytes produces signatures signing the essence for every given AddressKeys.
	// The produced signatures are in the same order as the AddressKeys.
	if inputsCommitment == nil {
		return nil, fmt.Errorf("invalid inputs commitment: nil")
	}
	if len(inputsCommitment) != iotago.InputsCommitmentLength {
		return nil, fmt.Errorf("invalid inputs commitment: expected %v, got %v", iotago.InputsCommitmentLength, len(inputsCommitment))
	}

	copy(essence.InputsCommitment[:], inputsCommitment)

	signMsg, err := essence.SigningMessage()
	if err != nil {
		return nil, err
	}

	sigs := make([]*cryptolib.Signature, len(signers))

	for i, signer := range signers {
		sig, err := signer.Sign(signMsg)
		if err != nil {
			return nil, err
		}
		sigs[i] = sig
	}

	return sigs, nil
}

func SignEssence(essence *iotago.TransactionEssence, inputsCommitment []byte, signer cryptolib.Signer) ([]*cryptolib.Signature, error) {
	//signer := keyPair.AsAddressSigner()
	//addressKeys := keyPair.AddressKeysForEd25519Address(keyPair.Address())
	return alternateSignEssence(essence, inputsCommitment, signer)
}
