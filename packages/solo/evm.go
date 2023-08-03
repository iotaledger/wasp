package solo

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/tracers"

	hivedb "github.com/iotaledger/hive.go/kvstore/database"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chainutil"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
)

// jsonRPCSoloBackend is the implementation of [jsonrpc.ChainBackend] for Solo
// tests.
type jsonRPCSoloBackend struct {
	Chain     *Chain
	baseToken *parameters.BaseToken
}

func newJSONRPCSoloBackend(chain *Chain, baseToken *parameters.BaseToken) jsonrpc.ChainBackend {
	return &jsonRPCSoloBackend{Chain: chain, baseToken: baseToken}
}

func (b *jsonRPCSoloBackend) EVMSendTransaction(tx *types.Transaction) error {
	_, err := b.Chain.PostEthereumTransaction(tx)
	return err
}

func (b *jsonRPCSoloBackend) EVMCall(aliasOutput *isc.AliasOutputWithID, callMsg ethereum.CallMsg) ([]byte, error) {
	return chainutil.EVMCall(b.Chain, aliasOutput, callMsg)
}

func (b *jsonRPCSoloBackend) EVMEstimateGas(aliasOutput *isc.AliasOutputWithID, callMsg ethereum.CallMsg) (uint64, error) {
	return chainutil.EVMEstimateGas(b.Chain, aliasOutput, callMsg)
}

func (b *jsonRPCSoloBackend) EVMTraceTransaction(
	aliasOutput *isc.AliasOutputWithID,
	blockTime time.Time,
	iscRequestsInBlock []isc.Request,
	txIndex uint64,
	tracer tracers.Tracer,
) error {
	return chainutil.EVMTraceTransaction(
		b.Chain,
		aliasOutput,
		blockTime,
		iscRequestsInBlock,
		txIndex,
		tracer,
	)
}

func (b *jsonRPCSoloBackend) ISCCallView(chainState state.State, scName, funName string, args dict.Dict) (dict.Dict, error) {
	return b.Chain.CallViewAtState(chainState, scName, funName, args)
}

func (b *jsonRPCSoloBackend) ISCLatestAliasOutput() (*isc.AliasOutputWithID, error) {
	latestAliasOutput, err := b.Chain.LatestAliasOutput(chain.ActiveOrCommittedState)
	if err != nil {
		return nil, fmt.Errorf("could not get latest AliasOutput: %w", err)
	}
	return latestAliasOutput, nil
}

func (b *jsonRPCSoloBackend) ISCLatestState() state.State {
	latestState, err := b.Chain.LatestState(chain.ActiveOrCommittedState)
	if err != nil {
		panic(err)
	}
	return latestState
}

func (b *jsonRPCSoloBackend) ISCStateByBlockIndex(blockIndex uint32) (state.State, error) {
	return b.Chain.store.StateByIndex(blockIndex)
}

func (b *jsonRPCSoloBackend) ISCStateByTrieRoot(trieRoot trie.Hash) (state.State, error) {
	return b.Chain.store.StateByTrieRoot(trieRoot)
}

func (b *jsonRPCSoloBackend) BaseToken() *parameters.BaseToken {
	return b.baseToken
}

func (b *jsonRPCSoloBackend) ISCChainID() *isc.ChainID {
	return &b.Chain.ChainID
}

func (ch *Chain) EVM() *jsonrpc.EVMChain {
	return jsonrpc.NewEVMChain(
		newJSONRPCSoloBackend(ch, parameters.L1().BaseToken),
		ch.Env.publisher,
		true,
		hivedb.EngineMapDB,
		"",
		ch.log,
	)
}

func (ch *Chain) PostEthereumTransaction(tx *types.Transaction) (dict.Dict, error) {
	req, err := isc.NewEVMOffLedgerTxRequest(ch.ChainID, tx)
	if err != nil {
		return nil, err
	}
	return ch.RunOffLedgerRequest(req)
}

var EthereumAccounts []*ecdsa.PrivateKey

func init() {
	EthereumAccounts = make([]*ecdsa.PrivateKey, 4)
	EthereumAccounts[0], _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	EthereumAccounts[1], _ = crypto.HexToECDSA("289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032")
	EthereumAccounts[2], _ = crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
	EthereumAccounts[3], _ = crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
}

func (ch *Chain) EthereumAccountByIndexWithL2Funds(i int, baseTokens ...uint64) (*ecdsa.PrivateKey, common.Address) {
	key := EthereumAccounts[i]
	addr := crypto.PubkeyToAddress(key.PublicKey)
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

func (ch *Chain) NewEthereumAccountWithL2Funds(baseTokens ...uint64) (*ecdsa.PrivateKey, common.Address) {
	key, addr := NewEthereumAccount()
	ch.GetL2FundsFromFaucet(isc.NewEthereumAddressAgentID(addr), baseTokens...)
	return key, addr
}
