package apilib

import (
	"math/rand"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"golang.org/x/xerrors"
)

// RunDKG runs DKG procedure on specifiec Wasp hosts. In case of success, generated address is returned
func RunDKG(apiHosts []string, peeringHosts []string, threshold uint16, timeout ...time.Duration) (ledgerstate.Address, error) {
	// TODO temporary. Correct type of timeout.
	to := uint32(60 * 1000)
	if len(timeout) > 0 {
		n := timeout[0].Milliseconds()
		if n < int64(util.MaxUint16) {
			to = uint32(n)
		}
	}
	dkgInitiatorIndex := rand.Intn(len(apiHosts))
	dkShares, err := client.NewWaspClient(apiHosts[dkgInitiatorIndex]).DKSharesPost(&model.DKSharesPostRequest{
		PeerNetIDs:  peeringHosts,
		PeerPubKeys: nil,
		Threshold:   threshold,
		TimeoutMS:   to, // 1 min
	})
	if err != nil {
		return nil, err
	}
	var ret ledgerstate.Address
	if ret, err = ledgerstate.AddressFromBase58EncodedString(dkShares.Address); err != nil {
		return nil, xerrors.Errorf("invalid address from DKG: %w", err)
	}
	return ret, nil
}
