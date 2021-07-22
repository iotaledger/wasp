package tests

import (
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

type chainEnvBuilder struct {
	t *testing.T

	createCluster            bool
	clusterNumValidatorNodes int
	clusterSize              int

	createDKG bool

	createDeployChain bool
	chainName         string

	createDeployContract bool
	contractName         string
	contractProgHash     hashing.HashValue
	contractDescription  string
	contractInitParams   map[string]interface{}

	quorum      uint16
	committee   []int
	cluster     *cluster.Cluster
	chain       *cluster.Chain
	ledgerState *ledgerstate.Address
}

func CreateTestEnvironment(t *testing.T) *chainEnvBuilder {
	environment := chainEnvBuilder{
		t: t,
	}

	return &environment
}

func (b *chainEnvBuilder) buildCluster() error {
	if b.createCluster {
		b.committee = util.MakeRange(0, b.clusterNumValidatorNodes)
		b.cluster = newCluster(b.t, b.clusterSize)
	}

	return nil
}

func (b *chainEnvBuilder) buildDKG() error {
	if b.createCluster && b.createDKG {
		b.quorum = uint16((2*len(b.committee))/3 + 1)
		addr, err := b.cluster.RunDKG(b.committee, b.quorum)

		b.ledgerState = &addr

		return err
	}

	return nil
}

func (b *chainEnvBuilder) buildChainDeployment() error {
	if b.createCluster && b.createDKG && b.createDeployChain {
		chain, err := b.cluster.DeployChain(b.chainName, b.cluster.Config.AllNodes(), b.committee, b.quorum, *b.ledgerState)

		b.chain = chain

		return err
	}

	return nil
}

func (b *chainEnvBuilder) buildContractDeployment() error {
	if b.createCluster && b.createDKG && b.createDeployChain && b.createDeployContract {
		_, err := b.chain.DeployContract(b.contractName, b.contractProgHash.String(), b.contractDescription, b.contractInitParams)

		return err
	}

	return nil
}

func (b *chainEnvBuilder) Build() *chainEnv {
	err := b.buildCluster()
	require.NoError(b.t, err)
	err = b.buildDKG()
	require.NoError(b.t, err)
	err = b.buildChainDeployment()
	require.NoError(b.t, err)
	err = b.buildContractDeployment()
	require.NoError(b.t, err)

	e := &env{t: b.t, clu: b.cluster}

	chEnv := &chainEnv{
		env:   e,
		chain: b.chain,
	}

	return chEnv
}

func (b *chainEnvBuilder) WithCluster(numValidatorNodes int, clusterSize int) *chainEnvBuilder {
	b.createCluster = true
	b.clusterNumValidatorNodes = numValidatorNodes
	b.clusterSize = clusterSize

	return b
}

func (b *chainEnvBuilder) WithDKG() *chainEnvBuilder {
	b.createDKG = true

	return b
}

func (b *chainEnvBuilder) WithDeployChain(chainName string) *chainEnvBuilder {
	b.chainName = chainName

	return b
}

func (b *chainEnvBuilder) WithDeployContract(contractName string, progHash hashing.HashValue, description string, initParams map[string]interface{}) *chainEnvBuilder {
	b.contractName = contractName
	b.contractProgHash = progHash
	b.contractDescription = description
	b.contractInitParams = initParams

	return b
}
