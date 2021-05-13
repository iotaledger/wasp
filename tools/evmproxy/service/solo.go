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

func NewSoloEVMChain(alloc core.GenesisAlloc) *SoloEVMChain {
	env := solo.New(solo.NewFakeTestingT("evmproxy"), true, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "evmchain", evmchain.Interface.ProgramHash,
		evmchain.FieldGenesisAlloc, evmchain.EncodeGenesisAlloc(alloc),
	)
	require.NoError(env.T, err)
	return &SoloEVMChain{env, chain}
}

type SoloEVMChain struct {
	Env   *solo.Solo
	Chain *solo.Chain
}

var _ EVMChain = &SoloEVMChain{}

func (s *SoloEVMChain) SendTransaction(tx *types.Transaction) {
	txdata, err := tx.MarshalBinary()
	require.NoError(s.Env.T, err)

	req, toUpload := solo.NewCallParamsOptimized(evmchain.Interface.Name, evmchain.FuncSendTransaction, 1024,
		evmchain.FieldTransactionData, txdata,
	)
	req.WithIotas(1)
	for _, v := range toUpload {
		s.Chain.Env.PutBlobDataIntoRegistry(v)
	}

	_, err = s.Chain.PostRequestSync(req, nil)
	require.NoError(s.Env.T, err)
}

func (s *SoloEVMChain) TransactionCount(address common.Address, blockNumber *big.Int) uint64 {
	params := []interface{}{
		evmchain.FieldAddress, address.Bytes(),
	}
	if blockNumber != nil {
		params = append(params, evmchain.FieldBlockNumber, blockNumber.Bytes())
	}
	ret, err := s.Chain.CallView(evmchain.Interface.Name, evmchain.FuncGetNonce, params...)
	require.NoError(s.Env.T, err)

	n, _, err := codec.DecodeUint64(ret.MustGet(evmchain.FieldResult))
	require.NoError(s.Env.T, err)
	return n
}

func (s *SoloEVMChain) BlockByNumber(blockNumber *big.Int) *types.Block {
	params := []interface{}{}
	if blockNumber != nil {
		params = append(params, evmchain.FieldBlockNumber, blockNumber.Bytes())
	}
	ret, err := s.Chain.CallView(evmchain.Interface.Name, evmchain.FuncGetBlockByNumber, params...)
	require.NoError(s.Env.T, err)

	if !ret.MustHas(evmchain.FieldResult) {
		return nil
	}

	block, err := evmchain.DecodeBlock(ret.MustGet(evmchain.FieldResult))
	require.NoError(s.Env.T, err)
	return block
}

func (s *SoloEVMChain) Balance(address common.Address, blockNumber *big.Int) *big.Int {
	params := []interface{}{
		evmchain.FieldAddress, address.Bytes(),
	}
	if blockNumber != nil {
		params = append(params, evmchain.FieldBlockNumber, blockNumber.Bytes())
	}
	ret, err := s.Chain.CallView(evmchain.Interface.Name, evmchain.FuncGetBalance, params...)
	require.NoError(s.Env.T, err)

	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(evmchain.FieldBalance))
	return bal
}

func (s *SoloEVMChain) BlockNumber() *big.Int {
	ret, err := s.Chain.CallView(evmchain.Interface.Name, evmchain.FuncGetBlockNumber)
	require.NoError(s.Env.T, err)

	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(evmchain.FieldResult))
	return bal
}

var _ EVMChain = &SoloEVMChain{}
