package wallets

import (
	iotago "github.com/iotaledger/iota.go/v3"
	walletsdk "github.com/iotaledger/wasp-wallet-sdk"
	"github.com/iotaledger/wasp-wallet-sdk/types"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

type ExternalWallet struct {
	secretManager *walletsdk.SecretManager

	addressIndex uint32
	Bech32Hrp    string
	CoinType     types.CoinType
}

var (
	_ cryptolib.Signer = &ExternalWallet{}
	_ Wallet           = &ExternalWallet{}
)

func NewExternalWallet(secretManager *walletsdk.SecretManager, addressIndex uint32, bech32Hrp string, coinType types.CoinType) *ExternalWallet {
	return &ExternalWallet{
		secretManager: secretManager,
		addressIndex:  addressIndex,
		Bech32Hrp:     bech32Hrp,
		CoinType:      coinType,
	}
}

func (l *ExternalWallet) IsNil() bool {
	return l == nil
}

func (l *ExternalWallet) AddressIndex() uint32 {
	return l.addressIndex
}

func (l *ExternalWallet) Sign(payload []byte) (*cryptolib.Signature, error) {
	publicKey, signature, err := l.sign(payload)
	if err != nil {
		return nil, err
	}
	return cryptolib.NewSignature(publicKey, signature), nil
}

func (l *ExternalWallet) SignBytes(payload []byte) []byte {
	signature, _ := l.signBytes(payload)
	return signature
}

func (l *ExternalWallet) sign(payload []byte) (*cryptolib.PublicKey, []byte, error) {
	signature, signResult := l.signBytes(payload)
	publicKeyBytes, err := cryptolib.DecodeHex(signResult.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	publicKey, err := cryptolib.PublicKeyFromBytes(publicKeyBytes)
	if err != nil {
		return nil, nil, err
	}
	return publicKey, signature, nil
}

func (l *ExternalWallet) signBytes(payload []byte) ([]byte, *types.Ed25519Signature) {
	bip32Chain := walletsdk.BuildBip44Chain(l.CoinType, 0, l.addressIndex)
	signResult, err := l.secretManager.SignTransactionEssence(types.HexEncodedString(iotago.EncodeHex(payload)), bip32Chain)
	log.Check(err)

	signature, err := iotago.DecodeHex(signResult.Signature)
	log.Check(err)

	return signature, signResult
}

func (l *ExternalWallet) GetPublicKey() *cryptolib.PublicKey {
	// To get the public key, it's required to sign some data first.
	payload := make([]byte, 32)
	publicKey, _, err := l.sign(payload)
	log.Check(err)
	return publicKey
}

func (l *ExternalWallet) Address() *cryptolib.Address {
	addressStr, err := l.secretManager.GenerateEd25519Address(l.addressIndex, 0, l.Bech32Hrp, l.CoinType, &types.IGenerateAddressOptions{
		Internal:         false,
		LedgerNanoPrompt: false,
	})
	log.Check(err)

	_, address, err := cryptolib.NewAddressFromBech32(addressStr)
	log.Check(err)

	return address
}

/*func (l *ExternalWallet) AsAddressSigner() iotago.AddressSigner {
	return &ExternalAddressSigner{
		keyPair: l,
	}
}

func (l *ExternalWallet) AddressKeysForEd25519Address(addr *iotago.Ed25519Address) iotago.AddressKeys {
	// Not required, as we override the address signer above, and address keys are not used there.
	return iotago.AddressKeys{}
}*/
