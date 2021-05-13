package service

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iotaledger/wasp/contracts/native/evmchain"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

func NewSoloEVMChain(alloc core.GenesisAlloc) EVMChain {
	env := solo.New(solo.NewFakeTestingT("evmproxy"), true, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "evmchain", evmchain.Interface.ProgramHash,
		evmchain.FieldGenesisAlloc, evmchain.EncodeGenesisAlloc(alloc),
	)
	require.NoError(env.T, err)
	return &soloEVMChain{env, chain}
}

type soloEVMChain struct {
	env   *solo.Solo
	chain *solo.Chain
}

func (s *soloEVMChain) SendTransaction(tx *types.Transaction) {
	txdata, err := tx.MarshalBinary()
	require.NoError(s.env.T, err)

	req, toUpload := solo.NewCallParamsOptimized(evmchain.Interface.Name, evmchain.FuncSendTransaction, 1024,
		evmchain.FieldTransactionData, txdata,
	)
	req.WithIotas(1)
	for _, v := range toUpload {
		s.chain.Env.PutBlobDataIntoRegistry(v)
	}

	_, err = s.chain.PostRequestSync(req, nil)
	require.NoError(s.env.T, err)
}

func (s *soloEVMChain) TransactionCount(address common.Address, blockNumber *big.Int) uint64 {
	params := []interface{}{
		evmchain.FieldAddress, address.Bytes(),
	}
	if blockNumber != nil {
		params = append(params, evmchain.FieldBlockNumber, blockNumber.Bytes())
	}
	ret, err := s.chain.CallView(evmchain.Interface.Name, evmchain.FuncGetNonce, params...)
	require.NoError(s.env.T, err)

	n, _, err := codec.DecodeUint64(ret.MustGet(evmchain.FieldResult))
	require.NoError(s.env.T, err)
	return n
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
