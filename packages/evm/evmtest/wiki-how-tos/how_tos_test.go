package wikihowtos_test

import (
	_ "embed"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmtest"
)

//go:generate sh -c "solc --abi --bin --overwrite @iscmagic=`realpath ../../../vm/core/evm/iscmagic` GetBalance.sol -o ."
var (
	//go:embed GetBalance.abi
	GetBalanceContractABI string
	//go:embed GetBalance.bin
	GetBalanceContractBytecodeHex string
	GetBalanceContractBytecode    = common.FromHex(strings.TrimSpace(GetBalanceContractBytecodeHex))
)

func TestBaseBalance(t *testing.T) {
	env := evmtest.InitEVMWithSolo(t, solo.New(t), true)
	privateKey, deployer := env.Chain.NewEthereumAccountWithL2Funds()

	instance := env.DeployContract(privateKey, GetBalanceContractABI, GetBalanceContractBytecode)

	balance, _ := env.Chain.EVM().Balance(deployer, nil)
	decimals := env.Chain.EVM().BaseToken().Decimals
	var value uint64
	instance.CallFnExpectEvent(nil, "GotBaseBalance", &value, "getBalanceBaseTokens")
	realBalance := util.BaseTokensDecimalsToEthereumDecimals(value, decimals)
	assert.Equal(t, balance, realBalance)
}

func TestNativeBalance(t *testing.T) {
	env := evmtest.InitEVMWithSolo(t, solo.New(t), true)
	privateKey, deployer := env.Chain.NewEthereumAccountWithL2Funds()

	instance := env.DeployContract(privateKey, GetBalanceContractABI, GetBalanceContractBytecode)

	// create a new native token on L1
	foundry, tokenID, err := env.Chain.NewNativeTokenParams(100000000000000).CreateFoundry()
	require.NoError(t, err)
	// the token id in bytes, used to call the contract
	nativeTokenIDBytes := isc.NativeTokenIDToBytes(tokenID)

	// mint some native tokens to the chain originator
	err = env.Chain.MintTokens(foundry, 10000000, env.Chain.OriginatorPrivateKey)
	require.NoError(t, err)

	// get the agentId of the contract deployer
	senderAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, deployer)

	// send some native tokens to the contract deployer
	// and check if the balance returned by the contract is correct
	err = env.Chain.SendFromL2ToL2AccountNativeTokens(tokenID, senderAgentID, 100000, env.Chain.OriginatorPrivateKey)
	require.NoError(t, err)

	nativeBalance := new(big.Int)
	instance.CallFnExpectEvent(nil, "GotNativeTokenBalance", &nativeBalance, "getBalanceNativeTokens", nativeTokenIDBytes)
	assert.Equal(t, int64(100000), nativeBalance.Int64())
}

func TestNFTBalance(t *testing.T) {
	env := evmtest.InitEVMWithSolo(t, solo.New(t), true)
	privateKey, deployer := env.Chain.NewEthereumAccountWithL2Funds()

	instance := env.DeployContract(privateKey, GetBalanceContractABI, GetBalanceContractBytecode)

	// get the agentId of the contract deployer
	senderAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, deployer)

	// mint an NFToken to the contract deployer
	// and check if the balance returned by the contract is correct
	mockMetaData := []byte("sesa")
	nfti, info, err := env.Chain.Env.MintNFTL1(env.Chain.OriginatorPrivateKey, env.Chain.OriginatorAddress, mockMetaData)
	require.NoError(t, err)
	env.Chain.MustDepositNFT(nfti, env.Chain.OriginatorAgentID, env.Chain.OriginatorPrivateKey)

	transfer := isc.NewEmptyAssets()
	transfer.AddNFTs(info.NFTID)

	// send the NFT to the contract deployer
	err = env.Chain.SendFromL2ToL2Account(transfer, senderAgentID, env.Chain.OriginatorPrivateKey)
	require.NoError(t, err)

	// get the NFT balance of the contract deployer
	nftBalance := new(big.Int)
	instance.CallFnExpectEvent(nil, "GotNFTIDs", &nftBalance, "getBalanceNFTs")
	assert.Equal(t, int64(1), nftBalance.Int64())
}

func TestAgentID(t *testing.T) {
	env := evmtest.InitEVMWithSolo(t, solo.New(t), true)
	privateKey, deployer := env.Chain.NewEthereumAccountWithL2Funds()

	instance := env.DeployContract(privateKey, GetBalanceContractABI, GetBalanceContractBytecode)

	// get the agentId of the contract deployer
	senderAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, deployer)

	// get the agnetId of the contract deployer
	// and compare it with the agentId returned by the contract
	var agentID []byte
	instance.CallFnExpectEvent(nil, "GotAgentID", &agentID, "getAgentID")
	assert.Equal(t, senderAgentID.Bytes(), agentID)
}
