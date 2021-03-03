// +build ignore

package fairauction

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

// serde of FairAction binary data

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
	if err := util.WriteInt64(w, ai.WhenStarted); err != nil {
		return err
	}
	if err := util.WriteInt64(w, ai.DurationMinutes); err != nil {
		return err
	}
	if _, err := w.Write(ai.AuctionOwner[:]); err != nil {
		return err
	}
	if err := util.WriteInt64(w, ai.TotalDeposit); err != nil {
		return err
	}
	if err := util.WriteInt64(w, ai.OwnerMargin); err != nil {
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
	if err = util.ReadInt64(r, &ai.WhenStarted); err != nil {
		return err
	}
	if err = util.ReadInt64(r, &ai.DurationMinutes); err != nil {
		return err
	}
	if err = coretypes.ReadAgentID(r, &ai.AuctionOwner); err != nil {
		return err
	}
	if err = util.ReadInt64(r, &ai.TotalDeposit); err != nil {
		return err
	}
	if err = util.ReadInt64(r, &ai.OwnerMargin); err != nil {
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
	if err := util.WriteInt64(w, bi.When); err != nil {
		return err
	}
	return nil
}

func (bi *BidInfo) Read(r io.Reader) error {
	if err := util.ReadInt64(r, &bi.Total); err != nil {
		return err
	}
	if err := coretypes.ReadAgentID(r, &bi.Bidder); err != nil {
		return err
	}
	if err := util.ReadInt64(r, &bi.When); err != nil {
		return err
	}
	return nil
}
