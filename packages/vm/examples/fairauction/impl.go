// hard coded implementation of the FairAuction smart contract: the NFT aution (non-fungible tokens)
package fairauction

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type fairAuctionProcessor map[sctransaction.RequestCode]fairAuctionEntryPoint

type fairAuctionEntryPoint func(ctx vmtypes.Sandbox)

const (
	RequestInitSC               = sctransaction.RequestCode(uint16(0)) // NOP
	RequestStartAuction         = sctransaction.RequestCode(uint16(1))
	RequestRemoveAuction        = sctransaction.RequestCode(uint16(2))
	RequestFinalizeAuction      = sctransaction.RequestCode(uint16(3))
	RequestPlaceBid             = sctransaction.RequestCode(uint16(4))
	RequestSetServiceFeeAuction = sctransaction.RequestCode(uint16(5) | sctransaction.RequestCodeProtected)
	RequestSetServiceFeeBid     = sctransaction.RequestCode(uint16(6) | sctransaction.RequestCodeProtected)
)

// the processor is a map of entry points
var entryPoints = fairAuctionProcessor{
	RequestInitSC:               initSC,
	RequestStartAuction:         startAuction,
	RequestRemoveAuction:        removeAuction,
	RequestFinalizeAuction:      finalizeAuction,
	RequestSetServiceFeeAuction: setServiceFeeAuction,
	RequestSetServiceFeeBid:     setServiceFeeBid,
	RequestPlaceBid:             placeBid,
}

const ProgramHash = "4NbQFgvnsfgE3n9ZhtJ3p9hWZzfYUEDHfKU93wp8UowB"
const (
	// request vars
	VarReqAuctionColor                = "color"
	VarReqStartAuctionDescription     = "dscr"
	VarReqStartAuctionDurationMinutes = "duration"
	VarReqStartAuctionMinBid          = "minimum" // in iotas
	VarReqSetFee                      = "fee"

	// state vars
	VarStateAuctions   = "auctions"
	VarStateFeeAuction = "feeAuction"
	VarStateFeeBid     = "feeBid"
)

const (
	MinAuctionDurationMinutes     = 1
	FeeAuctionDefault             = 1000000
	FeeBidDefault                 = 10000
	AuctionDurationDefaultMinutes = 60
)

func GetProcessor() vmtypes.Processor {
	return entryPoints
}

func (v fairAuctionProcessor) GetEntryPoint(code sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	f, ok := v[code]
	return f, ok
}

func (ep fairAuctionEntryPoint) Run(ctx vmtypes.Sandbox) {
	ep(ctx)
}

func (ep fairAuctionEntryPoint) WithGasLimit(_ int) vmtypes.EntryPoint {
	return ep
}

type AuctionInfo struct {
	// color of the tokens for sale. Max one auction per color at each time is allowed
	// all tokens are being sold as one lot
	Color balance.Color
	// number of tokens for sale
	NumTokens int64
	// minimum bid
	MinimumBid int64
	// any text, like "Owner of the token have a right to call me for a date"
	Description string
	// timestamp when auction started
	WhenStarted int64
	// duration of the auctions in minutes. Should be no less than MinAuctionDurationMinutes
	DurationMinutes int64
	// address which issued StartAuction transaction
	Owner address.Address
	// list of bids
	Bids []*BidInfo
}

type BidInfo struct {
	// total sum of the bid = total amount of iotas available in the request - 1 - SC reward - ServiceFeeBid
	Total int64
	// address which sent the bid
	Bidder address.Address
	// timestamp
	WhenPlaced int64
}

func initSC(ctx vmtypes.Sandbox) {
}

// startAuction processes the StartAuction request
// Arguments:
// - VarReqAuctionColor: color of the tokens for sale
// - VarReqStartAuctionDescription: description of the lot
// - VarReqStartAuctionMinBid: minimum price for the whole lot
// - VarReqStartAuctionDurationMinutes: duration of auction
// Request transaction must contain at least:
// - 1 token for sale plus
// - VarStateFeeAuction + 1 iotas
func startAuction(ctx vmtypes.Sandbox) {
	ctx.Publish("startAuction begin")

	// find out who starts the action
	sender := takeSender(ctx)
	if sender == nil {
		// wrong transaction, hard reject (nothing is processed and nothing is refunded)
		return
	}
	// get currently set service fee values
	feeAuction, _ := getFeeValues(ctx)

	reqArgs := ctx.AccessRequest().Args()
	account := ctx.AccessOwnAccount()

	// determine color of the token for sale
	colh, ok, err := reqArgs.GetHashValue(VarReqAuctionColor)
	if err != nil || !ok {
		// incorrect request arguments
		refundFromRequest(ctx, &balance.ColorIOTA, feeAuction)

		ctx.Publish("startAuction: exit 1")
		return
	}
	colorForSale := (balance.Color)(*colh)
	if colorForSale == balance.ColorIOTA || colorForSale == balance.ColorNew {
		// reserved color not allowed
		refundFromRequest(ctx, &balance.ColorIOTA, feeAuction)

		ctx.Publish("startAuction: exit 2")
		return
	}

	// check if enough iotas for service fees to create the auction
	if account.AvailableBalanceFromRequest(&balance.ColorIOTA) < feeAuction {
		// not enough fees
		// return half of iotas and all tokens for sale (if any)
		refundFromRequest(ctx, &balance.ColorIOTA, feeAuction/2)
		refundFromRequest(ctx, &colorForSale, 0)

		ctx.Publish("startAuction: exit 3")
		return
	}

	// determine amount of colored tokens for sale
	tokensForSale := account.AvailableBalanceFromRequest(&colorForSale)
	if tokensForSale == 0 {
		// no tokens transferred
		// refund half fee
		refundFromRequest(ctx, &balance.ColorIOTA, feeAuction/2)

		ctx.Publish("startAuction: exit 4")
		return
	}
	//
	minimumBid, _, err := reqArgs.GetInt64(VarReqStartAuctionMinBid)
	if err != nil {
		// wrong argument. Hard reject

		ctx.Publish("startAuction: exit 5")
		return
	}

	// determine duration of the auction. Take default if no set in request and ensure minimum
	duration, ok, err := reqArgs.GetInt64(VarReqStartAuctionDurationMinutes)
	if err != nil {
		// fatal error
		return
	}
	if !ok {
		duration = AuctionDurationDefaultMinutes
	}
	if duration < MinAuctionDurationMinutes {
		duration = MinAuctionDurationMinutes
	}

	// read description text
	description, ok, err := reqArgs.GetString(VarReqStartAuctionDescription)
	if err != nil {
		return
	}
	if !ok {
		description = "N/A"
	}

	// find out if auction for this color already exist in the dictionary
	auctions := ctx.AccessState().GetDictionary(VarStateAuctions)
	if b := auctions.GetAt(colorForSale.Bytes()); b != nil {
		// auction already exists. Ignore sale auction.
		// refund iotas less fee
		refundFromRequest(ctx, &balance.ColorIOTA, feeAuction)
		// return all tokens for sale
		refundFromRequest(ctx, &colorForSale, 0)

		ctx.Publish("startAuction: exit 6")
		return
	}
	// create record for the new auction in the dictionary
	aiData := util.MustBytes(&AuctionInfo{
		Color:           colorForSale,
		NumTokens:       tokensForSale,
		MinimumBid:      minimumBid,
		Description:     description,
		WhenStarted:     ctx.GetTimestamp(),
		DurationMinutes: duration,
		Owner:           *sender,
	})
	auctions.SetAt(colorForSale.Bytes(), aiData)

	// prepare and send time locked for the duration FinalizeAuction request to self
	args := kv.NewMap()
	args.Codec().SetHashValue(VarReqAuctionColor, (*hashing.HashValue)(&colorForSale))
	ctx.SendRequestToSelfWithDelay(RequestFinalizeAuction, args, uint32(duration*60))

	account.HarvestFeesFromRequest(feeAuction)

	ctx.Publish("startAuction end")
}

// removeAuction processes remove auction request
// Arguments:
// - VarReqAuctionColor: color of the auction
func removeAuction(ctx vmtypes.Sandbox) {
	ctx.Publish("removeAuction begin")

	reqArgs := ctx.AccessRequest().Args()

	// determine color of the auction
	colh, ok, err := reqArgs.GetHashValue(VarReqAuctionColor)
	if err != nil || !ok {
		// incorrect request arguments.
		ctx.Publish("removeAuction: exit 1")
		return
	}
	col := (balance.Color)(*colh)
	if col == balance.ColorIOTA || col == balance.ColorNew {
		// reserved color not allowed
		ctx.Publish("removeAuction: exit 2")
		return
	}

	// find the record for the auction by color
	auctDict := ctx.AccessState().GetDictionary(VarStateAuctions)
	data := auctDict.GetAt(col.Bytes())
	if data == nil {
		// nothing to remove
		ctx.Publish("removeAuction: exit 3")
		return
	}
	// decode the record
	ai := &AuctionInfo{}
	if err := ai.Read(bytes.NewReader(data)); err != nil {
		// internal error
		ctx.Publish("removeAuction: exit 4")
		return
	}
	// check if sender is authorized to remove auction
	if !ctx.AccessRequest().IsAuthorisedByAddress(&ai.Owner) {
		// not authorised
		ctx.Publish("removeAuction: exit 5")
		return
	}
	// return bid amounts to bidders
	account := ctx.AccessOwnAccount()
	for _, bi := range ai.Bids {
		account.MoveTokens(&bi.Bidder, &balance.ColorIOTA, bi.Total)
	}
	// return tokens for sale to the auction owner
	account.MoveTokens(&ai.Owner, &ai.Color, ai.NumTokens)
	// delete auction record
	auctDict.DelAt(col.Bytes())

	ctx.Publish("removeAuction success")
}

// finalizeAuction selects the winner and sends him tokens.
// returns bid amounts to other bidders
// The request is normally time locked for the period of the action
// Arguments:
// - VarReqAuctionColor: color of the auction
func finalizeAuction(ctx vmtypes.Sandbox) {
	ctx.Publish("finalizeAuction begin")

	accessReq := ctx.AccessRequest()
	if !accessReq.IsAuthorisedByAddress(ctx.GetSCAddress()) {
		// finalizeAuction request can only be sent by the smart contract to itself
		return
	}
	reqArgs := accessReq.Args()

	// determine color of the auction to finalize
	colh, ok, err := reqArgs.GetHashValue(VarReqAuctionColor)
	if err != nil || !ok {
		// incorrect request arguments
		// internal error. Refund completely?
		ctx.Publish("finalizeAuction: exit 1")
		return
	}
	col := (balance.Color)(*colh)
	if col == balance.ColorIOTA || col == balance.ColorNew {
		// internal error. Refund completely?
		ctx.Publish("finalizeAuction: exit 2")
		return
	}

	// find the record of the auction by color
	auctDict := ctx.AccessState().GetDictionary(VarStateAuctions)
	data := auctDict.GetAt(col.Bytes())
	if data == nil {
		// auction with this color does not exist, most likely removed
		ctx.Publish("finalizeAuction: exit 3")
		return
	}

	// decode the Action record
	ai := &AuctionInfo{}
	if err := ai.Read(bytes.NewReader(data)); err != nil {
		// internal error. Refund completely?
		return
	}

	account := ctx.AccessOwnAccount()

	if len(ai.Bids) == 0 {
		// no bids
		// return tokens to owner
		account.MoveTokens(&ai.Owner, &ai.Color, ai.NumTokens)
		// delete auction record
		auctDict.DelAt(col.Bytes())

		ctx.Publish("finalizeAuction: exit 4")
		return
	}
	// find the winning amount
	winningBid := int64(0)
	for _, bi := range ai.Bids {
		if bi.Total > winningBid {
			winningBid = bi.Total
		}
	}
	var winner *BidInfo
	var winnerIndex int
	for i, bi := range ai.Bids {
		if bi.Total == winningBid {
			// taking the first among equals
			winner = bi
			winnerIndex = i
			break
		}
	}
	if winner == nil {
		// inconsistency
		ctx.Publish("finalizeAuction: exit 5")
		return
	}

	for i, bi := range ai.Bids {
		if i == winnerIndex {
			// send bid sum to the owner of the auction
			account.MoveTokens(&ai.Owner, &balance.ColorIOTA, winningBid)
			// send tokens to the winner
			account.MoveTokens(&winner.Bidder, &ai.Color, ai.NumTokens)
		} else {
			// return staked sum to the non-winner
			account.MoveTokens(&bi.Bidder, &balance.ColorIOTA, bi.Total)
		}
	}

	// delete auction record
	auctDict.DelAt(col.Bytes())

	ctx.Publish("finalizeAuction: success")
}

// setServiceFeeAuction is a request to set the service fee to start the auction
// It is protected, i.e. must be sent by the owner of the smart contract
// Arguments:
// - VarReqSetFee: the fee value
func setServiceFeeAuction(ctx vmtypes.Sandbox) {
	ctx.Publish("setServiceFeeAuction: begin")
	fee, ok, err := ctx.AccessRequest().Args().GetInt64(VarReqSetFee)
	if err != nil || !ok {
		ctx.Publish("setServiceFeeAuction: exit 1")
		return
	}
	ctx.AccessState().SetInt64(VarStateFeeAuction, fee)
	ctx.Publish("setServiceFeeAuction: success")
}

// setServiceFeeBid is a request to set the service fee to place a bid
// It is protected, i.e. must be sent by the owner of the smart contract
// Arguments:
// - VarReqSetFee: the fee value
func setServiceFeeBid(ctx vmtypes.Sandbox) {
	ctx.Publish("setServiceFeeBid: begin")
	fee, ok, err := ctx.AccessRequest().Args().GetInt64(VarReqSetFee)
	if err != nil || !ok {
		ctx.Publish("setServiceFeeBid: exit 1")
		return
	}
	ctx.AccessState().SetInt64(VarStateFeeBid, fee)
	ctx.Publish("setServiceFeeBid: success")
}

// placeBid is a request to place a bid in the auction for the particular color
// The request transaction must contain at least:
// - 1 request token + VarStateFeeBid + 1 Bid amount/rise amount
// In case it is not the first bid by this bidder, respective iotas are treated as
// a rise of the bid and is added to the total
// Arguments:
// - VarReqAuctionColor: color of the tokens for sale
func placeBid(ctx vmtypes.Sandbox) {
	ctx.Publish("placeBid: begin")
	_, feeBid := getFeeValues(ctx)

	// check if enough iotas
	total := ctx.AccessOwnAccount().AvailableBalanceFromRequest(&balance.ColorIOTA)
	if total <= feeBid {
		// not enough iotas to bid.
		refundFromRequest(ctx, &balance.ColorIOTA, feeBid/2)
		ctx.Publish("placeBid: exit 1")
		return
	}
	bidSum := total - feeBid
	reqArgs := ctx.AccessRequest().Args()

	// determine color of the bid
	colh, ok, err := reqArgs.GetHashValue(VarReqAuctionColor)
	if err != nil {
		// inconsistency. return all?
		ctx.Publish("placeBid: exit 2")
		return
	}
	if !ok {
		// incorrect arguments
		ctx.Publish("placeBid: exit 3")
		return
	}

	col := (balance.Color)(*colh)
	if col == balance.ColorIOTA || col == balance.ColorNew {
		// reserved color not allowed. Incorrect arguments
		ctx.Publish("placeBid: exit 4")
		return
	}

	// find record for the auction
	auctions := ctx.AccessState().GetDictionary(VarStateAuctions)
	data := auctions.GetAt(col.Bytes())
	if data == nil {
		// no such auction
		// return everything less fee
		refundFromRequest(ctx, &balance.ColorIOTA, feeBid)
		ctx.Publish("placeBid: exit 5")
		return
	}
	ai := &AuctionInfo{}
	if err := ai.Read(bytes.NewReader(data)); err != nil {
		// internal error
		ctx.Publish("placeBid: exit 6")
		return
	}

	sender := takeSender(ctx)
	if sender == nil {
		// bad transaction
		ctx.Publish("placeBid: exit 7")
		return
	}
	var bi *BidInfo
	for _, bitmp := range ai.Bids {
		if bitmp.Bidder == *sender {
			bi = bitmp
			break
		}
	}
	if bi == nil {
		// first bid by the sender
		ai.Bids = append(ai.Bids, &BidInfo{
			Total:      bidSum,
			Bidder:     *sender,
			WhenPlaced: ctx.GetTimestamp(),
		})
	} else {
		// bid is treated as rise
		bi.Total += bidSum
	}
	data = util.MustBytes(ai)
	auctions.SetAt(col.Bytes(), data)

	ctx.AccessOwnAccount().HarvestFeesFromRequest(feeBid)
	ctx.Publish("placeBid: success")
}

func takeSender(ctx vmtypes.Sandbox) *address.Address {
	// take input addresses of the request transaction. Must be exactly 1 otherwise.
	// Theoretically the transaction may have several addresses in inputs, then it is ignored
	senders := ctx.AccessRequest().Senders()
	if len(senders) != 1 {
		// wrong transaction, hardReject (nothing is processed and nothing is refunded)
		return nil
	}
	sender := senders[0]
	return &sender
}

// refundFromRequest returns all iotas tokens to the sender minus sunkFee
func refundFromRequest(ctx vmtypes.Sandbox, color *balance.Color, harvest int64) {
	account := ctx.AccessOwnAccount()
	ctx.AccessOwnAccount().HarvestFeesFromRequest(harvest)
	available := account.AvailableBalanceFromRequest(color)
	sender := takeSender(ctx)
	if sender == nil {
		return
	}
	ctx.AccessOwnAccount().HarvestFeesFromRequest(harvest)
	account.MoveTokensFromRequest(sender, color, available)

}

func getFeeValues(ctx vmtypes.Sandbox) (int64, int64) {
	stateAccess := ctx.AccessState()
	feeAuction, ok := stateAccess.GetInt64(VarStateFeeAuction)
	if !ok {
		feeAuction = FeeAuctionDefault
	}
	feeBid, ok := stateAccess.GetInt64(VarStateFeeBid)
	if !ok {
		feeAuction = FeeBidDefault
	}
	return feeAuction, feeBid
}
