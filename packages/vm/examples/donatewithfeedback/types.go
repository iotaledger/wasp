package donatewithfeedback

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

const (
	RequestDonate  = sctransaction.RequestCode(uint16(1))
	RequestHarvest = sctransaction.RequestCode(uint16(2) | sctransaction.RequestCodeProtected)

	// state vars
	VarStateTheLog = "l"

	// request vars
	VarReqFeedback   = "f"
	VarReqHarvestSum = "s"
)

type DonationInfo struct {
	Amount   int64
	Sender   address.Address
	Feedback string // max 16 bit length
	Error    string
}

func (di *DonationInfo) Write(w io.Writer) error {
	if err := util.WriteInt64(w, di.Amount); err != nil {
		return err
	}
	if _, err := w.Write(di.Sender[:]); err != nil {
		return err
	}
	if err := util.WriteString16(w, di.Feedback); err != nil {
		return err
	}
	if err := util.WriteString16(w, di.Error); err != nil {
		return err
	}
	return nil
}

func (di *DonationInfo) Read(r io.Reader) error {
	var err error
	if err = util.ReadInt64(r, &di.Amount); err != nil {
		return err
	}
	if err = util.ReadAddress(r, &di.Sender); err != nil {
		return err
	}
	if di.Feedback, err = util.ReadString16(r); err != nil {
		return err
	}
	if di.Error, err = util.ReadString16(r); err != nil {
		return err
	}
	return nil
}

func (di *DonationInfo) Bytes() []byte {
	return util.MustBytes(di)
}

func DonationInfoFromBytes(data []byte) (*DonationInfo, error) {
	ret := &DonationInfo{}
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}
