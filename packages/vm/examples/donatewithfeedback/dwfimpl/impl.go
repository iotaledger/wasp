// smart contract implements Token Registry. User can mint any number of new colored tokens to own address
// and in the same transaction can register the whole Supply of new tokens in the TokenRegistry.
// TokenRegistry contains metadata. It can be changed by the owner of the record
// Initially the owner is the minter. Owner can transfer ownership of the metadata record to another address
package dwfimpl

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/examples/donatewithfeedback"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"strings"
)

const ProgramHash = "5ydEfDeAJZX6dh6Fy7tMoHcDeh42gENeqVDASGWuD64X"

type dwfProcessor map[sctransaction.RequestCode]dwfEntryPoint

type dwfEntryPoint func(ctx vmtypes.Sandbox)

// the processor is a map of entry points
var entryPoints = dwfProcessor{
	donatewithfeedback.RequestDonate:  donate,
	donatewithfeedback.RequestHarvest: harvest,
}

func GetProcessor() vmtypes.Processor {
	return entryPoints
}

func (v dwfProcessor) GetEntryPoint(code sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	f, ok := v[code]
	return f, ok
}

// does nothing, i.e. resulting state update is empty
func (ep dwfEntryPoint) Run(ctx vmtypes.Sandbox) {
	ep(ctx)
}

func (ep dwfEntryPoint) WithGasLimit(_ int) vmtypes.EntryPoint {
	return ep
}

func donate(ctx vmtypes.Sandbox) {
	ctx.Publishf("DonateWithFeedback: donate")

	// other color not taken into account
	donated := ctx.AccessSCAccount().AvailableBalanceFromRequest(&balance.ColorIOTA)
	feedback, ok, err := ctx.AccessRequest().Args().GetString(donatewithfeedback.VarReqFeedback)

	sender := ctx.AccessRequest().Sender()
	di := &donatewithfeedback.DonationInfo{
		Amount: donated,
		Sender: sender,
	}
	if err != nil {
		di.Error = err.Error()
	} else {
		if !ok || len(strings.TrimSpace(feedback)) == 0 || donated == 0 {
			di.Error = "error: empty feedback or donated amount = 0. The donated amount has been returned (if any)"
		}
	}
	if len(di.Error) != 0 && donated > 0 {
		ctx.AccessSCAccount().MoveTokensFromRequest(&sender, &balance.ColorIOTA, donated)
		di.Amount = 0
	}
	ctx.AccessState().GetArray(donatewithfeedback.VarStateTheLog).Push(di.Bytes())
	ctx.Publishf("DonateWithFeedback: donate. amount: %d, sender: %s, feedback: '%s', err: %s",
		di.Amount, di.Sender.String(), di.Feedback, di.Error)
}

// protected request. Owner can take iotas at any time
func harvest(ctx vmtypes.Sandbox) {
	ctx.Publishf("DonateWithFeedback: harvest")

	bal := ctx.AccessSCAccount().AvailableBalance(&balance.ColorIOTA)
	if bal > 0 {
		ctx.AccessSCAccount().MoveTokens(ctx.GetOwnerAddress(), &balance.ColorIOTA, bal)
		ctx.Publishf("DonateWithFeedback: harvest. Harvested %d iotas", bal)
		return
	}
	ctx.Publishf("DonateWithFeedback: harvest. account is empty")
	return
}
