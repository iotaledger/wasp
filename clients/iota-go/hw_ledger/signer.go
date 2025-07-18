package hw_ledger

import (
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
)

type Signer struct {
	device                      *HWLedger
	bip32Path                   string
	askForPublicKeyConfirmation bool
}

var _ iotasigner.Signer = &Signer{}

func NewLedgerSigner(device *HWLedger, bip32Path string, askForPublicKeyConfirmation bool) *Signer {
	return &Signer{
		device:                      device,
		bip32Path:                   bip32Path,
		askForPublicKeyConfirmation: askForPublicKeyConfirmation,
	}
}

func (s *Signer) Address() *iotago.Address {
	pubKey, err := s.device.GetPublicKey(s.bip32Path, s.askForPublicKeyConfirmation)
	if err != nil {
		panic(err)
	}

	return iotago.AddressFromArray(pubKey.Address)
}

func (s *Signer) Sign(msg []byte) (signature *iotasigner.Signature, err error) {
	pubKey, err := s.device.GetPublicKey(s.bip32Path, s.askForPublicKeyConfirmation)
	if err != nil {
		return nil, err
	}

	signedBytes, err := s.device.SignTransaction(s.bip32Path, msg)
	if err != nil {
		return nil, err
	}

	return &iotasigner.Signature{
		Ed25519Signature: iotasigner.NewEd25519Signature(
			pubKey.PublicKey[:],
			signedBytes.Signature[:],
		),
	}, nil
}

func (s *Signer) SignTransactionBlock(txnBytes []byte, intent iotasigner.Intent) (*iotasigner.Signature, error) {
	pubKey, err := s.device.GetPublicKey(s.bip32Path, s.askForPublicKeyConfirmation)
	if err != nil {
		return nil, err
	}

	signedBytes, err := s.device.SignTransaction(s.bip32Path, iotasigner.MessageWithIntent(intent, txnBytes))
	if err != nil {
		return nil, err
	}

	return &iotasigner.Signature{
		Ed25519Signature: iotasigner.NewEd25519Signature(
			pubKey.PublicKey[:],
			signedBytes.Signature[:],
		),
	}, nil
}
