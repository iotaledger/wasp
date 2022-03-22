// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmsolo

import (
	"flag"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	cryptolib "github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/stretchr/testify/require"
)

const ( // TODO set back to false
	SoloDebug        = true
	SoloHostTracing  = true
	SoloStackTracing = true
)

var (
	// GoWasm / RsWasm / TsWasm are used to specify the Wasm language mode,
	// By default, SoloContext will try to run the Go SC code directly (no Wasm)
	// The 3 flags can be used to cause Wasm code to be loaded and run instead.
	// They are checked in sequence and the first one set determines the Wasm language used.
	GoWasm = flag.Bool("gowasm", false, "use Go Wasm smart contract code")
	RsWasm = flag.Bool("rswasm", false, "use Rust Wasm smart contract code")
	TsWasm = flag.Bool("tswasm", false, "use TypeScript Wasm smart contract code")

	// UseWasmEdge flag is kept here in case we decide to use WasmEdge again. Some tests
	// refer to this flag, so we keep it here instead of having to comment out a bunch
	// of code. To actually enable WasmEdge you need to uncomment the relevant lines in
	// NewSoloContextForChain(), and remove the go:build directives from wasmedge.go, so
	// that the linker can actually pull in the WasmEdge runtime.
	UseWasmEdge = flag.Bool("wasmedge", false, "use WasmEdge instead of WasmTime")
)

const (
	L2FundsContract   = 1_000_000
	L2FundsCreator    = 2_000_000
	L2FundsOriginator = 3_000_000

	WasmDustDeposit = 1000
)

type SoloContext struct {
	Chain       *solo.Chain
	Convertor   wasmhost.WasmConvertor
	creator     *SoloAgent
	Dust        uint64
	Err         error
	Gas         uint64
	GasFee      uint64
	Hprog       hashing.HashValue
	isRequest   bool
	IsWasm      bool
	keyPair     *cryptolib.KeyPair
	mint        uint64
	offLedger   bool
	scName      string
	Tx          *ledgerstate.Transaction
	wasmHostOld wasmlib.ScHost
	wc          *wasmhost.WasmContext
}

var (
	//_ iscp.Gas                  = &SoloContext{}
	_        wasmlib.ScFuncCallContext = &SoloContext{}
	_        wasmlib.ScViewCallContext = &SoloContext{}
	SoloSeed                           = cryptolib.NewSeed([]byte("SoloSeedSoloSeedSoloSeedSoloSeed"))
)

func (ctx *SoloContext) Burn(i int64) {
	// ignore gas for now
}

func (ctx *SoloContext) Budget() int64 {
	// ignore gas for now
	return 0
}

func (ctx *SoloContext) SetBudget(i int64) {
	// ignore gas for now
}

//nolint:unused,deadcode
func contains(s []*iscp.AgentID, e *iscp.AgentID) bool {
	for _, a := range s {
		if a.Equals(e) {
			return true
		}
	}
	return false
}

// NewSoloContext can be used to create a SoloContext associated with a smart contract
// with minimal information and will verify successful creation before returning ctx.
// It will start a default chain "chain1" before initializing the smart contract.
// It takes the scName and onLoad() function associated with the contract.
// Optionally, an init.Func that has been initialized with the parameters to pass to
// the contract's init() function can be specified.
// Unless you want to use a different chain than the default "chain1" this will be your
// function of choice to set up a smart contract for your tests
func NewSoloContext(t *testing.T, scName string, onLoad wasmhost.ScOnloadFunc, init ...*wasmlib.ScInitFunc) *SoloContext {
	ctx := NewSoloContextForChain(t, nil, nil, scName, onLoad, init...)
	require.NoError(t, ctx.Err)
	return ctx
}

// NewSoloContextForChain can be used to create a SoloContext associated with a smart contract
// on a particular chain.  When chain is nil the function will start a default chain "chain1"
// before initializing the smart contract.
// When creator is nil the creator will be the chain originator
// It takes the scName and onLoad() function associated with the contract.
// Optionally, an init.Func that has been initialized with the parameters to pass to
// the contract's init() function can be specified.
// You can check for any error that occurred by checking the ctx.Err member.
func NewSoloContextForChain(t *testing.T, chain *solo.Chain, creator *SoloAgent, scName string,
	onLoad wasmhost.ScOnloadFunc, init ...*wasmlib.ScInitFunc,
) *SoloContext {
	ctx := soloContext(t, chain, scName, creator)

	var keyPair *cryptolib.KeyPair
	if creator != nil {
		keyPair = creator.Pair
	}
	ctx.uploadWasm(keyPair)
	if ctx.Err != nil {
		return ctx
	}

	var params []interface{}
	if len(init) != 0 {
		params = init[0].Params()
	}
	if !ctx.IsWasm {
		wasmhost.GoWasmVM = func() wasmhost.WasmVM {
			return wasmhost.NewWasmGoVM(ctx.scName, onLoad)
		}
	}
	//if ctx.IsWasm && *UseWasmEdge && wasmproc.GoWasmVM == nil {
	//	wasmproc.GoWasmVM = wasmhost.NewWasmEdgeVM
	//}
	ctx.Err = ctx.Chain.DeployContract(keyPair, ctx.scName, ctx.Hprog, params...)
	if !ctx.IsWasm {
		// just in case deploy failed we don't want to leave this around
		wasmhost.GoWasmVM = nil
	}
	if ctx.Err != nil {
		return ctx
	}

	return ctx.init(onLoad)
}

// NewSoloContextForNative can be used to create a SoloContext associated with a native smart contract
// on a particular chain. When chain is nil the function will start a default chain "chain1" before
// deploying and initializing the smart contract.
// When creator is nil the creator will be the chain originator
// It takes the scName, onLoad() function, and processor associated with the contract.
// Optionally, an init.Func that has been initialized with the parameters to pass to
// the contract's init() function can be specified.
// You can check for any error that occurred by checking the ctx.Err member.
func NewSoloContextForNative(t *testing.T, chain *solo.Chain, creator *SoloAgent, scName string, onLoad wasmhost.ScOnloadFunc,
	proc *coreutil.ContractProcessor, init ...*wasmlib.ScInitFunc,
) *SoloContext {
	ctx := soloContext(t, chain, scName, creator)
	ctx.Chain.Env.WithNativeContract(proc)
	ctx.Hprog = proc.Contract.ProgramHash

	var keyPair *cryptolib.KeyPair
	if creator != nil {
		keyPair = creator.Pair
	}
	var params []interface{}
	if len(init) != 0 {
		params = init[0].Params()
	}
	ctx.Err = ctx.Chain.DeployContract(keyPair, scName, ctx.Hprog, params...)
	if ctx.Err != nil {
		return ctx
	}

	return ctx.init(onLoad)
}

func soloContext(t *testing.T, chain *solo.Chain, scName string, creator *SoloAgent) *SoloContext {
	ctx := &SoloContext{scName: scName, Chain: chain, creator: creator, Dust: WasmDustDeposit}
	if chain == nil {
		ctx.Chain = StartChain(t, "chain1")
	}
	return ctx
}

// StartChain starts a new chain named chainName.
func StartChain(t *testing.T, chainName string, env ...*solo.Solo) *solo.Chain {
	if SoloDebug {
		// avoid pesky timeouts during debugging
		wasmhost.DisableWasmTimeout = true
	}
	wasmhost.HostTracing = SoloHostTracing

	var soloEnv *solo.Solo
	if len(env) != 0 {
		soloEnv = env[0]
	}
	if soloEnv == nil {
		soloEnv = solo.New(t, SoloDebug, SoloStackTracing, SoloSeed)
	}
	return soloEnv.NewChain(nil, chainName)
}

// Account returns a SoloAgent for the smart contract associated with ctx
func (ctx *SoloContext) Account() *SoloAgent {
	return &SoloAgent{
		Env:     ctx.Chain.Env,
		Pair:    nil,
		address: ctx.Chain.ChainID.AsAddress(),
		hname:   iscp.Hn(ctx.scName),
	}
}

func (ctx *SoloContext) AccountID() wasmtypes.ScAgentID {
	return ctx.Account().ScAgentID()
}

// AdvanceClockBy is used to forward the internal clock by the provided step duration.
func (ctx *SoloContext) AdvanceClockBy(step time.Duration) {
	ctx.Chain.Env.AdvanceClockBy(step)
}

// Balance returns the account balance of the specified agent on the chain associated with ctx.
// The optional color parameter can be used to retrieve the balance for the specific color.
// When color is omitted, wasmlib.IOTA is assumed.
func (ctx *SoloContext) Balance(agent *SoloAgent, color ...wasmtypes.ScColor) uint64 {
	account := iscp.NewAgentID(agent.address, agent.hname)
	balances := ctx.Chain.GetAccountBalance(account)
	switch len(color) {
	case 0:
		return balances.Get(colored.IOTA)
	case 1:
		col, err := colored.ColorFromBytes(color[0].Bytes())
		require.NoError(ctx.Chain.Env.T, err)
		return balances.Get(col)
	default:
		require.Fail(ctx.Chain.Env.T, "too many color arguments")
		return 0
	}
}

func (ctx *SoloContext) ChainID() wasmtypes.ScChainID {
	return ctx.Convertor.ScChainID(ctx.Chain.ChainID)
}

func (ctx *SoloContext) ChainOwnerID() wasmtypes.ScAgentID {
	return ctx.Convertor.ScAgentID(ctx.Chain.OriginatorAgentID)
}

func (ctx *SoloContext) ContractCreator() wasmtypes.ScAgentID {
	return ctx.Creator().ScAgentID()
}

// ContractExists checks to see if the contract named scName exists in the chain associated with ctx.
func (ctx *SoloContext) ContractExists(scName string) error {
	_, err := ctx.Chain.FindContract(scName)
	return err
}

// Creator returns a SoloAgent representing the contract creator
func (ctx *SoloContext) Creator() *SoloAgent {
	if ctx.creator != nil {
		return ctx.creator
	}
	return ctx.Originator()
}

func (ctx *SoloContext) EnqueueRequest() {
	ctx.isRequest = true
}

func (ctx *SoloContext) existFile(path, ext string) string {
	fileName := ctx.scName + ext

	// first check for new file in path
	pathName := path + fileName
	exists, _ := util.ExistsFilePath(pathName)
	if exists {
		return pathName
	}

	// check for file in current folder
	exists, _ = util.ExistsFilePath(fileName)
	if exists {
		return fileName
	}

	// file not found
	return ""
}

func (ctx *SoloContext) Host() wasmlib.ScHost {
	return nil
}

// init further initializes the SoloContext.
func (ctx *SoloContext) init(onLoad wasmhost.ScOnloadFunc) *SoloContext {
	ctx.wc = wasmhost.NewWasmContextForSoloContext("-solo-", NewSoloSandbox(ctx))
	ctx.wasmHostOld = wasmhost.Connect(ctx.wc)
	onLoad(-1)
	return ctx
}

// InitFuncCallContext is a function that is required to use SoloContext as an ScFuncCallContext
func (ctx *SoloContext) InitFuncCallContext() {
	_ = wasmhost.Connect(ctx.wc)
}

// InitViewCallContext is a function that is required to use SoloContext as an ScViewCallContext
func (ctx *SoloContext) InitViewCallContext(hContract wasmtypes.ScHname) wasmtypes.ScHname {
	_ = wasmhost.Connect(ctx.wc)
	return ctx.Convertor.ScHname(iscp.Hn(ctx.scName))
}

// Minted returns the color and amount of newly minted tokens
func (ctx *SoloContext) Minted() (wasmtypes.ScColor, uint64) {
	t := ctx.Chain.Env.T
	t.Logf("minting request tx: %s", ctx.Tx.ID().Base58())
	mintedAmounts := colored.BalancesFromL1Map(utxoutil.GetMintedAmounts(ctx.Tx))
	require.Len(t, mintedAmounts, 1)
	var mintedColor wasmtypes.ScColor
	var mintedAmount uint64
	for c := range mintedAmounts {
		mintedColor = ctx.Convertor.ScColor(&c)
		mintedAmount = mintedAmounts[c]
		break
	}
	t.Logf("Minted: amount = %d color = %s", mintedAmount, mintedColor.String())
	return mintedColor, mintedAmount
}

// NewSoloAgent creates a new SoloAgent with solo.Saldo tokens in its address
func (ctx *SoloContext) NewSoloAgent() *SoloAgent {
	return NewSoloAgent(ctx.Chain.Env)
}

// OffLedger tells SoloContext to Post() the next request off-ledger
func (ctx *SoloContext) OffLedger(agent *SoloAgent) wasmlib.ScFuncCallContext {
	ctx.offLedger = true
	ctx.keyPair = agent.Pair
	return ctx
}

// Originator returns a SoloAgent representing the chain originator
func (ctx *SoloContext) Originator() *SoloAgent {
	c := ctx.Chain
	return &SoloAgent{Env: c.Env, Pair: c.OriginatorKeyPair, address: c.OriginatorAddress}
}

// Sign is used to force a different agent for signing a Post() request
func (ctx *SoloContext) Sign(agent *SoloAgent, mint ...uint64) wasmlib.ScFuncCallContext {
	ctx.keyPair = agent.Pair
	if len(mint) != 0 {
		ctx.mint = mint[0]
	}
	return ctx
}

func (ctx *SoloContext) SoloContextForCore(t *testing.T, scName string, onLoad wasmhost.ScOnloadFunc) *SoloContext {
	ctxCore := soloContext(t, ctx.Chain, scName, nil).init(onLoad)
	ctxCore.wasmHostOld = ctx.wasmHostOld
	return ctxCore
}

// Transfer creates a new ScTransfers proxy
func (ctx *SoloContext) Transfer() wasmlib.ScTransfers {
	return wasmlib.NewScTransfers()
}

func (ctx *SoloContext) uploadWasm(keyPair *cryptolib.KeyPair) {
	wasmFile := ""
	if *GoWasm {
		// find Go Wasm file
		wasmFile = ctx.existFile("../go/pkg/", "_go.wasm")
	} else if *RsWasm {
		// find Rust Wasm file
		wasmFile = ctx.existFile("../pkg/", "_bg.wasm")
	} else if *TsWasm {
		// find TypeScript Wasm file
		wasmFile = ctx.existFile("../ts/pkg/", "_ts.wasm")
	} else {
		// none of the Wasm modes selected, use WasmGoVM to run Go SC code directly
		ctx.Hprog, ctx.Err = ctx.Chain.UploadWasm(keyPair, []byte("go:"+ctx.scName))
		return
	}

	if wasmFile == "" {
		panic("cannot find Wasm file for: " + ctx.scName)
	}

	// upload the Wasm code into the core blob contract
	ctx.Hprog, ctx.Err = ctx.Chain.UploadWasmFromFile(keyPair, wasmFile)
	ctx.IsWasm = true
}

// WaitForPendingRequests waits for expectedRequests pending requests to be processed.
// a negative value indicates the absolute amount of requests
// The function will wait for maxWait (default 5 seconds) duration before giving up with a timeout.
// The function returns the false in case of a timeout.
func (ctx *SoloContext) WaitForPendingRequests(expectedRequests int, maxWait ...time.Duration) bool {
	_ = wasmhost.Connect(ctx.wasmHostOld)
	if expectedRequests > 0 {
		info := ctx.Chain.MempoolInfo()
		expectedRequests += info.OutPoolCounter
	} else {
		expectedRequests = -expectedRequests
	}

	result := ctx.Chain.WaitForRequestsThrough(expectedRequests, maxWait...)
	_ = wasmhost.Connect(ctx.wc)
	return result
}
