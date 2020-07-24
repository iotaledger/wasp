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
	"io"
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
	VarStateAuctions = "auctions"
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
	Color       balance.Color
	NumTokens   int64
	MinimumBid  int64
	Description string
	WhenPlaced  int64
	Duration    int64
	Owner       address.Address
	Bids        []*BidInfo
}

type BidInfo struct {
	Total  int64
	Bidder address.Address
}

func initSC(ctx vmtypes.Sandbox) {
}

func startAuction(ctx vmtypes.Sandbox) {
	ctx.Publish("startAuction")

	// validate request arguments
	reqArgs := ctx.AccessRequest().Args()
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
	duration, ok, err := reqArgs.GetInt64(VarReqStartAuctionDurationMinutes)
	if err != nil || !ok {
		return
	}
	description, ok, err := reqArgs.GetString(VarReqStartAuctionDescription)
	if err != nil {
		return
	}
	if !ok {
		description = "N/A"
	}

	// take input addresses of the request transaction. Must be exactly 1 otherwise.
	// Theoretically the transaction may have several addresses in inputs, then it is ignored
	senders := ctx.AccessRequest().Senders()
	if len(senders) != 1 {
		return
	}
	sender := senders[0]

	numTokens := ctx.AccessOwnAccount().AvailableBalanceFromRequest(&col)
	if numTokens == 0 {
		// no tokens to sale. Ignore
		return
	}
	auctDict := ctx.AccessState().GetDictionary(VarStateAuctions)
	if b := auctDict.GetAt(col.Bytes()); b != nil {
		// already exists. Ignore sale auction
		// TODO return colored tokens
		return
	}
	aiData := util.MustBytes(&AuctionInfo{
		Color:       col,
		NumTokens:   numTokens,
		Description: description,
		WhenPlaced:  ctx.GetTimestamp(),
		Owner:       sender,
	})
	auctDict.SetAt(col.Bytes(), aiData)

	args := kv.NewMap()
	args.Codec().SetHashValue(VarReqAuctionColor, (*hashing.HashValue)(&col))

	// send timelocked request to run an auction
	ctx.SendRequestToSelfWithDelay(RequestFinalizeAuction, args, uint32(duration*60))

	ctx.Publish("startAuction")
}

func removeAuction(ctx vmtypes.Sandbox) {
	reqArgs := ctx.AccessRequest().Args()
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
	auctDict := ctx.AccessState().GetDictionary(VarStateAuctions)
	if b := auctDict.GetAt(col.Bytes()); b != nil {
		auctDict.DelAt(col.Bytes())
		return
	}
	ctx.Publish("removeAuction")
}

func finalizeAuction(ctx vmtypes.Sandbox) {
	ctx.Publish("finalizeAuction")

	// TODO select thw winner. If two bids are equal, select the one with smaller timestamp
	// send colored tokens to the winner
	// return bids to others (but not the fee)
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

// ser/de ActionInfo

func (ai *AuctionInfo) Write(w io.Writer) error {
	if _, err := w.Write(ai.Color[:]); err != nil {
		return err
	}
	if err := util.WriteInt64(w, ai.NumTokens); err != nil {
		return err
	}
	if err := util.WriteInt64(w, ai.MinimumBid); err != nil {
		return err
	}
	if err := util.WriteString16(w, ai.Description); err != nil {
		return err
	}
	if err := util.WriteInt64(w, ai.WhenPlaced); err != nil {
		return err
	}
	if err := util.WriteInt64(w, ai.Duration); err != nil {
		return err
	}
	if _, err := w.Write(ai.Owner[:]); err != nil {
		return err
	}
	if err := util.WriteUint16(w, uint16(len(ai.Bids))); err != nil {
		return err
	}
	for _, bi := range ai.Bids {
		if err := bi.Write(w); err != nil {
			return err
		}
	}
	return nil
}

func (ai *AuctionInfo) Read(r io.Reader) error {
	var err error
	if err = util.ReadColor(r, &ai.Color); err != nil {
		return err
	}
	if err = util.ReadInt64(r, &ai.NumTokens); err != nil {
		return err
	}
	if err = util.ReadInt64(r, &ai.MinimumBid); err != nil {
		return err
	}
	if ai.Description, err = util.ReadString16(r); err != nil {
		return err
	}
	if err = util.ReadInt64(r, &ai.WhenPlaced); err != nil {
		return err
	}
	if err = util.ReadInt64(r, &ai.Duration); err != nil {
		return err
	}
	if err = util.ReadAddress(r, &ai.Owner); err != nil {
		return err
	}
	var size uint16
	if err := util.ReadUint16(r, &size); err != nil {
		return err
	}
	ai.Bids = make([]*BidInfo, size)
	for i := range ai.Bids {
		ai.Bids[i] = &BidInfo{}
		if err := ai.Bids[i].Read(r); err != nil {
			return err
		}
	}
	return nil
}

func (bi *BidInfo) Write(w io.Writer) error {
	if err := util.WriteInt64(w, bi.Total); err != nil {
		return err
	}
	if _, err := w.Write(bi.Bidder[:]); err != nil {
		return err
	}
	return nil
}

func (bi *BidInfo) Read(r io.Reader) error {
	if err := util.ReadInt64(r, &bi.Total); err != nil {
		return err
	}
	if err := util.ReadAddress(r, &bi.Bidder); err != nil {
		return err
	}
	return nil
}
