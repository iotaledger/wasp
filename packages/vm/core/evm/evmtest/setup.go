package evmtest

import (
	"crypto/ecdsa"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type SoloChainEnv struct {
	t          testing.TB
	solo       *solo.Solo
	Chain      *solo.Chain
	evmChainID uint16
	evmChain   *jsonrpc.EVMChain
}

type iscCallOptions struct {
	wallet *cryptolib.KeyPair
}

type ethCallOptions struct {
	sender   *ecdsa.PrivateKey
	value    *big.Int
	gasLimit uint64
	gasPrice *big.Int
}

func InitEVM(t testing.TB) *SoloChainEnv {
	env := solo.New(t, &solo.InitOptions{
		Debug:             true,
		PrintStackTrace:   true,
		GasBurnLogEnabled: false,
	})
	return InitEVMWithSolo(t, env)
}

func InitEVMWithSolo(t testing.TB, env *solo.Solo) *SoloChainEnv {
	soloChain, _ := env.NewChainExt(nil, 0, "evmchain", evm.DefaultChainID, governance.DefaultBlockKeepAmount)
	return &SoloChainEnv{
		t:          t,
		solo:       env,
		Chain:      soloChain,
		evmChainID: evm.DefaultChainID,
		evmChain:   soloChain.EVM(),
	}
}

func (e *SoloChainEnv) parseISCCallOptions(opts []iscCallOptions) iscCallOptions {
	if len(opts) == 0 {
		opts = []iscCallOptions{{}}
	}
	opt := opts[0]
	if opt.wallet == nil {
		opt.wallet = e.Chain.ChainAdmin
	}
	return opt
}

func (e *SoloChainEnv) resolveError(err error) error {
	if err == nil {
		return nil
	}
	if vmError, ok := err.(*isc.UnresolvedVMError); ok {
		resolvedErr := e.Chain.ResolveVMError(vmError)
		return resolvedErr.AsGoError()
	}
	return err
}

func (e *SoloChainEnv) getBlockNumber() uint64 {
	n := e.evmChain.BlockNumber()
	return n.Uint64()
}

func (e *SoloChainEnv) getCode(addr common.Address) []byte {
	ret, err := e.evmChain.Code(addr, nil)
	require.NoError(e.t, err)
	return ret
}

func (e *SoloChainEnv) getEVMGasRatio() util.Ratio32 {
	ret, err := e.Chain.CallView(governance.ViewGetEVMGasRatio.Message())
	require.NoError(e.t, err)
	return lo.Must(governance.ViewGetEVMGasRatio.DecodeOutput(ret))
}

func (e *SoloChainEnv) setEVMGasRatio(newGasRatio util.Ratio32, opts ...iscCallOptions) error {
	opt := e.parseISCCallOptions(opts)
	req := solo.NewCallParams(governance.FuncSetEVMGasRatio.Message(newGasRatio))
	_, err := e.Chain.PostRequestSync(req, opt.wallet)
	return err
}

func (e *SoloChainEnv) setFeePolicy(p gas.FeePolicy, opts ...iscCallOptions) error { //nolint:unparam
	opt := e.parseISCCallOptions(opts)
	req := solo.NewCallParams(governance.FuncSetFeePolicy.Message(&p))
	_, err := e.Chain.PostRequestSync(req, opt.wallet)
	return err
}

func (e *SoloChainEnv) getNonce(addr common.Address) uint64 {
	nonce, err := e.evmChain.TransactionCount(addr, nil)
	require.NoError(e.t, err)
	return nonce
}

func (e *SoloChainEnv) contractFromABI(address common.Address, abiJSON string, defaultSender *ecdsa.PrivateKey) *IscContractInstance {
	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	require.NoError(e.t, err)
	return &IscContractInstance{
		EVMContractInstance: &EVMContractInstance{
			chain:         e,
			defaultSender: defaultSender,
			address:       address,
			abi:           parsedABI,
		},
	}
}

func (e *SoloChainEnv) ISCMagicSandbox(defaultSender *ecdsa.PrivateKey) *IscContractInstance {
	return e.contractFromABI(iscmagic.Address, iscmagic.SandboxABI, defaultSender)
}

func (e *SoloChainEnv) ISCMagicUtil(defaultSender *ecdsa.PrivateKey) *IscContractInstance {
	return e.contractFromABI(iscmagic.Address, iscmagic.UtilABI, defaultSender)
}

func (e *SoloChainEnv) ISCMagicAccounts(defaultSender *ecdsa.PrivateKey) *IscContractInstance {
	return e.contractFromABI(iscmagic.Address, iscmagic.AccountsABI, defaultSender)
}

func (e *SoloChainEnv) ISCMagicPrivileged(defaultSender *ecdsa.PrivateKey) *IscContractInstance {
	return e.contractFromABI(iscmagic.Address, iscmagic.PrivilegedABI, defaultSender)
}

func (e *SoloChainEnv) ERC20Coin(defaultSender *ecdsa.PrivateKey, coinType coin.Type) *IscContractInstance {
	ntABI, err := abi.JSON(strings.NewReader(iscmagic.ERC20CoinABI))
	require.NoError(e.t, err)
	return &IscContractInstance{
		EVMContractInstance: &EVMContractInstance{
			chain:         e,
			defaultSender: defaultSender,
			address:       iscmagic.ERC20CoinAddress(coinType),
			abi:           ntABI,
		},
	}
}

func (e *SoloChainEnv) deployISCTestContract(creator *ecdsa.PrivateKey) *iscTestContractInstance {
	return &iscTestContractInstance{e.DeployContract(creator, evmtest.ISCTestContractABI, evmtest.ISCTestContractBytecode)}
}

func (e *SoloChainEnv) deployStorageContract(creator *ecdsa.PrivateKey) *storageContractInstance {
	return &storageContractInstance{e.DeployContract(creator, evmtest.StorageContractABI, evmtest.StorageContractBytecode, uint32(42))}
}

func (e *SoloChainEnv) deployERC20Contract(creator *ecdsa.PrivateKey, name, symbol string) *erc20ContractInstance {
	return &erc20ContractInstance{e.DeployContract(creator, evmtest.ERC20ContractABI, evmtest.ERC20ContractBytecode, name, symbol)}
}

func (e *SoloChainEnv) deployLoopContract(creator *ecdsa.PrivateKey) *loopContractInstance {
	return &loopContractInstance{e.DeployContract(creator, evmtest.LoopContractABI, evmtest.LoopContractBytecode)}
}

func (e *SoloChainEnv) deployFibonacciContract(creator *ecdsa.PrivateKey) *fibonacciContractInstance {
	return &fibonacciContractInstance{e.DeployContract(creator, evmtest.FibonacciContractABI, evmtest.FibonacciContractByteCode)}
}

func (e *SoloChainEnv) maxGasLimit() uint64 {
	fp := e.Chain.GetGasFeePolicy()
	gl := e.Chain.GetGasLimits()
	return gas.EVMCallGasLimit(gl, &fp.EVMGasRatio)
}

func (e *SoloChainEnv) DeployContract(creator *ecdsa.PrivateKey, abiJSON string, bytecode []byte, args ...any) *EVMContractInstance {
	contractAddr, contractABI := e.Chain.DeployEVMContract(creator, abiJSON, bytecode, big.NewInt(0), args...)

	return &EVMContractInstance{
		chain:         e,
		defaultSender: creator,
		address:       contractAddr,
		abi:           contractABI,
	}
}

func (e *SoloChainEnv) registerERC20Coin(kp *cryptolib.KeyPair, coinType coin.Type) error {
	_, err := e.Chain.PostRequestOffLedger(
		solo.NewCallParams(evm.FuncRegisterERC20Coin.Message(coinType)).
			WithMaxAffordableGasBudget(),
		kp,
	)
	return err
}

func (e *SoloChainEnv) latestEVMTxs() types.Transactions {
	block, err := e.Chain.EVM().BlockByNumber(nil)
	require.NoError(e.t, err)
	return block.Transactions()
}

func (e *SoloChainEnv) LastBlockEVMLogs() []*types.Log {
	blockNumber := e.getBlockNumber()
	logs, err := e.evmChain.Logs(&ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(blockNumber),
		ToBlock:   new(big.Int).SetUint64(blockNumber),
	}, &jsonrpc.ParametersDefault().Logs)
	require.NoError(e.t, err)
	return logs
}
