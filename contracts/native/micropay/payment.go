package micropay

import (
	"bytes"
	"errors"
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util"
)

type Payment struct {
	Ord            uint32
	Amount         uint64
	SignatureShort []byte
}

type BatchPayment struct {
	Payer    iotago.Address
	Provider iotago.Address
	Payments []Payment
}

func NewPayment(ord uint32, amount uint64, targetAddr iotago.Address, payerKeyPair *cryptolib.KeyPair) *Payment {
	panic("TODO implement")
	// payerAddr := ledgerstate.NewED25519Address(payerKeyPair.PublicKey)
	// data := paymentEssence(ord, amount, payerAddr, targetAddr)
	// shortSig := payerKeyPair.PrivateKey.Sign(data)
	// signature := ledgerstate.NewED25519Signature(payerKeyPair.PublicKey, shortSig)
	// if !signature.AddressSignatureValid(payerAddr, data) {
	// 	panic("NewPayment: internal error, signature invalid")
	// }
	// return &Payment{
	// 	Ord:            ord,
	// 	Amount:         amount,
	// 	SignatureShort: shortSig[:],
	// }
}

func NewPaymentFromBytes(data []byte) (*Payment, error) {
	ret := &Payment{}
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}

func (p *Payment) Bytes() []byte {
	ret, err := util.Bytes(p)
	if err != nil {
		panic(err)
	}
	return ret
}

func (p *Payment) Write(w io.Writer) error {
	if err := util.WriteUint32(w, p.Ord); err != nil {
		return err
	}
	if err := util.WriteUint64(w, p.Amount); err != nil {
		return err
	}
	return util.WriteBytes16(w, p.SignatureShort)
}

func (p *Payment) Read(r io.Reader) error {
	if err := util.ReadUint32(r, &p.Ord); err != nil {
		return err
	}
	if err := util.ReadUint64(r, &p.Amount); err != nil {
		return err
	}
	var err error
	if p.SignatureShort, err = util.ReadBytes16(r); err != nil {
		return err
	}
	if len(p.SignatureShort) != cryptolib.SignatureSize {
		return errors.New("wrong public key bytes")
	}
	return nil
}
