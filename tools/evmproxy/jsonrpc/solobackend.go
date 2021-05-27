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

func NewSoloBackend(alloc core.GenesisAlloc) *SoloBackend {
	env := solo.New(solo.NewFakeTestingT("evmproxy"), true, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "evmchain", evmchain.Interface.ProgramHash,
		evmchain.FieldGenesisAlloc, evmchain.EncodeGenesisAlloc(alloc),
	)
	require.NoError(env.T, err)
	return &SoloBackend{env, chain}
}

func (s *SoloBackend) PostRequest(keyPair *ed25519.KeyPair, transfer map[ledgerstate.Color]uint64, scName string, funName string, optSize int, params ...interface{}) error {
	req, toUpload := solo.NewCallParamsOptimized(scName, funName, optSize, params...)
	req.WithTransfers(transfer)
	for _, v := range toUpload {
		s.Chain.Env.PutBlobDataIntoRegistry(v)
	}
	_, err := s.Chain.PostRequestSync(req, keyPair)
	return err
}

func (s *SoloBackend) CallView(scName string, funName string, params ...interface{}) (dict.Dict, error) {
	return s.Chain.CallView(scName, funName, params...)
}
