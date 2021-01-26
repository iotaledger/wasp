// +build ignore

// DonateWithFeedback is a smart contract which handles donation account and log of feedback messages
// sent together with the donations
package donatewithfeedback

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/util"
	"io"
	"time"
)

// main external constants
var (
	RequestDonate   = coretypes.Hn("donate")
	RequestWithdraw = coretypes.Hn("withdraw")
)

const (

	// state vars
	// name of the feedback message log
	VarStateTheLog = "l"
	// largest donation so far
	VarStateMaxDonation = "maxd"
	// total donation so far
	VarStateTotalDonations = "total"

	// request arguments
	// variable containing feedback text
	VarReqFeedback = "f"
	// sum to withdraw with the 'withdraw' request
	VarReqWithdrawSum = "s"
)

// DonationInfo is a structure which contains one donation
// it is marshalled to the deterministic binary form and saves as one entry in the state
type DonationInfo struct {
	Seq      int64
	Id       coretypes.RequestID
	When     time.Time // not marshaled, filled in from timestamp
	Amount   int64
	Sender   coretypes.AgentID
	Feedback string // max 16 bit length
	Error    string
}

// serde of the DonationInfo

func (di *DonationInfo) Write(w io.Writer) error {
	if err := util.WriteInt64(w, di.Seq); err != nil {
		return err
	}
	if err := di.Id.Write(w); err != nil {
		return err
	}
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
	if err := util.ReadInt64(r, &di.Seq); err != nil {
		return err
	}
	if err := di.Id.Read(r); err != nil {
		return err
	}
	if err = util.ReadInt64(r, &di.Amount); err != nil {
		return err
	}
	if err = coretypes.ReadAgentID(r, &di.Sender); err != nil {
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
