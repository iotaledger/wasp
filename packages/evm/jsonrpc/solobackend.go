package jsonrpc

import (
	"github.com/ethereum/go-ethereum/core"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/contracts/native/evmchain"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

type SoloBackend struct {
	Env   *solo.Solo
	Chain *solo.Chain
}

var _ ChainBackend = &SoloBackend{}

func NewSoloBackend(alloc core.GenesisAlloc) *SoloBackend {
	env := solo.New(solo.NewFakeTestingT("evmproxy"), true, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "evmchain", evmchain.Interface.ProgramHash,
		evmchain.FieldGenesisAlloc, evmchain.EncodeGenesisAlloc(alloc),
	)
	require.NoError(env.T, err)
	return &SoloBackend{env, chain}
}

func (s *SoloBackend) PostOnLedgerRequest(keyPair *ed25519.KeyPair, scName, funName string, transfer map[ledgerstate.Color]uint64, args dict.Dict) error {
	_, err := s.Chain.PostRequestSync(
		solo.NewCallParamsFromDic(scName, funName, args).WithTransfers(transfer),
		keyPair,
	)
	return err
}

func (s *SoloBackend) PostOffLedgerRequest(keyPair *ed25519.KeyPair, scName, funName string, transfer map[ledgerstate.Color]uint64, args dict.Dict) error {
	_, err := s.Chain.PostRequestOffLedger(
		solo.NewCallParamsFromDic(scName, funName, args).WithTransfers(transfer),
		keyPair,
	)
	return err
}

func (s *SoloBackend) CallView(scName, funName string, args dict.Dict) (dict.Dict, error) {
	return s.Chain.CallView(scName, funName, args)
}
