// fairroulette is a PoC smart contract for IOTA Smart Contracts, the Wasp node
// In this package smart contract is implemented as a hardcoded program. However, the program
// is wrapped into the VM wrapper interfaces and uses exactly the same sandbox interface
// as if it were Wasm VM.
// The smart contract implements simple gambling dapp.
// Players can place bets by sending requests to the smart contract. Each request is a value transaction.
// Committee is taking some minimum number of iotas as a reward for processing the transaction
// (configurable, usually several thousands).
// The rest of the iotas transferred to the smart contracts are taken as
// a bet placed on particular color on the roulette wheel.
//
// 2 minutes after first bet the smart contract automatically plays roulette wheel using
// unpredictable entropy provided by the BLS threshold signatures. Therefore FairRoulette is provably fair
// because even committee members can't predict the winning color.
//
// Then smart contract automatically distributes total staked amount to those players which placed their
// bets on the winning color proportionally to the amount staked.
// If nobody places on the winning color the total staked amount remains in the smart contracts account
package fairroulette

import (
	"bytes"
	"fmt"
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
	// request to place the bet. Public
	RequestPlaceBet = sctransaction.RequestCode(uint16(1))
	// request to lock the bets. Rejected if sent not from the smart contract itself
	RequestLockBets = sctransaction.RequestCode(uint16(2))
	// request to play and distribute. Rejected if sent not from the smart contract itself
	RequestPlayAndDistribute = sctransaction.RequestCode(uint16(3))
	// request to set the play period. By default it is 2 minutes
	RequestSetPlayPeriod = sctransaction.RequestCode(uint16(4) | sctransaction.RequestCodeProtected)
)

var entryPoints = fairRouletteProcessor{
	RequestPlaceBet:          placeBet,
	RequestLockBets:          lockBets,
	RequestPlayAndDistribute: playAndDistribute,
	RequestSetPlayPeriod:     setPlayPeriod,
}

const (
	ProgramHash = "FNT6snmmEM28duSg7cQomafbJ5fs596wtuNRn18wfaAz"

	// request argument to specify color of the bet. It always is taken modulo 5, so there are 5 possible colors
	ReqVarColor = "color"
	// state array to store all current bets
	StateVarBets = "bets"
	// state array to store locked bets
	StateVarLockedBets = "lockedBest"
	// state variable to store last winning color. Just for information
	StateVarLastWinningColor = "lastWinningColor"
	// 32 bytes of entropy taken from the hash of the transaction which locked current bets
	StateVarEntropyFromLocking = "entropyFromLocking"
	// set play period in seconds
	VarPlayPeriodSec = "playPeriod"

	// number of colors
	NumColors = 5
	// automatically lock and play 2 min after first current bet is confirmed
	DefaultPlaySecondsAfterFirstBet = 120
)

type betInfo struct {
	player address.Address
	reqId  sctransaction.RequestId
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

// WithGasLimit: not implemented, has no effect
func (f fairRouletteEntryPoint) WithGasLimit(i int) vmtypes.EntryPoint {
	return f
}

func (f fairRouletteEntryPoint) Run(ctx vmtypes.Sandbox) {
	f(ctx)
}

// the request places bet into the smart contract
func placeBet(ctx vmtypes.Sandbox) {
	ctx.Publish("placeBet")

	//ctx.GetWaspLog().Infof("$$$$$$$$$$ dump:\n%s\n", ctx.DumpAccount())

	state := ctx.AccessState()

	// if there are some bets locked, save the entropy derived immediately from it.
	// it is not predictable at the moment of locking and this saving makes it not playable later
	// entropy saved this way is essentially derived (hashed) from the locking transaction hash
	if state.GetArray(StateVarLockedBets).Len() > 0 {
		_, ok, err := state.GetHashValue(StateVarEntropyFromLocking)
		if !ok || err != nil {
			ehv := ctx.GetEntropy()
			state.SetHashValue(StateVarEntropyFromLocking, &ehv)
		}
	}

	// take input addresses of the request transaction. Must be exactly 1 otherwise.
	// Theoretically the transaction may have several addresses in inputs, then it is ignored
	senders := ctx.AccessRequest().Senders()
	if len(senders) != 1 {
		return
	}
	sender := senders[0]
	// look if there're some iotas left for the bet.
	// it is after minimum rewards are already taken. Here we accessing only the part of the smart contract
	// UTXOs: the ones which are coming with the current request
	sum := ctx.AccessOwnAccount().AvailableBalanceFromRequest(&balance.ColorIOTA)
	if sum == 0 {
		// nothing to bet
		ctx.Publish("placeBet: sum == 0: nothing to bet")
		return
	}
	// check if there's a Color variable among args
	// if not, ignore the request
	col, ok, _ := ctx.AccessRequest().Args().GetInt64(ReqVarColor)
	if !ok {
		ctx.Publish("wrong request, no Color specified")
		return
	}
	firstBet := state.GetArray(StateVarBets).Len() == 0

	// save the bet info in the array
	reqid := ctx.AccessRequest().ID()
	state.GetArray(StateVarBets).Push(encodeBetInfo(&betInfo{
		player: sender,
		sum:    sum,
		reqId:  reqid,
		color:  byte(col % NumColors),
	}))
	ctx.Publish(fmt.Sprintf("Place bet. player: %s sum: %d color: %d req: %s",
		sender.String(), sum, col, reqid.Short()))

	// if it is the first bet in the array, send time locked 'LockBets' request to itself.
	// it will be time-locked by default for the next 2 minutes, the it will be processed by smart contract
	if firstBet {
		period, ok, err := state.GetInt64(VarPlayPeriodSec)
		if err != nil || !ok || period < 10 {
			period = DefaultPlaySecondsAfterFirstBet
		}
		ctx.Publish(fmt.Sprintf("SendRequestToSelfWithDelay period = %d", period))
		ctx.SendRequestToSelfWithDelay(RequestLockBets, nil, uint32(period))
	}
}

func setPlayPeriod(ctx vmtypes.Sandbox) {
	ctx.Publish("setPlayPeriod")

	period, ok, err := ctx.AccessRequest().Args().GetInt64(VarPlayPeriodSec)
	if err != nil || !ok || period < 10 {
		// incorrect request arguments
		// minimum is 10 seconds
		return
	}
	ctx.AccessState().SetInt64(VarPlayPeriodSec, period)

	ctx.Publish(fmt.Sprintf("setPlayPeriod = %d", period))
}

// lockBet moves all current bets into the LockedBets array and erases current bets array
// it only processed if sent from the smart contract to itself
func lockBets(ctx vmtypes.Sandbox) {
	ctx.Publish("lockBets")

	if !ctx.AccessRequest().IsAuthorisedByAddress(ctx.GetOwnAddress()) {
		// ignore if request is not from itself
		return
	}
	state := ctx.AccessState()
	// append all current bets to the locked bets array
	lockedBets := state.GetArray(StateVarLockedBets)
	lockedBets.Append(state.GetArray(StateVarBets))
	state.GetArray(StateVarBets).Erase()

	numLockedBets := lockedBets.Len()
	ctx.Publish(fmt.Sprintf("lockBets: num = %d", numLockedBets))

	// clear entropy to be picked in the next request
	state.Del(StateVarEntropyFromLocking)

	// send request to self for playing the wheel with the entropy whicl will be known
	// after signing this state update transaction therefore unpredictable
	ctx.SendRequestToSelf(RequestPlayAndDistribute, nil)
}

// playAndDistribute takes the entropy, plays the game and distributes rewards to winners
func playAndDistribute(ctx vmtypes.Sandbox) {
	ctx.Publish("playAndDistribute")

	if !ctx.AccessRequest().IsAuthorisedByAddress(ctx.GetOwnAddress()) {
		// ignore if request is not from itself
		return
	}
	state := ctx.AccessState()

	lockedBetsArray := state.GetArray(StateVarLockedBets)
	numLockedBets := lockedBetsArray.Len()
	if numLockedBets == 0 {
		// nothing to play. Should not happen
		return
	}

	// take the entropy from the signing of the locked bets
	// it was saved by some 'place bet' request or otherwise it is taken from
	// the current context
	entropy, ok, err := state.GetHashValue(StateVarEntropyFromLocking)
	if !ok || err != nil {
		h := ctx.GetEntropy()
		entropy = &h
	}

	// 'playing the wheel' means taking first 8 bytes of the entropy as uint64 number and
	// calculating it modulo 5.
	winningColor := byte(util.Uint64From8Bytes(entropy[:8]) / NumColors)
	ctx.AccessState().SetInt64(StateVarLastWinningColor, int64(winningColor))

	// take locked bets from the array
	lockedBets := make([]*betInfo, numLockedBets)
	for i := range lockedBets {
		biData, ok := lockedBetsArray.At(uint16(i))
		if !ok {
			// inconsistency. Very sad
			return
		}
		bi, err := decodeBetInfo(biData)
		if err != nil {
			// inconsistency. Even more sad
			return
		}
		lockedBets = append(lockedBets, bi)
	}

	// calculate total placed amount
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

	// locked bets neither entropy are not needed anymore
	lockedBetsArray.Erase()
	state.Del(StateVarEntropyFromLocking)

	if len(winningBets) == 0 {
		// nobody played on winning Color -> all sums stay in the smart contract
		// move tokens to itself.
		// It is not necessary because all tokens are in the own account anyway.
		// However, it is healthy to compress number of outputs in the address
		if !ctx.AccessOwnAccount().MoveTokens(ctx.GetOwnAddress(), &balance.ColorIOTA, totalLockedAmount) {
			// inconsistency. A disaster
			ctx.Rollback()
			return
		}
	}

	// distribute total staked amount to players
	if !distributeLockedAmount(ctx, winningBets, totalLockedAmount) {
		ctx.Rollback()
		return
	}
}

// distributeLockedAmount distributes total locked amount proportionally to placed sums
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
	// NOTE 2: beware overflows

	for player, sum := range sumsByPlayers {
		sumsByPlayers[player] = (totalLockedAmount * sum) / totalWinningAmount
	}

	// make deterministic sequence by sorting. Eliminate possible rounding effects
	seqPlayers := make([]address.Address, 0, len(sumsByPlayers))
	resultSum := int64(0)
	for player, sum := range sumsByPlayers {
		seqPlayers = append(seqPlayers, player)
		resultSum += sum
	}
	sort.Slice(seqPlayers, func(i, j int) bool {
		return bytes.Compare(seqPlayers[i][:], seqPlayers[j][:]) < 0
	})

	// ensure we distribute not more than totalLockedAmount iotas
	if resultSum > totalLockedAmount {
		sumsByPlayers[seqPlayers[0]] -= resultSum - totalLockedAmount
	}

	// filter out those who proportionally got 0
	finalWinners := seqPlayers[:0]
	for _, player := range seqPlayers {
		if sumsByPlayers[player] <= 0 {
			continue
		}
		finalWinners = append(finalWinners, player)
	}
	// distribute iotas
	for i := range finalWinners {
		if !ctx.AccessOwnAccount().MoveTokens(&finalWinners[i], &balance.ColorIOTA, sumsByPlayers[finalWinners[i]]) {
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
