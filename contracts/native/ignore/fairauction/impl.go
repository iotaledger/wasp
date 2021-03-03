// +build ignore

// hard coded implementation of the FairAuction smart contract
// The auction dApp is automatically run by committee, a distributed market for colored tokens
package fairauction

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/mr-tron/base58"
)

// program has is an id of the program
const ProgramHash = "4NbQFgvnsfgE3n9ZhtJ3p9hWZzfYUEDHfKU93wp8UowB"
const Description = "FairAuction, a PoC smart contract"

// implement Processor and EntryPoint interfaces

type fairAuctionProcessor map[coretypes.Hname]fairAuctionEntryPoint

type fairAuctionEntryPoint func(ctx coretypes.Sandbox) error

var (
	RequestStartAuction    = coretypes.Hn("startAuction")
	RequestFinalizeAuction = coretypes.Hn("finalizeAuction")
	RequestPlaceBid        = coretypes.Hn("placeBid")
	RequestSetOwnerMargin  = coretypes.Hn("setOwnerMargin")
)

// the processor is a map of entry points
var entryPoints = fairAuctionProcessor{
	RequestStartAuction:    startAuction,
	RequestFinalizeAuction: finalizeAuction,
	RequestPlaceBid:        placeBid,
	RequestSetOwnerMargin:  setOwnerMargin,
}

// string constants for request arguments and state variable names
const (
	// request vars
	VarReqAuctionColor                = "color"
	VarReqStartAuctionDescription     = "dscr"
	VarReqStartAuctionDurationMinutes = "duration"
	VarReqStartAuctionMinimumBid      = "minimum" // in iotas
	VarReqOwnerMargin                 = "ownerMargin"

	// state vars
	VarStateAuctions            = "auctions"
	VarStateLog                 = "log"
	VarStateOwnerMarginPromille = "ownerMargin" // owner margin in percents
)

const (
	// minimum duration of auction
	MinAuctionDurationMinutes = 1
	MaxAuctionDurationMinutes = 120 // max 2 hours

	// default duration of the auction
	AuctionDurationDefaultMinutes = 60
	// Owner of the smart contract takes %% from the winning bid. The default, min, max
	OwnerMarginDefault = 50  // 5%
	OwnerMarginMin     = 5   // minimum 0.5%
	OwnerMarginMax     = 100 // max 10%
	MaxDescription     = 150
)

// validating constants at node boot
func init() {
	if OwnerMarginMax > 1000 ||
		OwnerMarginMin < 0 ||
		OwnerMarginDefault < OwnerMarginMin ||
		OwnerMarginDefault > OwnerMarginMax ||
		OwnerMarginMin > OwnerMarginMax {
		panic("wrong constants")
	}
}

// statical link point to the Wasp node
func GetProcessor() coretypes.Processor {
	return entryPoints
}

func (v fairAuctionProcessor) GetDescription() string {
	return "FairAuction hard coded smart contract program"
}

func (v fairAuctionProcessor) GetEntryPoint(code coretypes.Hname) (coretypes.EntryPoint, bool) {
	f, ok := v[code]
	return f, ok
}

func (ep fairAuctionEntryPoint) Call(ctx coretypes.Sandbox) (dict.Dict, error) {
	err := ep(ctx)
	if err != nil {
		ctx.Event(fmt.Sprintf("error %v", err))
	}
	return nil, err
}

// TODO
func (ep fairAuctionEntryPoint) IsView() bool {
	return false
}

// TODO
func (ep fairAuctionEntryPoint) CallView(ctx coretypes.SandboxView) (dict.Dict, error) {
	panic("implement me")
}

// AuctionInfo describes active auction
type AuctionInfo struct {
	// color of the tokens for sale. Max one auction per color at same time is allowed
	// all tokens are being sold as one lot
	Color balance.Color
	// number of tokens for sale
	NumTokens int64
	// minimum bid. Set by the auction initiator
	MinimumBid int64
	// any text, like "AuctionOwner of the token have a right to call me for a date". Set by auction initiator
	Description string
	// timestamp when auction started
	WhenStarted int64
	// duration of the auctions in minutes. Should be >= MinAuctionDurationMinutes
	DurationMinutes int64
	// address which issued StartAuction transaction
	AuctionOwner coretypes.AgentID
	// total deposit by the auction owner. Iotas sent by the auction owner together with the tokens for sale in the same
	// transaction.
	TotalDeposit int64
	// AuctionOwner's margin in promilles, taken at the moment of creation of smart contract
	OwnerMargin int64
	// list of bids to the auction
	Bids []*BidInfo
}

// BidInfo represents one bid to the auction
type BidInfo struct {
	// total sum of the bid = total amount of iotas available in the request - 1 - SC reward - ServiceFeeBid
	// the total is a cumulative sum of all bids from the same bidder
	Total int64
	// originator of the bid
	Bidder coretypes.AgentID
	// timestamp Unix nano
	When int64
}

func (ai *AuctionInfo) SumOfBids() int64 {
	sum := int64(0)
	for _, bid := range ai.Bids {
		sum += bid.Total
	}
	return sum
}

func (ai *AuctionInfo) WinningBid() *BidInfo {
	var winner *BidInfo
	for _, bi := range ai.Bids {
		if bi.Total < ai.MinimumBid {
			continue
		}
		if winner == nil || bi.WinsAgainst(winner) {
			winner = bi
		}
	}
	return winner
}

func (ai *AuctionInfo) Due() int64 {
	return ai.WhenStarted + ai.DurationMinutes*time.Minute.Nanoseconds()
}

func (bi *BidInfo) WinsAgainst(other *BidInfo) bool {
	if bi.Total < other.Total {
		return false
	}
	if bi.Total > other.Total {
		return true
	}
	return bi.When < other.When
}

// startAuction processes the StartAuction request
// Arguments:
// - VarReqAuctionColor: color of the tokens for sale
// - VarReqStartAuctionDescription: description of the lot
// - VarReqStartAuctionMinimumBid: minimum price for the whole lot
// - VarReqStartAuctionDurationMinutes: duration of auction
// Request transaction must contain at least number of iotas >= of current owner margin from the minimum bid
// (not including node reward with request token)
// Tokens for sale must be included into the request transaction
func startAuction(ctx coretypes.Sandbox) error {
	ctx.Event("startAuction begin")
	params := ctx.Params()

	sender := ctx.Caller()

	// check how many iotas the request contains
	totalDeposit := ctx.IncomingTransfer().Balance(balance.ColorIOTA)
	if totalDeposit < 1 {
		// it is expected at least 1 iota in deposit
		// this 1 iota is needed as a "operating capital for the time locked request to itself"
		// refund iotas
		refundFromRequest(ctx, &balance.ColorIOTA, 1)

		return fmt.Errorf("startAuction: exit 0: must be at least 1i in deposit")
	}

	// take current setting of the smart contract owner margin
	ownerMargin := GetOwnerMarginPromille(codec.DecodeInt64(ctx.State().MustGet(VarStateOwnerMarginPromille)))

	// determine color of the token for sale
	colh, ok, err := codec.DecodeString(params.MustGet(VarReqAuctionColor))
	if err != nil || !ok {
		// incorrect request arguments, colore for sale is not determined
		// refund half of the deposit in iotas
		refundFromRequest(ctx, &balance.ColorIOTA, totalDeposit/2)

		return fmt.Errorf("startAuction: exit 1")
	}
	colorh, err := base58.Decode(colh)
	if err != nil {
		return fmt.Errorf("startAuction: exit 1.1")
	}
	colorForSale, _, err := balance.ColorFromBytes(colorh)
	if err != nil {
		return fmt.Errorf("startAuction: exit 1.2")
	}
	if colorForSale == balance.ColorIOTA || colorForSale == balance.ColorNew {
		// reserved color code are not allowed
		// refund half
		refundFromRequest(ctx, &balance.ColorIOTA, totalDeposit/2)

		return fmt.Errorf("startAuction: exit 2")
	}

	// determine amount of colored tokens for sale. They must be in the outputs of the request transaction
	tokensForSale := ctx.IncomingTransfer().Balance(colorForSale)
	if tokensForSale == 0 {
		// no tokens transferred. Refund half of deposit
		refundFromRequest(ctx, &balance.ColorIOTA, totalDeposit/2)

		return fmt.Errorf("startAuction exit 3: no tokens for sale")
	}

	// determine minimum bid
	minimumBid, _, err := codec.DecodeInt64(params.MustGet(VarReqStartAuctionMinimumBid))
	if err != nil {
		// wrong argument. Hard reject, no refund

		return fmt.Errorf("startAuction: exit 4")
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

		return fmt.Errorf("startAuction: not enough iotas for the fee. Expected %d, got %d", expectedDeposit, totalDeposit)
	}

	// determine duration of the auction. Take default if no set in request and ensure minimum
	duration, ok, err := codec.DecodeInt64(params.MustGet(VarReqStartAuctionDurationMinutes))
	if err != nil {
		// fatal error
		return fmt.Errorf("!!! internal error")
	}
	if !ok {
		duration = AuctionDurationDefaultMinutes
	}
	if duration < MinAuctionDurationMinutes {
		duration = MinAuctionDurationMinutes
	}
	if duration > MaxAuctionDurationMinutes {
		duration = MaxAuctionDurationMinutes
	}

	// read description text from the request
	description, ok, err := codec.DecodeString(params.MustGet(VarReqStartAuctionDescription))
	if err != nil {
		return fmt.Errorf("!!! internal error")
	}
	if !ok {
		description = "N/A"
	}
	description = util.GentleTruncate(description, MaxDescription)

	// find out if auction for this color already exist in the dictionary
	auctions := collections.NewMap(ctx.State(), VarStateAuctions)
	if b := auctions.MustGetAt(colorForSale.Bytes()); b != nil {
		// auction already exists. Ignore sale auction.
		// refund iotas less fee
		refundFromRequest(ctx, &balance.ColorIOTA, expectedDeposit/2)
		// return all tokens for sale
		refundFromRequest(ctx, &colorForSale, 0)

		return fmt.Errorf("startAuction: exit 6")
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
	auctions.MustSetAt(colorForSale.Bytes(), aiData)

	ctx.Event(fmt.Sprintf("New auction record. color: %s, numTokens: %d, minBid: %d, ownerMargin: %d duration %d minutes",
		colorForSale.String(), tokensForSale, minimumBid, ownerMargin, duration))

	// prepare and send request FinalizeAuction to self time-locked for the duration
	// the FinalizeAuction request will be time locked for the duration and then auction will be run
	args := dict.FromGoMap(map[kv.Key][]byte{
		VarReqAuctionColor: codec.EncodeString(colorForSale.String()),
	})
	ctx.PostRequest(coretypes.PostRequestParams{
		TargetContractID: ctx.ContractID(),
		EntryPoint:       RequestFinalizeAuction,
		TimeLock:         uint32(duration * 60),
		Params:           args,
	})
	//logToSC(ctx, fmt.Sprintf("start auction. For sale %d tokens of color %s. Minimum bid: %di. Duration %d minutes",
	//	tokensForSale, colorForSale.String(), minimumBid, duration))

	ctx.Event(fmt.Sprintf("startAuction: success. Auction: '%s', color: %s, duration: %d",
		description, colorForSale.String(), duration))

	return nil
}

// placeBid is a request to place a bid in the auction for the particular color
// The request transaction must contain at least:
// - 1 request token + Bid/rise amount
// In case it is not the first bid by this bidder, respective iotas are treated as
// a rise of the bid and are added to the total
// Arguments:
// - VarReqAuctionColor: color of the tokens for sale
func placeBid(ctx coretypes.Sandbox) error {
	ctx.Event("placeBid: begin")
	params := ctx.Params()
	// all iotas in the request transaction are considered a bid/rise sum
	// it also means several bids can't be placed in the same transaction <-- TODO generic solution for it
	bidAmount := ctx.IncomingTransfer().Balance(balance.ColorIOTA)
	if bidAmount == 0 {
		// no iotas sent
		return fmt.Errorf("placeBid: exit 0")
	}

	// determine color of the bid
	colh, ok, err := codec.DecodeString(params.MustGet(VarReqAuctionColor))
	if err != nil {
		// inconsistency. return all?
		return fmt.Errorf("placeBid: exit 1")
	}
	if !ok {
		// missing argument
		refundFromRequest(ctx, &balance.ColorIOTA, 0)
		return fmt.Errorf("placeBid: exit 2")
	}

	colorh, err := base58.Decode(colh)
	if err != nil {
		return fmt.Errorf("startAuction: exit 1.1")
	}
	col, _, err := balance.ColorFromBytes(colorh)
	if err != nil {
		return fmt.Errorf("startAuction: exit 1.2")
	}
	if col == balance.ColorIOTA || col == balance.ColorNew {
		// reserved color not allowed. Incorrect arguments
		refundFromRequest(ctx, &balance.ColorIOTA, 0)
		return fmt.Errorf("placeBid: exit 3")
	}

	// find the auction
	auctions := collections.NewMap(ctx.State(), VarStateAuctions)
	data := auctions.MustGetAt(col.Bytes())
	if data == nil {
		// no such auction. refund everything
		refundFromRequest(ctx, &balance.ColorIOTA, 0)
		return fmt.Errorf("placeBid: exit 4")
	}
	// unmarshal auction data
	ai := &AuctionInfo{}
	if err := ai.Read(bytes.NewReader(data)); err != nil {
		// internal error
		return fmt.Errorf("placeBid: exit 6")
	}
	// determine the sender of the bid
	sender := ctx.Caller()

	// find bids of this bidder in the auction
	var bi *BidInfo
	for _, bitmp := range ai.Bids {
		if bitmp.Bidder == sender {
			bi = bitmp
			break
		}
	}
	if bi == nil {
		// first bid by the bidder. Create new bid record
		ai.Bids = append(ai.Bids, &BidInfo{
			Total:  bidAmount,
			Bidder: sender,
			When:   ctx.GetTimestamp(),
		})
		//logToSC(ctx, fmt.Sprintf("place bid. Auction color %s, total %di", col.String(), bidAmount))
	} else {
		// bidder has bid already. Treated it as a rise
		bi.Total += bidAmount
		bi.When = ctx.GetTimestamp()

		//logToSC(ctx, fmt.Sprintf("rise bid. Auction color %s, total %di", col.String(), bi.Total))
	}
	// marshal the whole auction info and save it into the state (the dictionary of auctions)
	data = util.MustBytes(ai)
	auctions.MustSetAt(col.Bytes(), data)

	ctx.Event(fmt.Sprintf("placeBid: success. Auction: '%s'", ai.Description))
	return nil
}

// finalizeAuction selects the winner and sends tokens to him.
// returns bid amounts to other bidders.
// The request is time locked for the period of the auction. It won't be executed if sent
// not by the smart contract instance itself
// Arguments:
// - VarReqAuctionColor: color of the auction
func finalizeAuction(ctx coretypes.Sandbox) error {
	ctx.Event("finalizeAuction begin")
	params := ctx.Params()

	scAddr := coretypes.NewAgentIDFromContractID(ctx.ContractID())
	if ctx.Caller() != scAddr {
		// finalizeAuction request can only be sent by the smart contract to itself. Otherwise it is NOP
		return fmt.Errorf("attempt of unauthorized assess")
	}

	// determine color of the auction to finalize
	colh, ok, err := codec.DecodeString(params.MustGet(VarReqAuctionColor))
	if err != nil || !ok {
		// wrong request arguments
		// internal error. Refund completely?
		return fmt.Errorf("finalizeAuction: exit 1")
	}
	colorh, err := base58.Decode(colh)
	if err != nil {
		return fmt.Errorf("startAuction: exit 1.1")
	}
	col, _, err := balance.ColorFromBytes(colorh)
	if err != nil {
		return fmt.Errorf("startAuction: exit 1.2")
	}
	if col == balance.ColorIOTA || col == balance.ColorNew {
		// inconsistency
		return fmt.Errorf("finalizeAuction: exit 2")
	}

	// find the record of the auction by color
	auctDict := collections.NewMap(ctx.State(), VarStateAuctions)
	data := auctDict.MustGetAt(col.Bytes())
	if data == nil {
		// auction with this color does not exist. Inconsistency
		return fmt.Errorf("finalizeAuction: exit 3")
	}

	// decode the Action record
	ai := &AuctionInfo{}
	if err := ai.Read(bytes.NewReader(data)); err != nil {
		// internal error. Refund completely?
		return fmt.Errorf("finalizeAuction: exit 4")
	}

	// find the winning amount and determine respective ownerFee
	winningAmount := int64(0)
	for _, bi := range ai.Bids {
		if bi.Total > winningAmount {
			winningAmount = bi.Total
		}
	}

	var winner *BidInfo
	//var winnerIndex int

	// SC owner takes OwnerMargin (promille) fee from either minimum bid or from winning sum but not less than 1i
	ownerFee := (ai.MinimumBid * ai.OwnerMargin) / 1000
	if ownerFee < 1 {
		ownerFee = 1
	}

	// find the winner (if any). Take first if equal sums
	// minimum bid is always positive, at least 1 iota per colored token
	if winningAmount >= ai.MinimumBid {
		// there's winner. Select it.
		// OwnerFee is re-calculated according to the winning sum
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
		//for i, bi := range ai.Bids {
		//	if bi == winner {
		//		//winnerIndex = i
		//		break
		//	}
		//}
	}

	// take fee for the smart contract owner TODO
	//feeTaken := ctx.AccessSCAccount().HarvestFees(ownerFee - 1)
	//ctx.Event(fmt.Sprintf("finalizeAuction: harvesting SC owner fee: %d (+1 self request token left in SC)", feeTaken))

	if winner != nil {
		// send sold tokens to the winner
		//
		//ctx.MoveTokens(ai.Bids[winnerIndex].Bidder, ai.Color, ai.NumTokens)
		//// send winning amount and return deposit sum less fees to the owner of the auction
		//ctx.MoveTokens(ai.AuctionOwner, balance.ColorIOTA, winningAmount+ai.TotalDeposit-ownerFee)
		//
		//for i, bi := range ai.Bids {
		//	if i != winnerIndex {
		//		// return staked sum to the non-winner
		//		ctx.MoveTokens(bi.Bidder, balance.ColorIOTA, bi.Total)
		//	}
		//}
		//logToSC(ctx, fmt.Sprintf("close auction. Color: %s. Winning bid: %di", col.String(), winner.Total))

		ctx.Event(fmt.Sprintf("finalizeAuction: winner is %s, winning amount = %d", winner.Bidder.String(), winner.Total))
	} else {
		// return unsold tokens to auction owner

		//if ctx.MoveTokens(ai.AuctionOwner, ai.Color, ai.NumTokens) {
		//	ctx.Event(fmt.Sprintf("returned unsold tokens to auction owner. %s: %d", ai.Color.String(), ai.NumTokens))
		//}
		//
		//// return deposit less fees less 1 iota
		//if ctx.MoveTokens(ai.AuctionOwner, balance.ColorIOTA, ai.TotalDeposit-ownerFee) {
		//	ctx.Event(fmt.Sprintf("returned deposit less fees: %d", ai.TotalDeposit-ownerFee))
		//}

		// return bids to bidders
		//for _, bi := range ai.Bids {
		//	if ctx.MoveTokens(bi.Bidder, balance.ColorIOTA, bi.Total) {
		//		ctx.Event(fmt.Sprintf("returned bid to bidder: %d -> %s", bi.Total, bi.Bidder.String()))
		//	} else {
		//		avail := ctx.Balance(balance.ColorIOTA)
		//		ctx.Event(fmt.Sprintf("failed to return bid to bidder: %d -> %s. Available: %d", bi.Total, bi.Bidder.String(), avail))
		//	}
		//}
		//logToSC(ctx, fmt.Sprintf("close auction. Color: %s. No winner.", col.String()))

		ctx.Event(fmt.Sprintf("finalizeAuction: winner wasn't selected out of %d bids", len(ai.Bids)))
	}

	// delete auction record
	auctDict.MustDelAt(col.Bytes())

	ctx.Event(fmt.Sprintf("finalizeAuction: success. Auction: '%s'", ai.Description))
	return nil
}

// setOwnerMargin is a request to set the service fee to place a bid
// Arguments:
// - VarReqOwnerMargin: the margin value in promilles
func setOwnerMargin(ctx coretypes.Sandbox) error {
	ctx.Event("setOwnerMargin: begin")
	params := ctx.Params()

	// TODO refactor to the new account system
	//if ctx.Caller() != *ctx.OriginatorAddress() {
	//	// not authorized
	//	return fmt.Errorf("setOwnerMargin: not authorized")
	//}
	margin, ok, err := codec.DecodeInt64(params.MustGet(VarReqOwnerMargin))
	if err != nil || !ok {
		return fmt.Errorf("setOwnerMargin: exit 1")
	}
	if margin < OwnerMarginMin {
		margin = OwnerMarginMin
	} else if margin > OwnerMarginMax {
		margin = OwnerMarginMax
	}
	ctx.State().Set(VarStateOwnerMarginPromille, codec.EncodeInt64(margin))
	ctx.Event(fmt.Sprintf("setOwnerMargin: success. ownerMargin set to %d%%", margin/10))
	return nil
}

// TODO implement universal 'refund' function to be used in rollback situations
// refundFromRequest returns all tokens of the given color to the sender minus sunkFee
func refundFromRequest(ctx coretypes.Sandbox, color *balance.Color, harvest int64) {
	// TODO
	//account := ctx.AccessSCAccount()
	//ctx.AccessSCAccount().HarvestFeesFromRequest(harvest)
	//available := account.AvailableBalanceFromRequest(color)
	//sender := ctx.Caller()
	//ctx.AccessSCAccount().HarvestFeesFromRequest(harvest)
	//account.MoveTokensFromRequest(&sender, color, available)
}
