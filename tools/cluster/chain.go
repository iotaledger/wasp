package cluster

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/client/wallet/packages/seed"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
)

type Chain struct {
	Description string

	OriginatorSeed *seed.Seed

	CommitteeNodes []int
	AccessNodes    []int
	Quorum         uint16
	Address        address.Address

	ChainID coretypes.ChainID
	Color   balance.Color

	Cluster *Cluster
}

func (ch *Chain) ContractID(contractHname coretypes.Hname) *coretypes.ContractID {
	cid := coretypes.NewContractID(ch.ChainID, contractHname)
	return &cid
}

func (ch *Chain) AllNodes() []int {
	r := make([]int, 0)
	r = append(r, ch.CommitteeNodes...)
	r = append(r, ch.AccessNodes...)
	return r
}

func (ch *Chain) CommitteeApi() []string {
	return ch.Cluster.WaspHosts(ch.CommitteeNodes, (*WaspNodeConfig).ApiHost)
}

func (ch *Chain) OriginatorAddress() *address.Address {
	addr := ch.OriginatorSeed.Address(0).Address
	return &addr
}

func (ch *Chain) OriginatorID() *coretypes.AgentID {
	ret := coretypes.NewAgentIDFromAddress(*ch.OriginatorAddress())
	return &ret
}

func (ch *Chain) OriginatorSigScheme() signaturescheme.SignatureScheme {
	return signaturescheme.ED25519(*ch.OriginatorSeed.KeyPair(0))
}

func (ch *Chain) OriginatorClient() *chainclient.Client {
	return ch.Client(ch.OriginatorSigScheme())
}

func (ch *Chain) Client(sigScheme signaturescheme.SignatureScheme) *chainclient.Client {
	return chainclient.New(
		ch.Cluster.NodeClient,
		ch.Cluster.WaspClient(ch.CommitteeNodes[0]),
		&ch.ChainID,
		sigScheme,
	)
}

func (ch *Chain) CommitteeMultiClient() *multiclient.MultiClient {
	return multiclient.New(ch.CommitteeApi())
}

func (ch *Chain) WithSCState(hname coretypes.Hname, f func(host string, blockIndex uint32, state codec.ImmutableMustCodec) bool) bool {
	pass := true
	for i, host := range ch.CommitteeApi() {
		if !ch.Cluster.Config.Nodes[i].IsUp() {
			continue
		}
		contractID := coretypes.NewContractID(ch.ChainID, hname)
		actual, err := ch.Cluster.WaspClient(i).DumpSCState(&contractID)
		if client.IsNotFound(err) {
			pass = false
			fmt.Printf("   FAIL: state does not exist\n")
			continue
		}
		if err != nil {
			panic(err)
		}
		if !f(host, actual.Index, codec.NewMustCodec(actual.Variables)) {
			pass = false
		}
	}
	return pass
}

func (ch *Chain) DeployBuiltinContract(name string, vmtype string, progHashStr string, description string, initParams map[string]interface{}) (*sctransaction.Transaction, error) {
	programHash, err := hashing.HashValueFromBase58(progHashStr)
	if err != nil {
		return nil, err
	}

	params := map[string]interface{}{
		root.ParamName:          name,
		root.ParamVMType:        vmtype,
		root.ParamDescription:   description,
		root.ParamProgramBinary: programHash[:],
	}
	for k, v := range initParams {
		params[k] = v
	}
	tx, err := ch.OriginatorClient().PostRequest(root.Hname, root.EntryPointDeployContract, nil, nil, params)
	if err != nil {
		return nil, err
	}

	err = ch.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (ch *Chain) DeployExternalContract(vmtype string, name string, description string, progBinary []byte, initParams map[string]interface{}) (*sctransaction.Transaction, error) {
	params := map[string]interface{}{
		root.ParamVMType:        vmtype,
		root.ParamName:          name,
		root.ParamDescription:   description,
		root.ParamProgramBinary: progBinary,
	}
	for k, v := range initParams {
		params[k] = v
	}
	tx, err := ch.OriginatorClient().PostRequest(root.Hname, root.EntryPointDeployContract, nil, nil, params)
	if err != nil {
		return nil, err
	}

	err = ch.CommitteeMultiClient().WaitUntilAllRequestsProcessed(tx, 30*time.Second)
	if err != nil {
		return nil, err
	}

	return tx, nil
}
