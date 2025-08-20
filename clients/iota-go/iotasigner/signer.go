package iotasigner

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/tyler-smith/go-bip39"
	"github.com/wollac/iota-crypto-demo/pkg/bip32path"
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

type Bip32CoinType int

const (
	// testnet/alphanet uses COIN_TYPE = 1
	TestnetCoinType Bip32CoinType = 1
	// IOTA coin type <https://github.com/satoshilabs/slips/blob/master/slip-0044.md>
	IotaCoinType Bip32CoinType = 4218
)

type SignatureFlag int

const (
	SignatureFlagEd25519   SignatureFlag = 0x0
	SignatureFlagSecp256k1 SignatureFlag = 0x1
)

func BuildBip32Path(signatureFlag SignatureFlag, coinType Bip32CoinType, accountIndex uint32) (string, error) {
	var derivationSignature int
	switch signatureFlag {
	case SignatureFlagEd25519:
		derivationSignature = 44

	case SignatureFlagSecp256k1:
		derivationSignature = 54
	default:
		return "", errors.New("invalid signature flag")
	}

	bip32Path := fmt.Sprintf(
		"%d'/%d'/%d'/0'/0'",
		derivationSignature,
		coinType,
		accountIndex,
	)

	_, err := bip32path.ParsePath(bip32Path)

	return bip32Path, err
}

type Signer interface {
	Address() *iotago.Address
	Sign(msg []byte) (signature *Signature, err error)
	SignTransactionBlock(txnBytes []byte, intent Intent) (*Signature, error)
}

type InMemorySigner struct {
	ed25519Keypair *KeypairEd25519

	address *iotago.Address
}

func NewSigner(seed []byte, flag KeySchemeFlag) *InMemorySigner {
	prikey := ed25519.NewKeyFromSeed(seed)
	pubkey := prikey.Public().(ed25519.PublicKey)

	// Right now only supports ED25519. Stardust *Addresses* used to include the scheme flag, Rebased does not.
	// https://docs.iota.org/developer/cryptography/transaction-auth/keys-addresses#address-format

	if flag != KeySchemeFlagEd25519 {
		panic("unrecognizable key scheme flag")
	}

	addrBytes := blake2b.Sum256(pubkey)
	addr := "0x" + hex.EncodeToString(addrBytes[:])

	return &InMemorySigner{
		ed25519Keypair: &KeypairEd25519{
			PriKey: prikey,
			PubKey: pubkey,
		},
		address: iotago.MustAddressFromHex(addr),
	}
}

// there are only 256 different signers can be generated
func NewSignerByIndex(seed []byte, flag KeySchemeFlag, index int) Signer {
	seed[0] += byte(index)
	return NewSigner(seed, flag)
}

// generate keypair (signer) with mnemonic which is referring the Iota monorepo in the following code
//
// let phrase = "asset pink record dawn hundred sure various crime client enforce carbon blossom";
// let mut keystore = Keystore::from(InMemKeystore::new_insecure_for_tests(0));
// let generated_address = keystore.import_from_mnemonic(&phrase, SignatureScheme::ED25519, None, None).unwrap();
func NewSignerWithMnemonic(mnemonic string, flag KeySchemeFlag) (Signer, error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return nil, err
	}
	bip32, err := BuildBip32Path(SignatureFlagEd25519, IotaCoinType, 0)
	if err != nil {
		return nil, err
	}
	key, err := DeriveForPath("m/"+bip32, seed)
	if err != nil {
		return nil, err
	}
	return NewSigner(key.Key, flag), nil
}

func (s *InMemorySigner) PrivateKey() []byte {
	switch {
	case s.ed25519Keypair != nil:
		return s.ed25519Keypair.PriKey
	default:
		return nil
	}
}

func (s *InMemorySigner) PublicKey() []byte {
	switch {
	case s.ed25519Keypair != nil:
		return s.ed25519Keypair.PubKey
	default:
		return nil
	}
}

func (s *InMemorySigner) Sign(msg []byte) (signature *Signature, err error) {
	sig := ed25519.Sign(s.ed25519Keypair.PriKey, msg)

	return &Signature{
		Ed25519Signature: NewEd25519Signature(s.ed25519Keypair.PubKey, sig),
	}, nil
}

func (s *InMemorySigner) Address() *iotago.Address {
	return s.address
}

func (s *InMemorySigner) SignTransactionBlock(txnBytes []byte, intent Intent) (*Signature, error) {
	data := MessageWithIntent(intent, txnBytes)
	hash := blake2b.Sum256(data)
	return s.Sign(hash[:])
}
