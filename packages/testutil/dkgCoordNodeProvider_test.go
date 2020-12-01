package testutil_test

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDkgCoordNodeProvider(t *testing.T) {
	addrs := []string{"a", "b", "c"}
	pubKey := []byte("someKey")
	dkgId := "someDkgId"
	targetProviders := make([]dkg.CoordNodeProvider, len(addrs))
	dummyProviders := make([]*dummyCoordNodeProvider, len(addrs))
	for i := range targetProviders {
		dummyProviders[i] = &dummyCoordNodeProvider{pubKey: pubKey}
		targetProviders[i] = dummyProviders[i]
	}
	provider := testutil.NewDkgCoordNodeProvider(targetProviders, time.Second)
	require.Nil(t, provider.DkgInit(addrs, dkgId, &dkg.InitReq{}))
	require.Nil(t, provider.DkgStep(addrs, dkgId, &dkg.StepReq{}))
	pubKeys, err := provider.DkgPubKey(addrs, dkgId)
	assert.Nil(t, err)
	for i := range pubKeys {
		assert.Equal(t, pubKeys[i].SharedPublic, pubKey)
	}
	for i := range dummyProviders {
		assert.True(t, dummyProviders[i].init)
		assert.True(t, dummyProviders[i].step)
	}
}

type dummyCoordNodeProvider struct {
	init   bool
	step   bool
	pubKey []byte
}

func (p *dummyCoordNodeProvider) DkgInit(peerAddrs []string, dkgID string, msg *dkg.InitReq) error {
	p.init = true
	return nil
}
func (p *dummyCoordNodeProvider) DkgStep(peerAddrs []string, dkgID string, msg *dkg.StepReq) error {
	p.step = true
	return nil
}
func (p *dummyCoordNodeProvider) DkgPubKey(peerAddrs []string, dkgID string) ([]*dkg.PubKeyResp, error) {
	resp := make([]*dkg.PubKeyResp, len(peerAddrs))
	for i := range resp {
		resp[i] = &dkg.PubKeyResp{SharedPublic: p.pubKey}
	}
	return resp, nil
}
