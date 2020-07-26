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
	RequestInitSC          = sctransaction.RequestCode(uint16(0)) // NOP
	RequestStartAuction    = sctransaction.RequestCode(uint16(1))
	RequestFinalizeAuction = sctransaction.RequestCode(uint16(2))
	RequestPlaceBid        = sctransaction.RequestCode(uint16(3))
	RequestSetOwnerMargin  = sctransaction.RequestCode(uint16(4) | sctransaction.RequestCodeProtected)
)

// the processor is a map of entry points
var entryPoints = fairAuctionProcessor{
	RequestInitSC:          initSC,
	RequestStartAuction:    startAuction,
	RequestFinalizeAuction: finalizeAuction,
	RequestSetOwnerMargin:  setOwnerMargin,
	RequestPlaceBid:        placeBid,
}

const ProgramHash = "4NbQFgvnsfgE3n9ZhtJ3p9hWZzfYUEDHfKU93wp8UowB"
const (
	// request vars
	VarReqAuctionColor                = "color"
	VarReqStartAuctionDescription     = "dscr"
	VarReqStartAuctionDurationMinutes = "duration"
	VarReqStartAuctionMinimumBid      = "minimum" // in iotas
	VarReqOwnerMargin                 = "ownerMargin"

	// state vars
	VarStateAuctions            = "auctions"
	VarStateOwnerMarginPromille = "ownerMargin" // owner margin in percents
)

const (
	MinAuctionDurationMinutes     = 1
	AuctionDurationDefaultMinutes = 60
	OwnerMarginDefault            = 30  // 3%
	OwnerMarginMin                = 5   // minimum 0.5%
	OwnerMarginMax                = 100 // max 10%
)

// validating constants
func init() {
	if OwnerMarginMax > 1000 ||
		OwnerMarginMin < 0 ||
		OwnerMarginDefault < OwnerMarginMin ||
		OwnerMarginDefault > OwnerMarginMax ||
		OwnerMarginMin > OwnerMarginMax {
		panic("wrong constants")
	}
}

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
	// any text, like "AuctionOwner of the token have a right to call me for a date"
	Description string
	// timestamp when auction started
	WhenStarted int64
	// duration of the auctions in minutes. Should be no less than MinAuctionDurationMinutes
	DurationMinutes int64
	// address which issued StartAuction transaction
	AuctionOwner address.Address
	// total deposit by the auction owner
	TotalDeposit int64
	// AuctionOwner's margin in promilles, frozen at the moment of creation of smart contract
	OwnerMargin int64
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
// - VarReqStartAuctionMinimumBid: minimum price for the whole lot
// - VarReqStartAuctionDurationMinutes: duration of auction
// Request transaction must contain at least number of iotas >= of current owner margin from the minimum bid
// (not including node reward with request token)
func startAuction(ctx vmtypes.Sandbox) {
	ctx.Publish("startAuction begin")

	// find out who starts the action
	sender := takeSender(ctx)
	if sender == nil {
		// wrong transaction, hard reject (nothing is processed and nothing is refunded)
		return
	}
	reqArgs := ctx.AccessRequest().Args()
	account := ctx.AccessOwnAccount()

	totalDeposit := account.AvailableBalanceFromRequest(&balance.ColorIOTA)

	ownerMargin := getOwnerMarginPromille(ctx)

	// determine color of the token for sale
	colh, ok, err := reqArgs.GetHashValue(VarReqAuctionColor)
	if err != nil || !ok {
		// incorrect request arguments
		refundFromRequest(ctx, &balance.ColorIOTA, totalDeposit/2)

		ctx.Publish("startAuction: exit 1")
		return
	}
	colorForSale := (balance.Color)(*colh)
	if colorForSale == balance.ColorIOTA || colorForSale == balance.ColorNew {
		// reserved color not allowed
		refundFromRequest(ctx, &balance.ColorIOTA, totalDeposit/2)

		ctx.Publish("startAuction: exit 2")
		return
	}

	// determine amount of colored tokens for sale
	tokensForSale := account.AvailableBalanceFromRequest(&colorForSale)
	if tokensForSale == 0 {
		// no tokens transferred
		refundFromRequest(ctx, &balance.ColorIOTA, totalDeposit/2)

		ctx.Publish("startAuction exit 3: no tokens for sale")
		return
	}

	minimumBid, _, err := reqArgs.GetInt64(VarReqStartAuctionMinimumBid)
	if err != nil {
		// wrong argument. Hard reject

		ctx.Publish("startAuction: exit 4")
		return
	}
	// ensure tokens are not sold for the minimum price less than 1 iota per token!
	if minimumBid < tokensForSale {
		minimumBid = tokensForSale
	}

	// check if enough iotas for service fees to create the auction
	// minimum deposit is owner margin from minimum bid
	expectedDeposit := (minimumBid * ownerMargin) / 1000
	if totalDeposit < expectedDeposit {
		// not enough fees
		// return iotas less half of expected deposit and all tokens for sale (if any)
		refundFromRequest(ctx, &balance.ColorIOTA, expectedDeposit/2)
		refundFromRequest(ctx, &colorForSale, 0)

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
		refundFromRequest(ctx, &balance.ColorIOTA, expectedDeposit/2)
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
		AuctionOwner:    *sender,
		TotalDeposit:    totalDeposit,
		OwnerMargin:     ownerMargin,
	})
	auctions.SetAt(colorForSale.Bytes(), aiData)

	// prepare and send request FinalizeAuction to self time-locked for the duration
	args := kv.NewMap()
	args.Codec().SetHashValue(VarReqAuctionColor, (*hashing.HashValue)(&colorForSale))
	ctx.SendRequestToSelfWithDelay(RequestFinalizeAuction, args, uint32(duration*60))

	ctx.Publish("startAuction: success")
}

// finalizeAuction selects the winner and sends tokens to him.
// returns bid amounts to other bidders.
// The request is time locked for the period of the action
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
		// inconsistency
		ctx.Publish("finalizeAuction: exit 2")
		return
	}

	// find the record of the auction by color
	auctDict := ctx.AccessState().GetDictionary(VarStateAuctions)
	data := auctDict.GetAt(col.Bytes())
	if data == nil {
		// auction with this color does not exist. Inconsistency
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
	ownerFee := (ai.MinimumBid * ai.OwnerMargin) / 1000

	if len(ai.Bids) == 0 {
		// no bids
		// return tokens to owner
		account.MoveTokens(&ai.AuctionOwner, &ai.Color, ai.NumTokens)
		//
		ctx.AccessOwnAccount().HarvestFees(ownerFee)

		// delete auction record
		auctDict.DelAt(col.Bytes())

		ctx.Publish("finalizeAuction: exit 4")
		return
	}
	// find the winning amount and determine respective ownerFee
	winningAmount := int64(0)
	for _, bi := range ai.Bids {
		if bi.Total > winningAmount {
			winningAmount = bi.Total
		}
	}

	var winner *BidInfo
	var winnerIndex int

	if winningAmount >= ai.MinimumBid {
		ownerFee = (winningAmount * ai.OwnerMargin) / 1000

		for i, bi := range ai.Bids {
			if bi.Total == winningAmount {
				// taking the first among equals
				winner = bi
				winnerIndex = i
				break
			}
		}
	}

	for i, bi := range ai.Bids {
		if i == winnerIndex && winner != nil {
			// send bid and return deposit sum less fees to the owner of the auction
			account.MoveTokens(&ai.AuctionOwner, &balance.ColorIOTA, winningAmount+ai.TotalDeposit-ownerFee)
			// send tokens to the winner
			account.MoveTokens(&winner.Bidder, &ai.Color, ai.NumTokens)
		} else {
			// return staked sum to the non-winner
			account.MoveTokens(&bi.Bidder, &balance.ColorIOTA, bi.Total)
		}
	}
	ctx.AccessOwnAccount().HarvestFees(ownerFee)

	// delete auction record
	auctDict.DelAt(col.Bytes())

	ctx.Publish("finalizeAuction: success")
}

// setOwnerMargin is a request to set the service fee to place a bid
// It is protected, i.e. must be sent by the owner of the smart contract
// Arguments:
// - VarReqOwnerMargin: the margin value in promilles
func setOwnerMargin(ctx vmtypes.Sandbox) {
	ctx.Publish("setOwnerMargin: begin")
	margin, ok, err := ctx.AccessRequest().Args().GetInt64(VarReqOwnerMargin)
	if err != nil || !ok {
		ctx.Publish("setOwnerMargin: exit 1")
		return
	}
	if margin < OwnerMarginMin {
		margin = OwnerMarginMin
	} else if margin > OwnerMarginMax {
		margin = OwnerMarginMax
	}
	ctx.AccessState().SetInt64(VarStateOwnerMarginPromille, margin)
	ctx.Publishf("setOwnerMargin: success. ownerMargin set to %d%%", margin/10)
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

	bidAmount := ctx.AccessOwnAccount().AvailableBalanceFromRequest(&balance.ColorIOTA)
	if bidAmount == 0 {
		// no iotas sent
		ctx.Publish("placeBid: exit 0")
		return
	}

	// check if enough iotas
	reqArgs := ctx.AccessRequest().Args()

	// determine color of the bid
	colh, ok, err := reqArgs.GetHashValue(VarReqAuctionColor)
	if err != nil {
		// inconsistency. return all?
		ctx.Publish("placeBid: exit 1")
		return
	}
	if !ok {
		// incorrect arguments
		ctx.Publish("placeBid: exit 2")
		refundFromRequest(ctx, &balance.ColorIOTA, 0)
		return
	}

	col := (balance.Color)(*colh)
	if col == balance.ColorIOTA || col == balance.ColorNew {
		// reserved color not allowed. Incorrect arguments
		refundFromRequest(ctx, &balance.ColorIOTA, 0)
		ctx.Publish("placeBid: exit 3")
		return
	}

	// find record for the auction
	auctions := ctx.AccessState().GetDictionary(VarStateAuctions)
	data := auctions.GetAt(col.Bytes())
	if data == nil {
		// no such auction. refund everything
		refundFromRequest(ctx, &balance.ColorIOTA, 0)
		ctx.Publish("placeBid: exit 4")
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
			Total:      bidAmount,
			Bidder:     *sender,
			WhenPlaced: ctx.GetTimestamp(),
		})
	} else {
		// bid is treated as a rise
		bi.Total += bidAmount
	}
	data = util.MustBytes(ai)
	auctions.SetAt(col.Bytes(), data)

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

func getOwnerMarginPromille(ctx vmtypes.Sandbox) int64 {
	ownerMargin, ok := ctx.AccessState().GetInt64(VarStateOwnerMarginPromille)
	if !ok {
		ownerMargin = OwnerMarginMin
	} else {
		if ownerMargin > OwnerMarginMax {
			ownerMargin = OwnerMarginMax
		}
	}
	return ownerMargin
}
