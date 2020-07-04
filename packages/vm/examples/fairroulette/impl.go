package fairroulette

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"io"
	"sort"
)

type fairRouletteProcessor map[sctransaction.RequestCode]fairRouletteEntryPoint

type fairRouletteEntryPoint func(ctx vmtypes.Sandbox)

const (
	RequestPlaceBet          = sctransaction.RequestCode(uint16(1))
	RequestLockBets          = sctransaction.RequestCode(uint16(2))
	RequestPlayAndDistribute = sctransaction.RequestCode(uint16(3))
)

var entryPoints = fairRouletteProcessor{
	RequestPlaceBet:          placeBet,
	RequestLockBets:          lockBets,
	RequestPlayAndDistribute: playAndDistribute,
}

const (
	ProgramHash = "3wo28GRrJu37v6D4xkjZsRLiVQrk3iMn7PifpMFoJEiM"

	ReqVarColor                = "Color"
	StateVarBets               = "bets"
	StateVarLockedBets         = "lockedBest"
	StateVarLastWinningColor   = "lastWinningColor"
	StateVarEntropyFromLocking = "entropyFromLocking"

	NumColors   = 8
	PlayEverSec = 120
)

type betInfo struct {
	player address.Address
	reqId  sctransaction.RequestId
	sum    int64
	color  byte
}

// all strings base58
type betInfoJson struct {
	PlayerAddr string `json:"player_addr"`
	ReqId      string `json:"req_id"`
	Sum        int64  `json:"sum"`
	Color      byte   `json:"color"`
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
	state := ctx.AccessState()

	if state.GetArray(StateVarLockedBets).Len() > 0 {
		// if there are some bets locked, save the entropy derived immediately from it.
		// it is not predictable at the moment of locking and saving makes it no playable later
		_, ok, err := state.GetHashValue(StateVarEntropyFromLocking)
		if !ok || err != nil {
			ehv := ctx.GetEntropy()
			state.SetHashValue(StateVarEntropyFromLocking, &ehv)
		}
	}
	// take senders. Must be exactly 1
	senders := ctx.AccessRequest().Senders()
	if len(senders) != 1 {
		return
	}
	sender := senders[0]
	// look if there're some iotas left for the bet.
	// it is after min rewards. Here we accessing only part which is coming with the current request
	sum := ctx.AccessOwnAccount().AvailableBalanceFromRequest(&balance.ColorIOTA)
	if sum == 0 {
		// nothing to bet
		return
	}
	// see if there's a Color among args
	col, ok, _ := ctx.AccessRequest().Args().GetInt64(ReqVarColor)
	if !ok {
		ctx.GetWaspLog().Errorf("wrong request, no Color specified")
		return
	}
	noBets := ctx.AccessState().GetArray(StateVarBets).Len() == 0
	ctx.AccessState().GetArray(StateVarBets).Push(encodeBetInfo(&betInfo{
		player: sender,
		sum:    sum,
		reqId:  ctx.AccessRequest().ID(),
		color:  byte(col % NumColors),
	}))
	if noBets {
		ctx.SendRequestToSelfWithDelay(RequestLockBets, nil, PlayEverSec)
	}
}

// anyone can lockBets, they can't predict the outcome anyway
// alternatively, only betters could be allowed to bet --> need for hashmap structure
func lockBets(ctx vmtypes.Sandbox) {
	if !ctx.AccessRequest().IsAuthorisedByAddress(ctx.GetOwnAddress()) {
		// ignore if request is not from itself
		return
	}
	ctx.AccessState().GetArray(StateVarLockedBets).Append(ctx.AccessState().GetArray(StateVarBets))
	ctx.AccessState().GetArray(StateVarBets).Erase()

	// clear entropy to be picked in the next request
	ctx.AccessState().Del(StateVarEntropyFromLocking)

	ctx.SendRequestToSelf(RequestPlayAndDistribute, nil)
}

func playAndDistribute(ctx vmtypes.Sandbox) {
	state := ctx.AccessState()

	if !ctx.AccessRequest().IsAuthorisedByAddress(ctx.GetOwnAddress()) {
		// ignore if request is not from itself
		return
	}
	lockedBetsArray := state.GetArray(StateVarLockedBets)
	numLockedBets := lockedBetsArray.Len()
	if numLockedBets == 0 {
		// nothing is to play
		return
	}

	// take the entropy from the signing of the locked bets
	entropy, ok, err := state.GetHashValue(StateVarEntropyFromLocking)
	if !ok || err != nil {
		h := ctx.GetEntropy()
		entropy = &h
	}

	winningColor := byte(util.Uint64From8Bytes(entropy[:8]) / NumColors)
	ctx.AccessState().SetInt64(StateVarLastWinningColor, int64(winningColor))

	// take locked bets
	lockedBets := make([]*betInfo, numLockedBets)
	for i := range lockedBets {
		biData, ok := lockedBetsArray.At(uint16(i))
		if !ok {
			// inconsistency
			return
		}
		bi, err := decodeBetInfo(biData)
		if err != nil {
			// inconsistency
			return
		}
		lockedBets = append(lockedBets, bi)
	}

	totalLockedAmount := int64(0)
	for _, bet := range lockedBets {
		totalLockedAmount += bet.sum
	}
	// select bets on winning Color
	winningBets := lockedBets[:0] // same underlying array
	for _, bet := range lockedBets {
		if bet.color == winningColor {
			winningBets = append(winningBets, bet)
		}
	}

	// locked bets are not needed anymore
	lockedBetsArray.Erase()
	state.Del(StateVarEntropyFromLocking)

	if len(winningBets) == 0 {
		// nobody played on winning Color -> all sums stay in the smart contract
		// move tokens to itself in order to compress number of outputs in the address
		if !ctx.AccessOwnAccount().MoveTokens(ctx.GetOwnAddress(), &balance.ColorIOTA, totalLockedAmount) {
			ctx.Rollback()
			return
		}
	}

	if !distributeLockedAmount(ctx, winningBets, totalLockedAmount) {
		ctx.Rollback()
		return
	}
}

func distributeLockedAmount(ctx vmtypes.Sandbox, bets []*betInfo, totalLockedAmount int64) bool {
	sumsByPlayers := make(map[address.Address]int64)
	totalWinningAmount := int64(0)
	for _, bet := range bets {
		if _, ok := sumsByPlayers[bet.player]; !ok {
			sumsByPlayers[bet.player] = 0
		}
		sumsByPlayers[bet.player] += bet.sum
		totalWinningAmount += bet.sum
	}

	// NOTE 1: float64 was avoided for determinism reasons
	// NOTE: beware overflows
	for player, sum := range sumsByPlayers {
		sumsByPlayers[player] = (totalLockedAmount * sum) / totalWinningAmount
	}
	// make deterministic sequence
	seqPlayers := make([]address.Address, 0, len(sumsByPlayers))
	resultSum := int64(0)
	for player, sum := range sumsByPlayers {
		seqPlayers = append(seqPlayers, player)
		resultSum += sum
	}
	sort.Slice(seqPlayers, func(i, j int) bool {
		return bytes.Compare(seqPlayers[i][:], seqPlayers[j][:]) < 0
	})

	if resultSum > totalLockedAmount {
		sumsByPlayers[seqPlayers[0]] -= resultSum - totalLockedAmount
	}
	finalWinners := seqPlayers[:0]
	for _, player := range seqPlayers {
		if sumsByPlayers[player] <= 0 {
			continue
		}
		finalWinners = append(finalWinners, player)
	}
	for _, player := range finalWinners {
		if !ctx.AccessOwnAccount().MoveTokens(&player, &balance.ColorIOTA, sumsByPlayers[player]) {
			return false
		}
	}
	return true
}

func encodeBetInfo(bi *betInfo) []byte {
	ret, _ := util.Bytes(bi)
	return ret
}

func decodeBetInfo(data []byte) (*betInfo, error) {
	var ret betInfo
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return &ret, nil
}

func (bi *betInfo) Write(w io.Writer) error {
	if _, err := w.Write(bi.player[:]); err != nil {
		return err
	}
	if err := bi.reqId.Write(w); err != nil {
		return err
	}
	if err := util.WriteInt64(w, bi.sum); err != nil {
		return err
	}
	if err := util.WriteByte(w, bi.color); err != nil {
		return err
	}
	return nil
}

func (bi *betInfo) Read(r io.Reader) error {
	var err error
	if err = util.ReadAddress(r, &bi.player); err != nil {
		return err
	}
	if err = bi.reqId.Read(r); err != nil {
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
