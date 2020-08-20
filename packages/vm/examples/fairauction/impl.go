// hard coded implementation of the FairAuction smart contract: the NFT aution (non-fungible tokens)
package fairauction

import (
	"bytes"
	"sort"

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
	OwnerMarginDefault            = 50  // 5%
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
	When int64
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
	sender := ctx.AccessRequest().Sender()
	reqArgs := ctx.AccessRequest().Args()
	account := ctx.AccessSCAccount()

	totalDeposit := account.AvailableBalanceFromRequest(&balance.ColorIOTA)
	if totalDeposit < 1 {
		// it is expected at least 1 iota in deposit
		refundFromRequest(ctx, &balance.ColorIOTA, 1)

		ctx.Publish("startAuction: exit 0: must be at least 1i in deposit")
		return
	}

	ownerMargin := GetOwnerMarginPromille(ctx.AccessState().GetInt64(VarStateOwnerMarginPromille))

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
	expectedDeposit := GetExpectedDeposit(minimumBid, ownerMargin)

	if totalDeposit < expectedDeposit {
		// not enough fees
		// return half of expected deposit and all tokens for sale (if any)
		harvest := expectedDeposit / 2
		if harvest < 1 {
			harvest = 1
		}
		refundFromRequest(ctx, &balance.ColorIOTA, harvest)
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
		AuctionOwner:    sender,
		TotalDeposit:    totalDeposit,
		OwnerMargin:     ownerMargin,
	})
	auctions.SetAt(colorForSale.Bytes(), aiData)

	ctx.Publishf("New auction record. color: %s, numTokens: %d, minBid: %d, ownerMargin: %d",
		colorForSale.String(), tokensForSale, minimumBid, ownerMargin)

	// prepare and send request FinalizeAuction to self time-locked for the duration
	args := kv.NewMap()
	args.Codec().SetHashValue(VarReqAuctionColor, (*hashing.HashValue)(&colorForSale))
	ctx.SendRequestToSelfWithDelay(RequestFinalizeAuction, args, uint32(duration*60))

	ctx.Publishf("startAuction: success. Auction: '%s'", description)
}

// placeBid is a request to place a bid in the auction for the particular color
// The request transaction must contain at least:
// - 1 request token + Bid amount/rise amount
// In case it is not the first bid by this bidder, respective iotas are treated as
// a rise of the bid and is added to the total
// Arguments:
// - VarReqAuctionColor: color of the tokens for sale
func placeBid(ctx vmtypes.Sandbox) {
	ctx.Publish("placeBid: begin")

	bidAmount := ctx.AccessSCAccount().AvailableBalanceFromRequest(&balance.ColorIOTA)
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

	sender := ctx.AccessRequest().Sender()

	var bi *BidInfo
	for _, bitmp := range ai.Bids {
		if bitmp.Bidder == sender {
			bi = bitmp
			break
		}
	}
	if bi == nil {
		// first bid by the sender
		ai.Bids = append(ai.Bids, &BidInfo{
			Total:  bidAmount,
			Bidder: sender,
			When:   ctx.GetTimestamp(),
		})
	} else {
		// bid is treated as a rise
		bi.Total += bidAmount
		bi.When = ctx.GetTimestamp()
	}
	data = util.MustBytes(ai)
	auctions.SetAt(col.Bytes(), data)

	ctx.Publishf("placeBid: success. Auction: '%s'", ai.Description)
}

// finalizeAuction selects the winner and sends tokens to him.
// returns bid amounts to other bidders.
// The request is time locked for the period of the action
// Arguments:
// - VarReqAuctionColor: color of the auction
func finalizeAuction(ctx vmtypes.Sandbox) {
	ctx.Publish("finalizeAuction begin")

	accessReq := ctx.AccessRequest()
	if accessReq.Sender() != *ctx.GetSCAddress() {
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
		ctx.Publish("finalizeAuction: exit 4")
		return
	}

	account := ctx.AccessSCAccount()

	// find the winning amount and determine respective ownerFee
	winningAmount := int64(0)
	for _, bi := range ai.Bids {
		if bi.Total > winningAmount {
			winningAmount = bi.Total
		}
	}

	var winner *BidInfo
	var winnerIndex int

	// SC owner takes OwnerMargin (promille) fee from either minimum bid or from winning sum but not less than 1i
	ownerFee := (ai.MinimumBid * ai.OwnerMargin) / 1000
	if ownerFee < 1 {
		ownerFee = 1
	}

	// find the winner (if any). Take first if equal sums
	// minimum bid is always positive, at least 1 iota per colored token
	if winningAmount >= ai.MinimumBid {
		// there's winner. Select it.
		// Fee is re-calculated according to the winning sum
		ownerFee = (winningAmount * ai.OwnerMargin) / 1000
		if ownerFee < 1 {
			ownerFee = 1
		}

		winners := make([]*BidInfo, 0)
		for _, bi := range ai.Bids {
			if bi.Total == winningAmount {
				winners = append(winners, bi)
			}
		}
		sort.Slice(winners, func(i, j int) bool {
			return winners[i].When < winners[j].When
		})
		winner = winners[0]
		for i, bi := range ai.Bids {
			if bi == winner {
				winnerIndex = i
				break
			}
		}
	}

	feeTaken := ctx.AccessSCAccount().HarvestFees(ownerFee - 1)
	ctx.Publishf("finalizeAuction: harvesting SC owner fee: %d (+1 self request token left in SC)", feeTaken)

	if winner != nil {
		// send sold tokens to the winner
		account.MoveTokens(&ai.Bids[winnerIndex].Bidder, &ai.Color, ai.NumTokens)
		// send winning amount and return deposit sum less fees to the owner of the auction
		account.MoveTokens(&ai.AuctionOwner, &balance.ColorIOTA, winningAmount+ai.TotalDeposit-ownerFee)

		for i, bi := range ai.Bids {
			if i != winnerIndex {
				// return staked sum to the non-winner
				account.MoveTokens(&bi.Bidder, &balance.ColorIOTA, bi.Total)
			}
		}
		ctx.Publishf("finalizeAuction: winner is %s, winning amount = %d", winner.Bidder.String(), winner.Total)
	} else {
		// return unsold tokens to auction owner
		if account.MoveTokens(&ai.AuctionOwner, &ai.Color, ai.NumTokens) {
			ctx.Publishf("returned unsold tokens to auction owner. %s: %d", ai.Color.String(), ai.NumTokens)
		}

		// return deposit less fees less 1 iota
		if account.MoveTokens(&ai.AuctionOwner, &balance.ColorIOTA, ai.TotalDeposit-ownerFee) {
			ctx.Publishf("returned deposit less fees: %d", ai.TotalDeposit-ownerFee)
		}

		// return bids to bidders
		for _, bi := range ai.Bids {
			if account.MoveTokens(&bi.Bidder, &balance.ColorIOTA, bi.Total) {
				ctx.Publishf("returned bid to bidder: %d -> %s", bi.Total, bi.Bidder.String())
			} else {
				avail := ctx.AccessSCAccount().AvailableBalance(&balance.ColorIOTA)
				ctx.Publishf("failed to return bid to bidder: %d -> %s. Available: %d", bi.Total, bi.Bidder.String(), avail)
			}
		}
		ctx.Publishf("finalizeAuction: winner wasn't selected out of %d bids", len(ai.Bids))
	}

	// delete auction record
	auctDict.DelAt(col.Bytes())

	ctx.Publishf("finalizeAuction: success. Auction: '%s'", ai.Description)
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

// refundFromRequest returns all iotas tokens to the sender minus sunkFee
func refundFromRequest(ctx vmtypes.Sandbox, color *balance.Color, harvest int64) {
	account := ctx.AccessSCAccount()
	ctx.AccessSCAccount().HarvestFeesFromRequest(harvest)
	available := account.AvailableBalanceFromRequest(color)
	sender := ctx.AccessRequest().Sender()
	ctx.AccessSCAccount().HarvestFeesFromRequest(harvest)
	account.MoveTokensFromRequest(&sender, color, available)

}
