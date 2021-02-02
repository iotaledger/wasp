// +build ignore

// FairRoulette is a PoC smart contract for IOTA Smart Contracts and the Wasp node
// In this package smart contract is implemented as a hardcoded Go program.
// The program is wrapped into the VM wrapper interfaces and uses exactly the same sandbox interface
// as if it were Wasm VM.
// The smart contract implements simple gambling dapp.
// Players can place bets by sending requests to the smart contract. Each request is a value transaction.
// Smart contract is taking some minimum number of iotas as a reward for processing the transaction
// (configurable, usually several thousands).
// The rest of the iotas are taken as a bet placed on particular color on the roulette wheel.
//
// Approx 2 minutes after first bet, the smart contract automatically plays roulette wheel using
// unpredictable entropy provided by the BLS threshold signatures. This way FairRoulette is provably fair
// because even committee members can't predict the winning color.
//
// Then smart contract automatically distributes total betted amount to those players which placed their
// bets on the winning color proportionally to the amount.
// If nobody places on the winning color the total staked amount remains in the smart contracts account
package fairroulette

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
)

// implement Processor and EntryPoint interfaces
type fairRouletteProcessor map[coretypes.Hname]fairRouletteEntryPoint

type fairRouletteEntryPoint func(ctx coretypes.Sandbox) error

// ID of the smart contract program
const ProgramHash = "FNT6snmmEM28duSg7cQomafbJ5fs596wtuNRn18wfaAz"

// constants for request codes
var (
	// request to place the bet
	RequestPlaceBet = coretypes.Hn("placeBet")
	// request to lock the bets
	RequestLockBets = coretypes.Hn("lockBets")
	// request to play and distribute
	RequestPlayAndDistribute = coretypes.Hn("playAndDistribute")
	// request to set the play period. By default it is 2 minutes.
	// It only will be processed is sent by the owner of the smart contract
	RequestSetPlayPeriod = coretypes.Hn("setPlayPeriod")
)

// the processor is a map of entry points
var entryPoints = fairRouletteProcessor{
	RequestPlaceBet:          placeBet,
	RequestLockBets:          lockBets,
	RequestPlayAndDistribute: playAndDistribute,
	RequestSetPlayPeriod:     setPlayPeriod,
}

// string constants for names of state and request argument variables
const (
	/// General constants
	// number of colors
	NumColors = 5
	// automatically lock and play 2 min after first current bet is confirmed
	DefaultPlaySecondsAfterFirstBet = 120

	/// State variables
	// state array to store all current bets
	StateVarBets = "bets"
	// state array to store locked bets
	StateVarLockedBets = "lockedBets"
	// state variable to store last winning color. Just for information
	StateVarLastWinningColor = "lastWinningColor"
	// 32 bytes of entropy taken from the hash of the transaction which locked current bets
	StateVarEntropyFromLocking = "entropyFromLocking"
	// estimated timestamp for next play (nanoseconds)
	StateVarNextPlayTimestamp = "nextPlayTimestamp"
	// array color => amount of wins so far
	StateArrayWinsPerColor = "winsPerColor"
	// dictionary address => PlayerStats
	StateVarPlayerStats = "playerStats"

	/// Request variables (arguments)
	// request argument to specify color of the bet. It always is taken modulo NumColors,
	// so there are NumColors possible colors
	ReqVarColor = "color"
	// specify play period in seconds
	ReqVarPlayPeriodSec = "playPeriod"
)

// BetInfo contains data of the bet
type BetInfo struct {
	Player coretypes.AgentID
	reqId  coretypes.RequestID
	Sum    int64
	Color  byte
}

// Smart contract keep historical stats for players. For fun
type PlayerStats struct {
	Bets uint32
	Wins uint32
}

// coonnection of the SC program with the Wasp node
func GetProcessor() coretypes.Processor {
	return entryPoints
}

func (f fairRouletteProcessor) GetEntryPoint(code coretypes.Hname) (coretypes.EntryPoint, bool) {
	ep, ok := entryPoints[code]
	return ep, ok
}

func (v fairRouletteProcessor) GetDescription() string {
	return "FairRoulette hard coded smart contract processor"
}

func (f fairRouletteEntryPoint) Call(ctx coretypes.Sandbox) (dict.Dict, error) {
	err := f(ctx)
	if err != nil {
		ctx.Event(fmt.Sprintf("error %v", err))
	}
	return nil, err
}

// TODO
func (ep fairRouletteEntryPoint) IsView() bool {
	return false
}

// TODO
func (ep fairRouletteEntryPoint) CallView(ctx coretypes.SandboxView) (dict.Dict, error) {
	panic("implement me")
}

// the request places bet into the smart contract
func placeBet(ctx coretypes.Sandbox) error {
	ctx.Event("placeBet")
	params := ctx.Params()

	state := ctx.State()

	// if there are some bets locked, save the entropy derived immediately from it.
	// it is not predictable at the moment of locking and this saving makes it not playable later
	// entropy saved this way is derived (hashed) from the locking transaction hash
	// we do this trick to be able to deterministically check if smart contract is really fair.
	// The played color is a deterministic function of the hash of transaction which locked the bets
	if collections.NewArray(state, StateVarLockedBets).MustLen() > 0 {
		ok := state.MustHas(StateVarEntropyFromLocking)
		if !ok {
			ehv := ctx.GetEntropy()
			state.Set(StateVarEntropyFromLocking, codec.EncodeHashValue(&ehv))
		}
	}

	// take input addresses of the request transaction. Must be exactly 1 otherwise.
	// Theoretically the transaction may have several addresses in inputs, then it is ignored
	sender := ctx.Caller()

	// look if there're some iotas left for the bet after minimum rewards are already taken.
	// Here we are accessing only the part of the UTXOs which the ones which are coming with the current request
	sum := ctx.IncomingTransfer().Balance(balance.ColorIOTA)
	if sum == 0 {
		// nothing to bet
		return fmt.Errorf("placeBet: sum == 0: nothing to bet")
	}
	// check if there's a Color variable among args. If not, ignore the request
	col, ok, _ := codec.DecodeInt64(params.MustGet(ReqVarColor))
	if !ok {
		return fmt.Errorf("wrong request, no Color specified")
	}
	firstBet := collections.NewArray(state, StateVarBets).MustLen() == 0

	reqid := ctx.RequestID()
	betInfo := &BetInfo{
		Player: sender,
		Sum:    sum,
		reqId:  reqid,
		Color:  byte(col % NumColors),
	}

	// save the bet info in the array
	collections.NewArray(state, StateVarBets).MustPush(encodeBetInfo(betInfo))

	ctx.Event(fmt.Sprintf("Place bet: player: %s sum: %d color: %d req: %s", sender.String(), sum, col, reqid.Short()))

	err := withPlayerStats(ctx, &betInfo.Player, func(ps *PlayerStats) {
		ps.Bets += 1
	})
	if err != nil {
		ctx.Log().Panicf("%v", err)
	}

	// if it is the first bet in the array, send time locked 'LockBets' request to itself.
	// it will be time-locked by default for the next 2 minutes, the it will be processed by smart contract
	if firstBet {
		period, ok, _ := codec.DecodeInt64(state.MustGet(ReqVarPlayPeriodSec))
		if !ok || period < 10 {
			period = DefaultPlaySecondsAfterFirstBet
		}

		nextPlayTimestamp := (time.Duration(ctx.GetTimestamp())*time.Nanosecond + time.Duration(period)*time.Second).Nanoseconds()
		state.Set(StateVarNextPlayTimestamp, codec.EncodeInt64(nextPlayTimestamp))

		ctx.Event(fmt.Sprintf("PostRequestToSelfWithDelay period = %d", period))

		// send the timelocked Lock request to self. TimeLock is for number of seconds taken from the state variable
		// By default it is 2 minutes, i.e. Lock request will be processed after 2 minutes.
		if ctx.PostRequest(coretypes.PostRequestParams{
			TargetContractID: ctx.ContractID(),
			EntryPoint:       RequestLockBets,
			TimeLock:         uint32(period),
		}) {
			ctx.Event(fmt.Sprintf("play deadline is set after %d seconds", period))
		} else {
			ctx.Event(fmt.Sprintf("failed to set play deadline"))
		}
	}
	return nil
}

// admin (protected) request to set the period of autoplay. It only can be processed by the owner of the smart contract
func setPlayPeriod(ctx coretypes.Sandbox) error {
	ctx.Event("setPlayPeriod")
	params := ctx.Params()

	// TODO refactor to new account system
	//if ctx.Caller() != *ctx.OriginatorAddress() {
	//	// not authorized
	//	return fmt.Errorf("setPlayPeriod: not authorized")
	//}

	period, ok, err := codec.DecodeInt64(params.MustGet(ReqVarPlayPeriodSec))
	if err != nil || !ok || period < 10 {
		// incorrect request arguments
		// minimum is 10 seconds
		return fmt.Errorf("wrong parameter '%s'", ReqVarPlayPeriodSec)
	}
	ctx.State().Set(ReqVarPlayPeriodSec, codec.EncodeInt64(period))

	ctx.Event(fmt.Sprintf("setPlayPeriod = %d", period))
	return nil
}

// lockBet moves all current bets into the LockedBets array and erases current bets array
// it only processed if sent from the smart contract to itself
func lockBets(ctx coretypes.Sandbox) error {
	ctx.Event("lockBets")

	scAddr := coretypes.NewAgentIDFromContractID(ctx.ContractID())
	if ctx.Caller() != scAddr {
		// ignore if request is not from itself
		return fmt.Errorf("attempt of unauthorised access")
	}
	state := ctx.State()
	// append all current bets to the locked bets array
	lockedBets := collections.NewArray(state, StateVarLockedBets)
	lockedBets.MustExtend(collections.NewArrayReadOnly(state, StateVarBets))
	collections.NewArray(state, StateVarBets).MustErase()

	numLockedBets := lockedBets.MustLen()
	ctx.Event(fmt.Sprintf("lockBets: num = %d", numLockedBets))

	// clear entropy to be picked in the next request
	state.Del(StateVarEntropyFromLocking)

	// send request to self for playing the wheel with the entropy whicl will be known
	// after signing this state update transaction therefore unpredictable
	ctx.PostRequest(coretypes.PostRequestParams{
		TargetContractID: ctx.ContractID(),
		EntryPoint:       RequestPlayAndDistribute,
	})

	return nil
}

// playAndDistribute takes the entropy, plays the game and distributes rewards to winners
func playAndDistribute(ctx coretypes.Sandbox) error {
	ctx.Event("playAndDistribute")

	scAddr := coretypes.NewAgentIDFromContractID(ctx.ContractID())
	if ctx.Caller() != scAddr {
		// ignore if request is not from itself
		return fmt.Errorf("playAndDistribute from the wrong sender")
	}
	state := ctx.State()

	lockedBetsArray := collections.NewArray(state, StateVarLockedBets)
	numLockedBets := lockedBetsArray.MustLen()
	if numLockedBets == 0 {
		// nothing to play. Should not happen
		return fmt.Errorf("internal error. Nothing to play")
	}

	// take the entropy from the signing of the locked bets
	// it was saved by some 'place bet' request or otherwise it is taken from
	// the current context
	entropy, ok, _ := codec.DecodeHashValue(state.MustGet(StateVarEntropyFromLocking))
	if !ok {
		h := ctx.GetEntropy()
		entropy = &h
	}

	// 'playing the wheel' means taking first 8 bytes of the entropy as uint64 number and
	// calculating it modulo NumColors.
	winningColor := byte(util.MustUint64From8Bytes(entropy[:8]) % NumColors)
	ctx.State().Set(StateVarLastWinningColor, codec.EncodeInt64(int64(winningColor)))

	ctx.Event(fmt.Sprintf("$$$$$$$$$$ winning color is = %d", winningColor))

	addToWinsPerColor(ctx, winningColor)

	// take locked bets from the array
	totalLockedAmount := int64(0)
	lockedBets := make([]*BetInfo, numLockedBets)
	for i := range lockedBets {
		bi, err := DecodeBetInfo(lockedBetsArray.MustGetAt(uint16(i)))
		if err != nil {
			// inconsistency. Even more sad
			panic(err)
		}
		totalLockedAmount += bi.Sum
		lockedBets[i] = bi
	}

	ctx.Event(fmt.Sprintf("$$$$$$$$$$ totalLockedAmount = %d", totalLockedAmount))

	// select bets on winning Color
	winningBets := lockedBets[:0] // same underlying array
	for _, bet := range lockedBets {
		if bet.Color == winningColor {
			winningBets = append(winningBets, bet)
		}
	}

	ctx.Event(fmt.Sprintf("$$$$$$$$$$ winningBets: %d", len(winningBets)))

	// locked bets neither entropy are not needed anymore
	lockedBetsArray.MustErase()
	state.Del(StateVarEntropyFromLocking)

	if len(winningBets) == 0 {

		ctx.Event(fmt.Sprintf("$$$$$$$$$$ nobody wins: amount of %d stays in the smart contract", totalLockedAmount))

		// nobody played on winning Color -> all sums stay in the smart contract
		// move tokens to itself.
		// It is not necessary because all tokens are in the own account anyway.
		// However, it is healthy to compress number of outputs in the address

		//agent := coretypes.NewAgentIDFromContractID(ctx.ContractID())
		//if !ctx.MoveTokens(agent, balance.ColorIOTA, totalLockedAmount) {
		//	// inconsistency. A disaster
		//	ctx.Event(fmt.Sprintf("$$$$$$$$$$ something went wrong 1"))
		//	ctx.Log().Panicf("MoveTokens failed")
		//}
	}

	// distribute total staked amount to players
	if !distributeLockedAmount(ctx, winningBets, totalLockedAmount) {
		ctx.Event(fmt.Sprintf("$$$$$$$$$$ something went wrong 2"))
		ctx.Log().Panicf("distributeLockedAmount failed")
	}

	for _, betInfo := range winningBets {
		err := withPlayerStats(ctx, &betInfo.Player, func(ps *PlayerStats) {
			ps.Wins += 1
		})
		if err != nil {
			ctx.Log().Panicf("%v", err)
		}
	}
	return nil
}

func addToWinsPerColor(ctx coretypes.Sandbox, winningColor byte) {
	winsPerColorArray := collections.NewArray(ctx.State(), StateArrayWinsPerColor)

	// first time? Initialize counters
	if winsPerColorArray.MustLen() == 0 {
		for i := 0; i < NumColors; i++ {
			winsPerColorArray.MustPush(util.Uint32To4Bytes(0))
		}
	}

	winsb := winsPerColorArray.MustGetAt(uint16(winningColor))
	wins := util.MustUint32From4Bytes(winsb)
	winsPerColorArray.MustSetAt(uint16(winningColor), util.Uint32To4Bytes(wins+1))
}

// distributeLockedAmount distributes total locked amount proportionally to placed sums
func distributeLockedAmount(ctx coretypes.Sandbox, bets []*BetInfo, totalLockedAmount int64) bool {
	sumsByPlayers := make(map[coretypes.AgentID]int64)
	totalWinningAmount := int64(0)
	for _, bet := range bets {
		if _, ok := sumsByPlayers[bet.Player]; !ok {
			sumsByPlayers[bet.Player] = 0
		}
		sumsByPlayers[bet.Player] += bet.Sum
		totalWinningAmount += bet.Sum
	}

	// NOTE 1: float64 was avoided for determinism reasons
	// NOTE 2: beware overflows

	for player, sum := range sumsByPlayers {
		sumsByPlayers[player] = (totalLockedAmount * sum) / totalWinningAmount
	}

	// make deterministic sequence by sorting. Eliminate possible rounding effects
	seqPlayers := make([]coretypes.AgentID, 0, len(sumsByPlayers))
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

		available := ctx.Balance(balance.ColorIOTA)
		ctx.Event(fmt.Sprintf("sending reward iotas %d to the winner %s. Available iotas: %d",
			sumsByPlayers[finalWinners[i]], finalWinners[i].String(), available))

		//if !ctx.MoveTokens(finalWinners[i], balance.ColorIOTA, sumsByPlayers[finalWinners[i]]) {
		//	return false
		//}
	}
	return true
}

func encodeBetInfo(bi *BetInfo) []byte {
	ret, _ := util.Bytes(bi)
	return ret
}

func DecodeBetInfo(data []byte) (*BetInfo, error) {
	var ret BetInfo
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return &ret, nil
}

func (bi *BetInfo) Write(w io.Writer) error {
	if _, err := w.Write(bi.Player[:]); err != nil {
		return err
	}
	if err := bi.reqId.Write(w); err != nil {
		return err
	}
	if err := util.WriteInt64(w, bi.Sum); err != nil {
		return err
	}
	if err := util.WriteByte(w, bi.Color); err != nil {
		return err
	}
	return nil
}

func (bi *BetInfo) Read(r io.Reader) error {
	var err error
	if err = coretypes.ReadAgentID(r, &bi.Player); err != nil {
		return err
	}
	if err = bi.reqId.Read(r); err != nil {
		return err
	}
	if err = util.ReadInt64(r, &bi.Sum); err != nil {
		return err
	}
	if bi.Color, err = util.ReadByte(r); err != nil {
		return err
	}
	return nil
}

func (b *BetInfo) String() string {
	return fmt.Sprintf("[player %s bets %d IOTAs on color %d]", b.Player.String()[:6], b.Sum, b.Color)
}

func encodePlayerStats(ps *PlayerStats) []byte {
	ret, _ := util.Bytes(ps)
	return ret
}

func DecodePlayerStats(data []byte) (*PlayerStats, error) {
	var ret PlayerStats
	if data != nil {
		if err := ret.Read(bytes.NewReader(data)); err != nil {
			return nil, err
		}
	}
	return &ret, nil
}

func (ps *PlayerStats) Write(w io.Writer) error {
	if err := util.WriteUint32(w, ps.Bets); err != nil {
		return err
	}
	if err := util.WriteUint32(w, ps.Wins); err != nil {
		return err
	}
	return nil
}

func (ps *PlayerStats) Read(r io.Reader) error {
	var err error
	if err = util.ReadUint32(r, &ps.Bets); err != nil {
		return err
	}
	if err = util.ReadUint32(r, &ps.Wins); err != nil {
		return err
	}
	return nil
}

func (ps *PlayerStats) String() string {
	return fmt.Sprintf("[bets: %d - wins: %d]", ps.Bets, ps.Wins)
}

func withPlayerStats(ctx coretypes.Sandbox, player *coretypes.AgentID, f func(ps *PlayerStats)) error {
	statsDict := collections.NewMap(ctx.State(), StateVarPlayerStats)
	b := statsDict.MustGetAt(player.Bytes())
	stats, err := DecodePlayerStats(b)
	if err != nil {
		return err
	}

	f(stats)

	statsDict.MustSetAt(player.Bytes(), encodePlayerStats(stats))

	return nil
}
