// hard coded smart contract code implements DonateWithFeedback smart cintract.
package dwfimpl

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/examples/donatewithfeedback"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"strings"
	"time"
)

// program hash: the ID of the code
const ProgramHash = "5ydEfDeAJZX6dh6Fy7tMoHcDeh42gENeqVDASGWuD64X"

// implementation of 'vmtypes.Processor' and 'vmtypes.EntryPoint' interfaces
type dwfProcessor map[sctransaction.RequestCode]dwfEntryPoint

type dwfEntryPoint func(ctx vmtypes.Sandbox)

// the processor implementation is a map of entry points: one for each request
var entryPoints = dwfProcessor{
	donatewithfeedback.RequestDonate:   donate,
	donatewithfeedback.RequestWithdraw: withdraw,
}

// point of attachment of hard coded code to the rest of Wasp
func GetProcessor() vmtypes.Processor {
	return entryPoints
}

// GetEntryPoint implements EntryPoint interfaces. It resolves request code to the
// function
func (v dwfProcessor) GetEntryPoint(code sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	f, ok := v[code]
	return f, ok
}

// GetDescription description of the smart contract
func (v dwfProcessor) GetDescription() string {
	return "DonateWithFeedback hard coded smart contract processor"
}

// Run calls the function wrapped into the EntryPoint
func (ep dwfEntryPoint) Run(ctx vmtypes.Sandbox) {
	ep(ctx)
}

// not used
func (ep dwfEntryPoint) WithGasLimit(_ int) vmtypes.EntryPoint {
	return ep
}

// donate implements request 'donate'. It takes feedback text from the request and adds it into the
// log of feedback messages
func donate(ctx vmtypes.Sandbox) {
	ctx.Publishf("DonateWithFeedback: donate")

	// how many iotas are sent by the request.
	// only iotas are considered donation. Other colors are ignored
	donated := ctx.AccessSCAccount().AvailableBalanceFromRequest(&balance.ColorIOTA)
	// take feedback text contained in the request
	feedback, ok, err := ctx.AccessRequest().Args().GetString(donatewithfeedback.VarReqFeedback)

	stateAccess := ctx.AccessState()
	tlog := stateAccess.GetTimestampedLog(donatewithfeedback.VarStateTheLog)

	// determine sender of the request
	// create donation info record
	reqId := ctx.AccessRequest().ID()
	sender := ctx.AccessRequest().Sender()
	di := &donatewithfeedback.DonationInfo{
		Seq:      int64(tlog.Len()),
		Id:       &reqId,
		Amount:   donated,
		Sender:   sender,
		Feedback: feedback,
	}
	if err != nil {
		di.Error = err.Error()
	} else {
		if !ok || len(strings.TrimSpace(feedback)) == 0 || donated == 0 {
			// empty feedback message is considered an error
			di.Error = "error: empty feedback or donated amount = 0. The donated amount has been returned (if any)"
		}
	}
	if len(di.Error) != 0 && donated > 0 {
		// if error occurred, return all donated tokens back to the sender
		// in this case error message will be recorded in the donation record
		ctx.AccessSCAccount().MoveTokensFromRequest(&sender, &balance.ColorIOTA, donated)
		di.Amount = 0
	}
	// store donation info record in the state (append to the timestamped log)
	tlog.Append(ctx.GetTimestamp(), di.Bytes())

	// save total and maximum donations
	maxd, _ := stateAccess.GetInt64(donatewithfeedback.VarStateMaxDonation)
	total, _ := stateAccess.GetInt64(donatewithfeedback.VarStateTotalDonations)
	if di.Amount > maxd {
		stateAccess.SetInt64(donatewithfeedback.VarStateMaxDonation, di.Amount)
	}
	stateAccess.SetInt64(donatewithfeedback.VarStateTotalDonations, total+di.Amount)

	// publish message for tracing
	ctx.Publishf("DonateWithFeedback: appended to tlog. Len: %d, Earliest: %v, Latest: %v",
		tlog.Len(),
		time.Unix(0, tlog.Earliest()).Format("2006-01-02 15:04:05"),
		time.Unix(0, tlog.Latest()).Format("2006-01-02 15:04:05"),
	)
	ctx.Publishf("DonateWithFeedback: donate. amount: %d, sender: %s, feedback: '%s', err: %s",
		di.Amount, di.Sender.String(), di.Feedback, di.Error)
}

// protected request. Only owner can withdraw iotas from smart contract at any time
// this function will only be called if the request transaction contains signature from the
// smart contract owner. It is checked before calling the VM
// TODO implement withdrawal of other than IOTA colored tokens
func withdraw(ctx vmtypes.Sandbox) {
	ctx.Publishf("DonateWithFeedback: withdraw")

	// take argument value coming with the request
	bal := ctx.AccessSCAccount().AvailableBalance(&balance.ColorIOTA)
	withdrawSum, amountGiven, err := ctx.AccessRequest().Args().GetInt64(donatewithfeedback.VarReqWithdrawSum)
	if err != nil {
		// the error from GetInt64 means binary data sent as a value of the variable
		// cannot be interpreted as int64
		ctx.Publishf("DonateWithFeedback: withdraw wrong argument %v", err)
		// return everything TODO RefundAll function ?
		sender := ctx.AccessRequest().Sender()
		sent := ctx.AccessSCAccount().AvailableBalanceFromRequest(&balance.ColorIOTA)
		ctx.AccessSCAccount().MoveTokensFromRequest(&sender, &balance.ColorIOTA, sent)
		return
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
		ctx.Publishf("DonateWithFeedback: withdraw. nothing to withdraw")
		return
	}
	// transfer iotas to the owner address
	ctx.AccessSCAccount().MoveTokens(ctx.GetOwnerAddress(), &balance.ColorIOTA, withdrawSum)
	ctx.Publishf("DonateWithFeedback: withdraw. Withdraw %d iotas", withdrawSum)
}
