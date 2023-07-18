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

func NewExternalWallet(secretManager *walletsdk.SecretManager, addressIndex uint32, bech32Hrp string, coinType types.CoinType) *ExternalWallet {
	return &ExternalWallet{
		secretManager: secretManager,
		addressIndex:  addressIndex,
		Bech32Hrp:     bech32Hrp,
		CoinType:      coinType,
	}
}

func (l *ExternalWallet) AddressIndex() uint32 {
	return l.addressIndex
}

func (l *ExternalWallet) Sign(addr iotago.Address, payload []byte) (signature iotago.Signature, err error) {
	bip32Chain := walletsdk.BuildBip44Chain(l.CoinType, 0, l.addressIndex)
	signResult, err := l.secretManager.SignTransactionEssence(types.HexEncodedString(iotago.EncodeHex(payload)), bip32Chain)
	if err != nil {
		return nil, err
	}

	return SDKED25519SignatureToIOTAGo(signResult)
}

func (l *ExternalWallet) SignBytes(payload []byte) []byte {
	bip32Chain := walletsdk.BuildBip44Chain(l.CoinType, 0, l.addressIndex)
	signResult, err := l.secretManager.SignTransactionEssence(types.HexEncodedString(iotago.EncodeHex(payload)), bip32Chain)
	log.Check(err)

	signature, err := iotago.DecodeHex(signResult.Signature)
	log.Check(err)

	return signature
}

func (l *ExternalWallet) GetPublicKey() *cryptolib.PublicKey {
	// To get the public key, it's required to sign some data first.
	payload := make([]byte, 32)
	signedPayload, err := l.Sign(nil, payload)
	log.Check(err)

	ed25519Signature, ok := signedPayload.(*iotago.Ed25519Signature)
	if !ok {
		log.Fatalf("signed payload is not an ED25519 signature")
	}

	publicKey, err := cryptolib.PublicKeyFromBytes(ed25519Signature.PublicKey[:])
	log.Check(err)

	return publicKey
}

func (l *ExternalWallet) Address() *iotago.Ed25519Address {
	addressStr, err := l.secretManager.GenerateEd25519Address(l.addressIndex, 0, l.Bech32Hrp, l.CoinType, &types.IGenerateAddressOptions{
		Internal:         false,
		LedgerNanoPrompt: false,
	})
	log.Check(err)

	_, address, err := iotago.ParseBech32(addressStr)
	log.Check(err)

	return address.(*iotago.Ed25519Address)
}

func (l *ExternalWallet) AsAddressSigner() iotago.AddressSigner {
	return &ExternalAddressSigner{
		keyPair: l,
	}
}

func (l *ExternalWallet) AddressKeysForEd25519Address(addr *iotago.Ed25519Address) iotago.AddressKeys {
	// Not required, as we override the address signer above, and address keys are not used there.
	return iotago.AddressKeys{}
}
