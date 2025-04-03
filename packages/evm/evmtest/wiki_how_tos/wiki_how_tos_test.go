package wiki_how_tos_test

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
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

//go:generate sh -c "solc --abi --bin --overwrite @iscmagic=`realpath ../../../vm/core/evm/iscmagic` Entropy.sol -o ."
var (
	//go:embed Entropy.abi
	EntropyContractABI string
	//go:embed Entropy.bin
	EntropyContractBytecodeHex string
	EntropyContractBytecode    = common.FromHex(strings.TrimSpace(EntropyContractBytecodeHex))
)

func TestBaseBalance(t *testing.T) {
	env := evmtest.InitEVMWithSolo(t, solo.New(t))
	privateKey, deployer := env.Chain.NewEthereumAccountWithL2Funds()

	instance := env.DeployContract(privateKey, GetBalanceContractABI, GetBalanceContractBytecode)

	balance, _ := env.Chain.EVM().Balance(deployer, nil)
	var value uint64
	instance.CallFnExpectEvent(nil, "GotBaseBalance", &value, "getBalanceBaseTokens")
	realBalance := util.BaseTokensDecimalsToEthereumDecimals(coin.Value(value), parameters.BaseTokenDecimals)
	assert.Equal(t, balance, realBalance)
}

func TestNFTBalance(t *testing.T) {
	t.Skip("TODO")
	/*
	env := evmtest.InitEVMWithSolo(t, solo.New(t))
	privateKey, deployer := env.Chain.NewEthereumAccountWithL2Funds()

	instance := env.DeployContract(privateKey, GetBalanceContractABI, GetBalanceContractBytecode)

	// get the agentId of the contract deployer
	senderAgentID := isc.NewEthereumAddressAgentID(deployer)

	// mint an NFToken to the contract deployer
	// and check if the balance returned by the contract is correct
	mockMetaData := []byte("sesa")
	nfti, err := env.Chain.Env.MintNFTL1(env.Chain.OwnerPrivateKey, env.Chain.OwnerAddress(), mockMetaData)
	require.NoError(t, err)
	env.Chain.MustDepositNFT(nfti, env.Chain.OwnerAgentID(), env.Chain.OwnerPrivateKey)

	transfer := isc.NewEmptyAssets()
	transfer.AddObject(nfti.ID)

	// send the NFT to the contract deployer
	err = env.Chain.SendFromL2ToL2Account(transfer, senderAgentID, env.Chain.OwnerPrivateKey)
	require.NoError(t, err)

	// get the NFT balance of the contract deployer
	nftBalance := new(big.Int)
	instance.CallFnExpectEvent(nil, "GotNFTIDs", &nftBalance, "getBalanceNFTs")
	assert.Equal(t, int64(1), nftBalance.Int64())
	*/
}

func TestAgentID(t *testing.T) {
	env := evmtest.InitEVMWithSolo(t, solo.New(t))
	privateKey, deployer := env.Chain.NewEthereumAccountWithL2Funds()

	instance := env.DeployContract(privateKey, GetBalanceContractABI, GetBalanceContractBytecode)

	// get the agentId of the contract deployer
	senderAgentID := isc.NewEthereumAddressAgentID(deployer)

	// get the agnetId of the contract deployer
	// and compare it with the agentId returned by the contract
	var agentID []byte
	instance.CallFnExpectEvent(nil, "GotAgentID", &agentID, "getAgentID")
	assert.Equal(t, senderAgentID.Bytes(), agentID)
}

func TestEntropy(t *testing.T) {
	env := evmtest.InitEVMWithSolo(t, solo.New(t))
	privateKey, _ := env.Chain.NewEthereumAccountWithL2Funds()

	instance := env.DeployContract(privateKey, EntropyContractABI, EntropyContractBytecode)

	// get the entropy of the contract
	// and check if it is different from the previous one
	var entropy [32]byte
	instance.CallFnExpectEvent(nil, "EntropyEvent", &entropy, "emitEntropy")
	var entropy2 [32]byte
	instance.CallFnExpectEvent(nil, "EntropyEvent", &entropy2, "emitEntropy")
	assert.NotEqual(t, entropy, entropy2)
}

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}
