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
	RequestInitSC = sctransaction.RequestCode(uint16(0))
	// request to start an auction. Args:
	// - NFT color (must be unique among active auctions)
	// - description of the lot
	// - minimum price for tokens for sale (all of them)
	// - duration of auction
	// Request transaction must contain at least 1 token with the color plus N iotas service fee
	RequestStartAuction = sctransaction.RequestCode(uint16(1))
	// remove from auction (service fee is lost)
	// args: color
	RequestRemoveAuction = sctransaction.RequestCode(uint16(2))
	// select the highest bid, send the NTFs, return bids minus minimum service fee
	// it has effect only if sent by SC to itself
	// it is timelocked for the duration of the auction
	// args: auction color
	RequestFinalizeAuction = sctransaction.RequestCode(uint16(3))
	// admin (protected) requests to set minimum fees
	RequestSetServiceFeeAuction = sctransaction.RequestCode(uint16(4) | sctransaction.RequestCodeProtected)
	RequestSetServiceFeeBid     = sctransaction.RequestCode(uint16(5) | sctransaction.RequestCodeProtected)
	// place bid. Args:
	// - color
	// Must contain at least ServiceFeeBid of iotas. It will be taken anyway. The rest of iotas will
	// be treated as bid sum. It will be returned to sender's address if bid won't win
	RequestPlaceBid = sctransaction.RequestCode(uint16(6))
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
	// tumestamp when auction started
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
func startAuction(ctx vmtypes.Sandbox) {
	ctx.Publish("startAuction begin")

	// find out who starts the action
	sender := takeSender(ctx)
	if sender == nil {
		// wrong transaction, hardReject (nothing is processed and nothing is refunded)
		return
	}
	// get currently set service fee values
	feeAuction, _ := getFeeValues(ctx)

	// validate request arguments
	reqArgs := ctx.AccessRequest().Args()

	// determine color of the token
	colh, ok, err := reqArgs.GetHashValue(VarReqAuctionColor)
	if err != nil || !ok {
		// incorrect request arguments
		gentlyRejectRequest(ctx, sender, feeAuction)
		return
	}
	colorForSale := (balance.Color)(*colh)
	if colorForSale == balance.ColorIOTA || colorForSale == balance.ColorNew {
		// reserved color not allowed
		gentlyRejectRequest(ctx, sender, feeAuction)
		return
	}

	// determine amount of colored tokens sent for sale
	tokensForSale := ctx.AccessOwnAccount().AvailableBalanceFromRequest(&colorForSale)
	if tokensForSale == 0 {
		// no tokens transferred
		gentlyRejectRequest(ctx, sender, feeAuction)
		return
	}

	// check if enough iotas for the auction fee
	if ctx.AccessOwnAccount().AvailableBalanceFromRequest(&balance.ColorIOTA) < feeAuction {
		// not enough fees
		// return tokens for sale and reject transaction
		ctx.AccessOwnAccount().MoveTokensFromRequest(sender, &colorForSale, tokensForSale)
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
	description, ok, err := reqArgs.GetString(VarReqStartAuctionDescription)
	if err != nil {
		return
	}
	if !ok {
		description = "N/A"
	}

	// find out if auction for this color already exist in the dictionary
	auctDict := ctx.AccessState().GetDictionary(VarStateAuctions)
	if b := auctDict.GetAt(colorForSale.Bytes()); b != nil {
		// auction already exists. Ignore sale auction. Return tokens for sale
		ctx.AccessOwnAccount().MoveTokensFromRequest(sender, &colorForSale, tokensForSale)
		gentlyRejectRequest(ctx, sender, feeAuction)

		return
	}
	// create record for the new auction in the dictionary
	aiData := util.MustBytes(&AuctionInfo{
		Color:           colorForSale,
		NumTokens:       tokensForSale,
		Description:     description,
		WhenStarted:     ctx.GetTimestamp(),
		DurationMinutes: duration,
		Owner:           *sender,
	})
	auctDict.SetAt(colorForSale.Bytes(), aiData)

	// prepare and send timelocked for the duration FinalizeAuction request to itself
	args := kv.NewMap()
	args.Codec().SetHashValue(VarReqAuctionColor, (*hashing.HashValue)(&colorForSale))
	ctx.SendRequestToSelfWithDelay(RequestFinalizeAuction, args, uint32(duration*60))

	ctx.Publish("startAuction end")
}

// removeAuction processes remove auction request
func removeAuction(ctx vmtypes.Sandbox) {
	ctx.Publish("removeAuction begin")

	reqArgs := ctx.AccessRequest().Args()

	// determine color of the auction
	colh, ok, err := reqArgs.GetHashValue(VarReqAuctionColor)
	if err != nil || !ok {
		// incorrect request arguments
		return
	}
	col := (balance.Color)(*colh)
	if col == balance.ColorIOTA || col == balance.ColorNew {
		// reserved color not allowed
		return
	}

	// find the record for the auction by color
	auctDict := ctx.AccessState().GetDictionary(VarStateAuctions)
	data := auctDict.GetAt(col.Bytes())
	if data == nil {
		// nothing to remove
		return
	}
	// decode the record
	ai := &AuctionInfo{}
	if err := ai.Read(bytes.NewReader(data)); err != nil {
		// internal error
		return
	}
	if !ctx.AccessRequest().IsAuthorisedByAddress(&ai.Owner) {
		// not authorised
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
// return bid amounts to other bidders
func finalizeAuction(ctx vmtypes.Sandbox) {
	ctx.Publish("finalizeAuction begin")

	accessReq := ctx.AccessRequest()
	if !accessReq.IsAuthorisedByAddress(ctx.GetOwnAddress()) {
		// finalizeAuction request can be sent only by the smart contract to itself
		return
	}
	reqArgs := accessReq.Args()

	// determine color of the auction to finalize
	colh, ok, err := reqArgs.GetHashValue(VarReqAuctionColor)
	if err != nil || !ok {
		// incorrect request arguments
		return
	}
	col := (balance.Color)(*colh)
	if col == balance.ColorIOTA || col == balance.ColorNew {
		// reserved color not allowed
		return
	}

	// find the record for the auction by color
	auctDict := ctx.AccessState().GetDictionary(VarStateAuctions)
	data := auctDict.GetAt(col.Bytes())
	if data == nil {
		// auction with this color does not exist, probably removed
		return
	}

	account := ctx.AccessOwnAccount()
	// decode the Action record
	ai := &AuctionInfo{}
	if err := ai.Read(bytes.NewReader(data)); err != nil {
		// internal error
		return
	}
	if len(ai.Bids) == 0 {
		// no bids
		// return tokens to owner
		account.MoveTokens(&ai.Owner, &ai.Color, ai.NumTokens)
		// delete auction record
		auctDict.DelAt(col.Bytes())
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
		return
	}
	// send bid sum to the owner of the auction
	account.MoveTokens(&ai.Owner, &balance.ColorIOTA, winningBid)
	// send tokens to the winner
	account.MoveTokens(&winner.Bidder, &ai.Color, ai.NumTokens)
	// return bids to losing bidders

	for i, bi := range ai.Bids {
		if i != winnerIndex {
			account.MoveTokens(&bi.Bidder, &balance.ColorIOTA, bi.Total)
		}
	}

	// delete auction record
	auctDict.DelAt(col.Bytes())

	ctx.Publish("finalizeAuction success")
}

func setServiceFeeAuction(ctx vmtypes.Sandbox) {
	ctx.Publish("setServiceFeeAuction")
}

func setServiceFeeBid(ctx vmtypes.Sandbox) {
	ctx.Publish("setServiceFeeBid")
}

func placeBid(ctx vmtypes.Sandbox) {
	bidSum := ctx.AccessOwnAccount().AvailableBalanceFromRequest(&balance.ColorIOTA)
	if bidSum == 0 {
		// no iotas to bid. Ignore
		return
	}
	reqArgs := ctx.AccessRequest().Args()
	colh, ok, err := reqArgs.GetHashValue(VarReqAuctionColor)
	if err != nil || !ok {
		// incorrect request arguments
		// return iotas minus service fee
		return
	}
	col := (balance.Color)(*colh)
	if col == balance.ColorIOTA || col == balance.ColorNew {
		// reserved color not allowed
		return
	}
	auctDict := ctx.AccessState().GetDictionary(VarStateAuctions)
	data := auctDict.GetAt(col[:])
	if data == nil {
		// no such auction
		return
	}
	ai := &AuctionInfo{}
	if err := ai.Read(bytes.NewReader(data)); err != nil {
		ctx.Panic("internal inconsistency")
		return
	}
	senders := ctx.AccessRequest().Senders()
	if len(senders) != 1 {
		return
	}
	sender := senders[0]
	var bi *BidInfo
	for _, bit := range ai.Bids {
		if bit.Bidder == sender {
			bi = bit
			break
		}
	}
	if bi == nil {
		ai.Bids = append(ai.Bids, &BidInfo{
			Total:  bidSum,
			Bidder: sender,
		})
	} else {
		bi.Total += bidSum
	}
	data = util.MustBytes(ai)
	auctDict.SetAt(col[:], data)
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

// gentlyRejectRequest returns all tokens to the sender minus sunkFee
func gentlyRejectRequest(ctx vmtypes.Sandbox, sender *address.Address, sunkFee int64) {

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
