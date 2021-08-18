// +build ignore

// hard coded smart contract code implements DonateWithFeedback
package dwfimpl

import (
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/contracts/native/ignore/donatewithfeedback"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
)

// program hash: the ID of the code
const (
	ProgramHash = "5ydEfDeAJZX6dh6Fy7tMoHcDeh42gENeqVDASGWuD64X"
	Description = "DonateWithFeedback, a PoC smart contract"
)

// implementation of 'vmtypes.VMProcessor' and 'vmtypes.VMProcessorEntryPoint' interfaces
type dwfProcessor map[iscp.Hname]dwfEntryPoint

type dwfEntryPoint func(ctx iscp.Sandbox) error

// the processor implementation is a map of entry points: one for each request
var entryPoints = dwfProcessor{
	donatewithfeedback.RequestDonate:   donate,
	donatewithfeedback.RequestWithdraw: withdraw,
}

// point of attachment of hard coded code to the rest of Wasp
func GetProcessor() iscp.VMProcessor {
	return entryPoints
}

// GetEntryPoint implements VMProcessorEntryPoint interfaces. It resolves request code to the
// function
func (v dwfProcessor) GetEntryPoint(code iscp.Hname) (iscp.VMProcessorEntryPoint, bool) {
	f, ok := v[code]
	return f, ok
}

// GetDescription description of the smart contract
func (v dwfProcessor) GetDescription() string {
	return "DonateWithFeedback hard coded smart contract processor"
}

// Run calls the function wrapped into the VMProcessorEntryPoint
func (ep dwfEntryPoint) Call(ctx iscp.Sandbox) (dict.Dict, error) {
	ret := ep(ctx)
	if ret != nil {
		ctx.Event(fmt.Sprintf("error %v", ret))
	}
	return nil, ret
}

// TODO
func (ep dwfEntryPoint) IsView() bool {
	return false
}

// TODO
func (ep dwfEntryPoint) CallView(ctx iscp.SandboxView) (dict.Dict, error) {
	panic("implement me")
}

const maxComment = 150

// donate implements request 'donate'. It takes feedback text from the request
// and adds it into the log of feedback messages
func donate(ctx iscp.Sandbox) error {
	ctx.Event(fmt.Sprintf("DonateWithFeedback: donate"))
	params := ctx.Params()

	// how many iotas are sent by the request.
	// only iotas are considered donation. Other colors are ignored
	donated := ctx.IncomingTransfer().Balance(ledgerstate.ColorIOTA)
	// take feedback text contained in the request
	feedback, ok, err := codec.DecodeString(params.MustGet(donatewithfeedback.VarReqFeedback))
	feedback = util.GentleTruncate(feedback, maxComment)

	stateAccess := ctx.State()
	tlog := collections.NewTimestampedLog(stateAccess, donatewithfeedback.VarStateTheLog)

	sender := ctx.Caller()
	// determine sender of the request
	// create donation info record
	di := &donatewithfeedback.DonationInfo{
		Seq:      int64(tlog.MustLen()),
		Id:       ctx.RequestID(),
		Amount:   donated,
		Sender:   sender,
		Feedback: feedback,
	}
	if err != nil {
		di.Error = err.Error()
	} else {
		if !ok || len(strings.TrimSpace(feedback)) == 0 || donated == 0 {
			// empty feedback message is considered an error
			di.Error = "empty feedback or donated amount = 0. The donated amount has been returned (if any)"
		}
	}
	if len(di.Error) != 0 && donated > 0 {
		// if error occurred, return all donated tokens back to the sender
		// in this case error message will be recorded in the donation record

		// ctx.MoveTokens(sender, ledgerstate.ColorIOTA, donated)
		di.Amount = 0
	}
	// store donation info record in the state (append to the timestamped log)
	tlog.MustAppend(ctx.GetTimestamp(), di.Bytes())

	// save total and maximum donations
	maxd, _, _ := codec.DecodeInt64(stateAccess.MustGet(donatewithfeedback.VarStateMaxDonation))
	total, _, _ := codec.DecodeInt64(stateAccess.MustGet(donatewithfeedback.VarStateTotalDonations))
	if di.Amount > maxd {
		stateAccess.Set(donatewithfeedback.VarStateMaxDonation, codec.EncodeInt64(di.Amount))
	}
	stateAccess.Set(donatewithfeedback.VarStateTotalDonations, codec.EncodeInt64(total+di.Amount))

	// publish message for tracing
	ctx.Event(fmt.Sprintf("DonateWithFeedback: appended to tlog. Len: %d, Earliest: %v, Latest: %v",
		tlog.MustLen(),
		time.Unix(0, tlog.MustEarliest()).Format("2006-01-02 15:04:05"),
		time.Unix(0, tlog.MustLatest()).Format("2006-01-02 15:04:05"),
	))
	ctx.Event(fmt.Sprintf("DonateWithFeedback: donate. amount: %d, sender: %s, feedback: '%s', err: %s",
		di.Amount, di.Sender.String(), di.Feedback, di.Error))
	return nil
}

// TODO implement withdrawal of other than IOTA colored tokens
func withdraw(ctx iscp.Sandbox) error {
	ctx.Event(fmt.Sprintf("DonateWithFeedback: withdraw"))
	params := ctx.Params()

	// TODO refactor to the new account system
	//if ctx.Caller() != *ctx.OriginatorAddress() {
	//	// not authorized
	//	return fmt.Errorf("withdraw: not authorized")
	//}
	// take argument value coming with the request
	bal := ctx.Balance(ledgerstate.ColorIOTA)
	withdrawSum, amountGiven, err := codec.DecodeInt64(params.MustGet(donatewithfeedback.VarReqWithdrawSum))
	if err != nil {
		// the error from MustGetInt64 means binary data sent as a value of the variable
		// cannot be interpreted as int64
		// return everything TODO RefundAll function ?
		//
		//sender := ctx.Caller()
		//sent := ctx.IncomingTransfer().Balance(ledgerstate.ColorIOTA)
		//ctx.MoveTokens(sender, ledgerstate.ColorIOTA, sent)
		return fmt.Errorf("DonateWithFeedback: withdraw wrong argument %v", err)
	}
	// determine how much we can withdraw
	if !amountGiven {
		withdrawSum = bal
	} else {
		if withdrawSum > bal {
			withdrawSum = bal
		}
	}
	if withdrawSum == 0 {
		return fmt.Errorf("DonateWithFeedback: withdraw. nothing to withdraw")
	}
	// transfer iotas to the owner address
	// TODO refactor to new account system
	// ctx.AccessSCAccount().MoveTokens(ctx.OriginatorAddress(), &ledgerstate.ColorIOTA, withdrawSum)
	// ctx.Event(fmt.Sprintf("DonateWithFeedback: withdraw. Withdraw %d iotas", withdrawSum))
	return nil
}
