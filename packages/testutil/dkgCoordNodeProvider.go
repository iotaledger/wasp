package testutil

import (
	"errors"
	"time"

	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/util/multicall"
)

// DkgCoordNodeProvider is an implementation for the
// dkg.CoordNodeProvider interface for unit tests.
type DkgCoordNodeProvider struct {
	providers   []dkg.CoordNodeProvider
	callTimeout time.Duration
}

// NewDkgCoordNodeProvider creates new fake network provider.
func NewDkgCoordNodeProvider(providers []dkg.CoordNodeProvider, callTimeout time.Duration) *DkgCoordNodeProvider {
	return &DkgCoordNodeProvider{
		providers:   providers,
		callTimeout: callTimeout,
	}
}

// DkgInit implements dkg.CoordNodeProvider interface.
func (p *DkgCoordNodeProvider) DkgInit(peerAddrs []string, dkgID string, msg *dkg.InitReq) error {
	funs := make([]func() error, len(peerAddrs))
	for i := range peerAddrs {
		ii := i // A copy for the closure.
		funs[ii] = func() error {
			return p.providers[ii].DkgInit(peerAddrs[ii:ii+1], dkgID, msg)
		}
	}
	if ok, errs := multicall.MultiCall(funs, p.callTimeout); !ok {
		return multicall.WrapErrors(errs)
	}
	return nil
}

// DkgStep implements dkg.CoordNodeProvider interface.
func (p *DkgCoordNodeProvider) DkgStep(peerAddrs []string, dkgID string, msg *dkg.StepReq) error {
	funs := make([]func() error, len(peerAddrs))
	for i := range peerAddrs {
		ii := i // A copy for the closure.
		funs[ii] = func() error {
			return p.providers[ii].DkgStep(peerAddrs[ii:ii+1], dkgID, msg)
		}
	}
	if ok, errs := multicall.MultiCall(funs, p.callTimeout); !ok {
		return multicall.WrapErrors(errs)
	}
	return nil
}

// DkgPubKey implements dkg.CoordNodeProvider interface.
func (p *DkgCoordNodeProvider) DkgPubKey(peerAddrs []string, dkgID string) ([]*dkg.PubKeyResp, error) {
	funs := make([]func() error, len(peerAddrs))
	pubs := make([]*dkg.PubKeyResp, len(peerAddrs))
	for i := range peerAddrs {
		ii := i // A copy for the closure.
		funs[ii] = func() error {
			var err error
			var pub []*dkg.PubKeyResp
			if pub, err = p.providers[ii].DkgPubKey(peerAddrs[ii:ii+1], dkgID); err != nil {
				return err
			}
			if len(pub) != 1 {
				return errors.New("single_response_expected")
			}
			pubs[ii] = pub[0]
			return nil
		}
	}
	if ok, errs := multicall.MultiCall(funs, p.callTimeout); !ok {
		return nil, multicall.WrapErrors(errs)
	}
	return pubs, nil
}
