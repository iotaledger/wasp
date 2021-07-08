package cluster

import (
	"bytes"
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/client/wallet/packages/seed"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/client/scclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/chainid"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type Chain struct {
	Description string

	OriginatorSeed *seed.Seed

	AllPeers       []int
	CommitteeNodes []int
	Quorum         uint16
	StateAddress   ledgerstate.Address

	ChainID chainid.ChainID

	Cluster *Cluster
}

func (ch *Chain) ChainAddress() ledgerstate.Address {
	return ch.ChainID.AsAddress()
}

func (ch *Chain) CommitteeAPIHosts() []string {
	return ch.Cluster.Config.APIHosts(ch.CommitteeNodes)
}

func (ch *Chain) CommitteePeeringHosts() []string {
	return ch.Cluster.Config.PeeringHosts(ch.CommitteeNodes)
}

func (ch *Chain) AllPeeringHosts() []string {
	return ch.Cluster.Config.PeeringHosts(ch.AllPeers)
}

func (ch *Chain) AllAPIHosts() []string {
	return ch.Cluster.Config.APIHosts(ch.AllPeers)
}

func (ch *Chain) OriginatorAddress() ledgerstate.Address {
	addr := ch.OriginatorSeed.Address(0).Address()
	return addr
}

func (ch *Chain) OriginatorID() *coretypes.AgentID {
	ret := coretypes.NewAgentID(ch.OriginatorAddress(), 0)
	return ret
}

func (ch *Chain) OriginatorKeyPair() *ed25519.KeyPair {
	return ch.OriginatorSeed.KeyPair(0)
}

func (ch *Chain) OriginatorClient() *chainclient.Client {
	return ch.Client(ch.OriginatorKeyPair())
}

func (ch *Chain) Client(sigScheme *ed25519.KeyPair, committeeIndex ...int) *chainclient.Client {
	i := 0
	if len(committeeIndex) == 1 {
		i = committeeIndex[0]
	}
	return chainclient.New(
		ch.Cluster.GoshimmerClient(),
		ch.Cluster.WaspClient(ch.CommitteeNodes[i]),
		ch.ChainID,
		sigScheme,
	)
}

func (ch *Chain) SCClient(contractHname coretypes.Hname, sigScheme *ed25519.KeyPair, committeeIndex ...int) *scclient.SCClient {
	return scclient.New(ch.Client(sigScheme, committeeIndex...), contractHname)
}

func (ch *Chain) CommitteeMultiClient() *multiclient.MultiClient {
	return multiclient.New(ch.CommitteeAPIHosts())
}

func (ch *Chain) DeployContract(name, progHashStr, description string, initParams map[string]interface{}) (*ledgerstate.Transaction, error) {
	programHash, err := hashing.HashValueFromBase58(progHashStr)
	if err != nil {
		return nil, err
	}

	params := map[string]interface{}{
		root.ParamName:        name,
		root.ParamProgramHash: programHash,
		root.ParamDescription: description,
	}
	for k, v := range initParams {
		params[k] = v
	}
	tx, err := ch.OriginatorClient().Post1Request(
		root.Interface.Hname(),
		coretypes.Hn(root.FuncDeployContract),
		chainclient.PostRequestParams{
			Args: requestargs.New().AddEncodeSimpleMany(codec.MakeDict(params)),
		},
	)
	if err != nil {
		return nil, err
	}
	err = ch.CommitteeMultiClient().WaitUntilAllRequestsProcessed(ch.ChainID, tx, 30*time.Second)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (ch *Chain) DeployWasmContract(name, description string, progBinary []byte, initParams map[string]interface{}) (*ledgerstate.Transaction, hashing.HashValue, error) {
	blobFieldValues := codec.MakeDict(map[string]interface{}{
		blob.VarFieldVMType:             vmtypes.WasmTime,
		blob.VarFieldProgramBinary:      progBinary,
		blob.VarFieldProgramDescription: description,
	})

	quorum := (2*len(ch.CommitteeAPIHosts()))/3 + 1
	programHash, tx, err := ch.OriginatorClient().UploadBlob(blobFieldValues, ch.CommitteeAPIHosts(), quorum, 256)
	if err != nil {
		return nil, hashing.NilHash, err
	}
	err = ch.CommitteeMultiClient().WaitUntilAllRequestsProcessed(ch.ChainID, tx, 30*time.Second)
	if err != nil {
		return nil, hashing.NilHash, err
	}

	progBinaryBack, err := ch.GetBlobFieldValue(programHash, blob.VarFieldProgramBinary)
	if err != nil {
		return nil, hashing.NilHash, err
	}
	if !bytes.Equal(progBinary, progBinaryBack) {
		return nil, hashing.NilHash, fmt.Errorf("!bytes.Equal(progBinary, progBinaryBack)")
	}
	fmt.Printf("---- blob installed correctly len = %d\n", len(progBinaryBack))

	params := make(map[string]interface{})
	for k, v := range initParams {
		params[k] = v
	}
	params[root.ParamName] = name
	params[root.ParamProgramHash] = programHash
	params[root.ParamDescription] = description

	args := requestargs.New().AddEncodeSimpleMany(codec.MakeDict(params))
	tx, err = ch.OriginatorClient().Post1Request(
		root.Interface.Hname(),
		coretypes.Hn(root.FuncDeployContract),
		chainclient.PostRequestParams{
			Args: args,
		},
	)
	if err != nil {
		return nil, hashing.NilHash, err
	}
	err = ch.CommitteeMultiClient().WaitUntilAllRequestsProcessed(ch.ChainID, tx, 30*time.Second)
	if err != nil {
		return nil, hashing.NilHash, err
	}

	return tx, programHash, nil
}

func (ch *Chain) GetBlobFieldValue(blobHash hashing.HashValue, field string) ([]byte, error) {
	v, err := ch.Cluster.WaspClient(0).CallView(
		ch.ChainID, blob.Interface.Hname(), blob.FuncGetBlobField,
		dict.Dict{
			blob.ParamHash:  blobHash[:],
			blob.ParamField: []byte(field),
		})
	if err != nil {
		return nil, err
	}
	if v.IsEmpty() {
		return nil, nil
	}
	ret, err := v.Get(blob.ParamBytes)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (ch *Chain) StartMessageCounter(expectations map[string]int) (*MessageCounter, error) {
	return NewMessageCounter(ch.Cluster, ch.CommitteeNodes, expectations)
}

func (ch *Chain) BlockIndex(committeeIndex ...int) (uint32, error) {
	cl := ch.SCClient(blocklog.Interface.Hname(), nil, committeeIndex...)
	ret, err := cl.CallView(blocklog.FuncGetLatestBlockInfo)
	if err != nil {
		return 0, err
	}
	n, _, err := codec.DecodeUint32(ret.MustGet(blocklog.ParamBlockIndex))
	return n, err
}

func (ch *Chain) ContractRegistry(committeeIndex ...int) (map[coretypes.Hname]*root.ContractRecord, error) {
	cl := ch.SCClient(root.Interface.Hname(), nil, committeeIndex...)
	ret, err := cl.CallView(root.FuncGetChainInfo)
	if err != nil {
		return nil, err
	}
	return root.DecodeContractRegistry(collections.NewMapReadOnly(ret, root.VarContractRegistry))
}

func (ch *Chain) GetCounterValue(inccounterSCHname coretypes.Hname, committeeIndex ...int) (int64, error) {
	cl := ch.SCClient(inccounterSCHname, nil, committeeIndex...)
	ret, err := cl.CallView(inccounter.FuncGetCounter)
	if err != nil {
		return 0, err
	}
	n, _, err := codec.DecodeInt64(ret.MustGet(inccounter.VarCounter))
	return n, err
}

func (ch *Chain) GetStateVariable(contractHname coretypes.Hname, key string, committeeIndex ...int) ([]byte, error) {
	cl := ch.SCClient(contractHname, nil, committeeIndex...)
	return cl.StateGet(key)
}
