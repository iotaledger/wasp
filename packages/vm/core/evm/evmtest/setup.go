package evmtest

import (
	"crypto/ecdsa"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/origin"
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

func InitEVM(t testing.TB, deployMagicWrap bool, nativeContracts ...*coreutil.ContractProcessor) *SoloChainEnv {
	env := solo.New(t, &solo.InitOptions{
		AutoAdjustStorageDeposit: true,
		Debug:                    true,
		PrintStackTrace:          true,
		GasBurnLogEnabled:        false,
	})
	for _, c := range nativeContracts {
		env = env.WithNativeContract(c)
	}
	return InitEVMWithSolo(t, env, deployMagicWrap)
}

func InitEVMWithSolo(t testing.TB, env *solo.Solo, deployMagicWrap bool) *SoloChainEnv {
	soloChain, _ := env.NewChainExt(nil, 0, "evmchain", dict.Dict{origin.ParamDeployBaseTokenMagicWrap: codec.Encode(deployMagicWrap)})
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
		opt.wallet = e.Chain.OriginatorPrivateKey
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
	ret, err := e.Chain.CallView(governance.Contract.Name, governance.ViewGetEVMGasRatio.Name)
	require.NoError(e.t, err)
	ratio, err := codec.DecodeRatio32(ret.Get(governance.ParamEVMGasRatio))
	require.NoError(e.t, err)
	return ratio
}

func (e *SoloChainEnv) setEVMGasRatio(newGasRatio util.Ratio32, opts ...iscCallOptions) error {
	opt := e.parseISCCallOptions(opts)
	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetEVMGasRatio.Name, governance.ParamEVMGasRatio, newGasRatio.Bytes())
	_, err := e.Chain.PostRequestSync(req, opt.wallet)
	return err
}

func (e *SoloChainEnv) setFeePolicy(p gas.FeePolicy, opts ...iscCallOptions) error { //nolint:unparam
	opt := e.parseISCCallOptions(opts)
	req := solo.NewCallParams(
		governance.Contract.Name, governance.FuncSetFeePolicy.Name,
		governance.ParamFeePolicyBytes,
		p.Bytes(),
	)
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

func (e *SoloChainEnv) ERC20BaseTokens(defaultSender *ecdsa.PrivateKey) *IscContractInstance {
	erc20BaseABI, err := abi.JSON(strings.NewReader(iscmagic.ERC20BaseTokensABI))
	require.NoError(e.t, err)
	return &IscContractInstance{
		EVMContractInstance: &EVMContractInstance{
			chain:         e,
			defaultSender: defaultSender,
			address:       iscmagic.ERC20BaseTokensAddress,
			abi:           erc20BaseABI,
		},
	}
}

func (e *SoloChainEnv) ERC20NativeTokens(defaultSender *ecdsa.PrivateKey, foundrySN uint32) *IscContractInstance {
	erc20BaseABI, err := abi.JSON(strings.NewReader(iscmagic.ERC20NativeTokensABI))
	require.NoError(e.t, err)
	return &IscContractInstance{
		EVMContractInstance: &EVMContractInstance{
			chain:         e,
			defaultSender: defaultSender,
			address:       iscmagic.ERC20NativeTokensAddress(foundrySN),
			abi:           erc20BaseABI,
		},
	}
}

func (e *SoloChainEnv) ERC20ExternalNativeTokens(defaultSender *ecdsa.PrivateKey, addr common.Address) *IscContractInstance {
	erc20BaseABI, err := abi.JSON(strings.NewReader(iscmagic.ERC20ExternalNativeTokensABI))
	require.NoError(e.t, err)
	return &IscContractInstance{
		EVMContractInstance: &EVMContractInstance{
			chain:         e,
			defaultSender: defaultSender,
			address:       addr,
			abi:           erc20BaseABI,
		},
	}
}

func (e *SoloChainEnv) ERC721NFTs(defaultSender *ecdsa.PrivateKey) *IscContractInstance {
	erc721ABI, err := abi.JSON(strings.NewReader(iscmagic.ERC721NFTsABI))
	require.NoError(e.t, err)
	return &IscContractInstance{
		EVMContractInstance: &EVMContractInstance{
			chain:         e,
			defaultSender: defaultSender,
			address:       iscmagic.ERC721NFTsAddress,
			abi:           erc721ABI,
		},
	}
}

func (e *SoloChainEnv) ERC721NFTCollection(defaultSender *ecdsa.PrivateKey, collectionID iotago.NFTID) *IscContractInstance {
	erc721NFTCollectionABI, err := abi.JSON(strings.NewReader(iscmagic.ERC721NFTCollectionABI))
	require.NoError(e.t, err)
	return &IscContractInstance{
		EVMContractInstance: &EVMContractInstance{
			chain:         e,
			defaultSender: defaultSender,
			address:       iscmagic.ERC721NFTCollectionAddress(collectionID),
			abi:           erc721NFTCollectionABI,
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

func (e *SoloChainEnv) deployERC20ExampleContract(creator *ecdsa.PrivateKey) *erc20ContractInstance {
	return &erc20ContractInstance{e.DeployContract(creator, evmtest.ERC20ExampleContractABI, evmtest.ERC20ExampleContractBytecode)}
}

func (e *SoloChainEnv) maxGasLimit() uint64 {
	fp := e.Chain.GetGasFeePolicy()
	gl := e.Chain.GetGasLimits()
	return gas.EVMCallGasLimit(gl, &fp.EVMGasRatio)
}

func (e *SoloChainEnv) DeployContract(creator *ecdsa.PrivateKey, abiJSON string, bytecode []byte, args ...interface{}) *EVMContractInstance {
	contractAddr, contractABI := e.Chain.DeployEVMContract(creator, abiJSON, bytecode, big.NewInt(0), args...)

	return &EVMContractInstance{
		chain:         e,
		defaultSender: creator,
		address:       contractAddr,
		abi:           contractABI,
	}
}

func (e *SoloChainEnv) registerERC20NativeToken(
	foundryOwner *cryptolib.KeyPair,
	foundrySN uint32,
	tokenName, tokenTickerSymbol string,
	tokenDecimals uint8,
) error {
	_, err := e.Chain.PostRequestOffLedger(solo.NewCallParams(evm.Contract.Name, evm.FuncRegisterERC20NativeToken.Name, dict.Dict{
		evm.FieldFoundrySN:         codec.EncodeUint32(foundrySN),
		evm.FieldTokenName:         codec.EncodeString(tokenName),
		evm.FieldTokenTickerSymbol: codec.EncodeString(tokenTickerSymbol),
		evm.FieldTokenDecimals:     codec.EncodeUint8(tokenDecimals),
	}).WithMaxAffordableGasBudget(), foundryOwner)
	return err
}

func (e *SoloChainEnv) registerERC721NFTCollection(collectionOwner *cryptolib.KeyPair, collectionID iotago.NFTID) error {
	_, err := e.Chain.PostRequestOffLedger(solo.NewCallParams(evm.Contract.Name, evm.FuncRegisterERC721NFTCollection.Name, dict.Dict{
		evm.FieldNFTCollectionID: codec.EncodeNFTID(collectionID),
	}).WithMaxAffordableGasBudget(), collectionOwner)
	return err
}
