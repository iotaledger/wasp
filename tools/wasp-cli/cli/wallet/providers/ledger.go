package providers

import (
	"github.com/iotaledger/wasp/v2/clients/iota-go/hw_ledger"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

type ExternalWallet struct {
	ledger       *hw_ledger.HWLedger
	addressIndex uint32
	CoinType     iotasigner.Bip32CoinType
	Bip32Path    string
}

func (l *ExternalWallet) Address() *cryptolib.Address {
	bip32Path, err := iotasigner.BuildBip32Path(iotasigner.SignatureFlagEd25519, l.CoinType, l.addressIndex)
	log.Check(err)

	pubKey, err := l.ledger.GetPublicKey(bip32Path, false)
	log.Check(err)

	addr, err := cryptolib.NewAddressFromBytes(pubKey.Address[:])
	log.Check(err)

	return addr
}

func (l *ExternalWallet) Sign(msg []byte) (signature *cryptolib.Signature, err error) {
	bip32Path, err := iotasigner.BuildBip32Path(iotasigner.SignatureFlagEd25519, l.CoinType, l.addressIndex)
	if err != nil {
		return nil, err
	}

	pubKey, err := l.ledger.GetPublicKey(bip32Path, false)
	if err != nil {
		return nil, err
	}

	signed, err := l.ledger.SignTransaction(bip32Path, msg)
	if err != nil {
		return nil, err
	}

	pubKeyIn, err := cryptolib.PublicKeyFromBytes(pubKey.PublicKey[:])
	if err != nil {
		return nil, err
	}

	return cryptolib.NewSignature(pubKeyIn, signed.Signature[:]), nil
}

func (l *ExternalWallet) SignTransactionBlock(txnBytes []byte, intent iotasigner.Intent) (*cryptolib.Signature, error) {
	return l.Sign(iotasigner.MessageWithIntent(intent, txnBytes))
}

func NewExternalWallet(ledger *hw_ledger.HWLedger, addressIndex uint32, coinType iotasigner.Bip32CoinType) *ExternalWallet {
	return &ExternalWallet{
		ledger:       ledger,
		addressIndex: addressIndex,
		CoinType:     coinType,
	}
}

func (l *ExternalWallet) IsNil() bool {
	return l == nil
}

func (l *ExternalWallet) AddressIndex() uint32 {
	return l.addressIndex
}

func (l *ExternalWallet) AsAddressSigner() iotasigner.Signer {
	bip32Path, err := iotasigner.BuildBip32Path(iotasigner.SignatureFlagEd25519, l.CoinType, l.addressIndex)
	log.Check(err)

	signer := hw_ledger.NewLedgerSigner(l.ledger, bip32Path, false)

	return signer
}
