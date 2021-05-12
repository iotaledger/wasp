package service

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/contracts/native/evmchain"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

var (
	faucetKey, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	faucetAddress = crypto.PubkeyToAddress(faucetKey.PublicKey)
	faucetSupply  = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))
)

func NewSoloEVMChain() EVMChain {
	env := solo.New(solo.NewFakeTestingT("evmproxy"), false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "evmchain", evmchain.Interface.ProgramHash,
		evmchain.FieldGenesisAlloc, evmchain.EncodeGenesisAlloc(map[common.Address]core.GenesisAccount{
			faucetAddress: {Balance: faucetSupply},
		}),
	)
	require.NoError(env.T, err)
	return &soloEVMChain{env, chain}
}

type soloEVMChain struct {
	env   *solo.Solo
	chain *solo.Chain
}

func (s *soloEVMChain) BlockByNumber(blockNumber *big.Int) *types.Block {
	params := []interface{}{}
	if blockNumber != nil {
		params = append(params, evmchain.FieldBlockNumber, blockNumber.Bytes())
	}
	ret, err := s.chain.CallView(evmchain.Interface.Name, evmchain.FuncGetBlockByNumber, params...)
	require.NoError(s.env.T, err)

	if !ret.MustHas(evmchain.FieldResult) {
		return nil
	}

	block, err := evmchain.DecodeBlock(ret.MustGet(evmchain.FieldResult))
	require.NoError(s.env.T, err)
	return block
}

func (s *soloEVMChain) Balance(address common.Address, blockNumber *big.Int) *big.Int {
	params := []interface{}{
		evmchain.FieldAddress, address.Bytes(),
	}
	if blockNumber != nil {
		params = append(params, evmchain.FieldBlockNumber, blockNumber.Bytes())
	}
	ret, err := s.chain.CallView(evmchain.Interface.Name, evmchain.FuncGetBalance, params...)
	require.NoError(s.env.T, err)

	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(evmchain.FieldBalance))
	return bal
}

func (s *soloEVMChain) BlockNumber() *big.Int {
	ret, err := s.chain.CallView(evmchain.Interface.Name, evmchain.FuncGetBlockNumber)
	require.NoError(s.env.T, err)

	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(evmchain.FieldResult))
	return bal
}

var _ EVMChain = &soloEVMChain{}
