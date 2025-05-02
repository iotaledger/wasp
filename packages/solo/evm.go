package solo

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/tracers"

	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/wasp/packages/chainutil"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// jsonRPCSoloBackend is the implementation of [jsonrpc.ChainBackend] for Solo
// tests.
type jsonRPCSoloBackend struct {
	Chain     *Chain
	snapshots []*Snapshot
}

func newJSONRPCSoloBackend(chain *Chain) jsonrpc.ChainBackend {
	return &jsonRPCSoloBackend{Chain: chain}
}

func (b *jsonRPCSoloBackend) FeePolicy(blockIndex uint32) (*gas.FeePolicy, error) {
	state, err := b.ISCStateByBlockIndex(blockIndex)
	if err != nil {
		return nil, err
	}
	ret, err := b.ISCCallView(state, governance.ViewGetFeePolicy.Message())
	if err != nil {
		return nil, err
	}
	return governance.ViewGetFeePolicy.DecodeOutput(ret)
}

func (b *jsonRPCSoloBackend) EVMSendTransaction(tx *types.Transaction) error {
	_, err := b.Chain.PostEthereumTransaction(tx)
	return err
}

func (b *jsonRPCSoloBackend) EVMCall(anchor *isc.StateAnchor, callMsg ethereum.CallMsg, l1Params *parameters.L1Params) ([]byte, error) {
	return chainutil.EVMCall(
		anchor,
		l1Params,
		b.Chain.store,
		b.Chain.proc,
		b.Chain.log,
		callMsg,
	)
}

func (b *jsonRPCSoloBackend) EVMEstimateGas(anchor *isc.StateAnchor, callMsg ethereum.CallMsg, l1Params *parameters.L1Params) (uint64, error) {
	return chainutil.EVMEstimateGas(
		anchor,
		l1Params,
		b.Chain.store,
		b.Chain.proc,
		b.Chain.log,
		callMsg,
	)
}

func (b *jsonRPCSoloBackend) EVMTrace(
	anchor *isc.StateAnchor,
	blockTime time.Time,
	iscRequestsInBlock []isc.Request,
	enforceGasBurned []vm.EnforceGasBurned,
	tracer *tracers.Tracer,
	l1Params *parameters.L1Params,
) error {
	return chainutil.EVMTrace(
		anchor,
		l1Params,
		b.Chain.store,
		b.Chain.proc,
		b.Chain.log,
		blockTime,
		iscRequestsInBlock,
		enforceGasBurned,
		tracer,
	)
}

func (b *jsonRPCSoloBackend) ISCCallView(chainState state.State, msg isc.Message) (isc.CallArguments, error) {
	return b.Chain.CallViewAtState(chainState, msg)
}

func (b *jsonRPCSoloBackend) ISCLatestAnchor() (*isc.StateAnchor, error) {
	anchor := b.Chain.GetLatestAnchor()
	return anchor, nil
}

func (b *jsonRPCSoloBackend) ISCLatestState() (state.State, error) {
	return b.Chain.LatestState()
}

func (b *jsonRPCSoloBackend) ISCStateByBlockIndex(blockIndex uint32) (state.State, error) {
	return b.Chain.store.StateByIndex(blockIndex)
}

func (b *jsonRPCSoloBackend) ISCStateByTrieRoot(trieRoot trie.Hash) (state.State, error) {
	return b.Chain.store.StateByTrieRoot(trieRoot)
}

func (b *jsonRPCSoloBackend) ISCChainID() *isc.ChainID {
	return &b.Chain.ChainID
}

func (b *jsonRPCSoloBackend) RevertToSnapshot(i int) error {
	if i < 0 || i >= len(b.snapshots) {
		return errors.New("invalid snapshot index")
	}
	b.Chain.Env.RestoreSnapshot(b.snapshots[i])
	b.snapshots = b.snapshots[:i]
	return nil
}

func (b *jsonRPCSoloBackend) TakeSnapshot() (int, error) {
	b.snapshots = append(b.snapshots, b.Chain.Env.TakeSnapshot())
	return len(b.snapshots) - 1, nil
}

func (ch *Chain) EVM() *jsonrpc.EVMChain {
	return jsonrpc.NewEVMChain(
		newJSONRPCSoloBackend(ch),
		ch.Env.publisher,
		true,
		hivedb.EngineMapDB,
		"",
		ch.log,
	)
}

func (ch *Chain) PostEthereumTransaction(tx *types.Transaction) (isc.CallArguments, error) {
	req, err := isc.NewEVMOffLedgerTxRequest(ch.ChainID, tx)
	if err != nil {
		return nil, err
	}
	_, res, err := ch.RunOffLedgerRequest(req)
	return res, err
}

var EthereumAccounts [10]*ecdsa.PrivateKey

func init() {
	for i := range len(EthereumAccounts) {
		seed := crypto.Keccak256(fmt.Appendf(nil, "seed %d", i))
		key, err := crypto.ToECDSA(seed)
		if err != nil {
			panic(err)
		}
		EthereumAccounts[i] = key
	}
}

func EthereumAccountByIndex(i int) (*ecdsa.PrivateKey, common.Address) {
	key := EthereumAccounts[i]
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return key, addr
}

func (ch *Chain) EthereumAccountByIndexWithL2Funds(i int, baseTokens ...coin.Value) (*ecdsa.PrivateKey, common.Address) {
	key, addr := EthereumAccountByIndex(i)
	ch.GetL2FundsFromFaucet(isc.NewEthereumAddressAgentID(addr), baseTokens...)
	return key, addr
}

func NewEthereumAccount() (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	return key, crypto.PubkeyToAddress(key.PublicKey)
}

func (ch *Chain) NewEthereumAccountWithL2Funds(baseTokens ...coin.Value) (*ecdsa.PrivateKey, common.Address) {
	key, addr := NewEthereumAccount()
	ch.GetL2FundsFromFaucet(isc.NewEthereumAddressAgentID(addr), baseTokens...)
	return key, addr
}
