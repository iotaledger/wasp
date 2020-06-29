package fairroulette

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"io"
)

type fairRouletteProcessor map[sctransaction.RequestCode]fairRouletteEntryPoint

type fairRouletteEntryPoint func(ctx vmtypes.Sandbox)

const (
	RequestPlaceBet          = sctransaction.RequestCode(uint16(1))
	RequestVoteForPlay       = sctransaction.RequestCode(uint16(2))
	RequestPlayAndDistribute = sctransaction.RequestCode(uint16(3))
)

var entryPoints = fairRouletteProcessor{
	RequestPlaceBet:          placeBet,
	RequestVoteForPlay:       vote,
	RequestPlayAndDistribute: playAndDistribute,
}

const (
	ProgramHash = "3wo28GRrJu37v6D4xkjZsRLiVQrk3iMn7PifpMFoJEiM"

	ReqVarColor        = "color"
	StateVarNumBets    = "numbets"
	StateVarBets       = "bets"
	StateVarLockedBets = "lockedBest"
	StateVarNumVotes   = "numvotes"

	NumColors       = 8
	NumVotesForPlay = 10
)

type betInfo struct {
	player address.Address
	sum    int64
	color  byte
}

func GetProcessor() vmtypes.Processor {
	return entryPoints
}

func (f fairRouletteProcessor) GetEntryPoint(code sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	ep, ok := entryPoints[code]
	return ep, ok
}

func (f fairRouletteEntryPoint) WithGasLimit(i int) vmtypes.EntryPoint {
	return f
}

func (f fairRouletteEntryPoint) Run(ctx vmtypes.Sandbox) {
	f(ctx)
}

// the request places bet into the smart contract
func placeBet(ctx vmtypes.Sandbox) {
	// take sender. Must exactly 1
	senders := ctx.AccessRequest().Senders()
	if len(senders) != 1 {
		return
	}
	sender := senders[0]
	// look if there're some iotas left for the bet.
	// it is after min rewards.
	sum := ctx.AccessAccount().AvailableBalance(&balance.ColorIOTA)
	if sum == 0 {
		// nothing to bet
		return
	}
	// see if there's a color
	col, ok := ctx.AccessRequest().GetInt64(ReqVarColor)
	if !ok {
		// wrong request, no color specified
		return
	}
	// marshal bet data to binary.
	betData, err := util.Bytes(&betInfo{
		player: sender,
		sum:    sum,
		color:  byte(col % NumColors),
	})
	if err != nil {
		return
	}
	// push bet data into the state
	// the following shall be replaced with one call
	// ctx.Push(StateVarNumBets, betData)
	// we are not limited to the number of bets
	numBets, _, _ := ctx.AccessState().GetInt64(StateVarNumBets)
	key := fmt.Sprintf("%s:%d", StateVarBets, numBets)
	ctx.AccessState().Set(key, betData)
	ctx.AccessState().SetInt64(StateVarNumBets, numBets+1)
}

// anyone can vote, they can't predict the outcome anyway
// alternatively, only betters could be allowed to bet --> need for hashmap structure
func vote(ctx vmtypes.Sandbox) {
	numVotes, _, _ := ctx.AccessState().GetInt64(StateVarNumVotes)
	if numVotes+1 < NumVotesForPlay {
		ctx.AccessState().SetInt64(StateVarNumVotes, numVotes+1)
		return
	}
	ctx.SendRequestToSelf(RequestPlayAndDistribute, nil)
	ctx.AccessState().SetInt64(StateVarNumVotes, 0)
	// move all bets from StateVarBets to StateVarLockedBets
	// clear bets from StateVarBets
}

func playAndDistribute(ctx vmtypes.Sandbox) {
	if !ctx.AccessRequest().IsAuthorisedByAddress(ctx.GetAddress()) {
		// ignore if request is not from itself
		return
	}
	// take StateVarLockedBets
	// take Entropy
	// sum up bets on each color
	// run entropy on betters proportionally betted sums on each color
	// select winning color
	// distribute ALL betted iotas to those who betted on winning color proportionally to the
	// betted sums.
	// distribute sums
	// reset locked bets
}

func (bi *betInfo) Write(w io.Writer) error {
	_, _ = w.Write(bi.player[:])
	_ = util.WriteInt64(w, bi.sum)
	_ = util.WriteByte(w, bi.color)
	return nil
}

func (bi *betInfo) Read(r io.Reader) error {
	var err error
	if err = util.ReadAddress(r, &bi.player); err != nil {
		return err
	}
	if err = util.ReadInt64(r, &bi.sum); err != nil {
		return err
	}
	if bi.color, err = util.ReadByte(r); err != nil {
		return err
	}
	return nil
}
