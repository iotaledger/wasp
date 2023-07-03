package dss

import (
	"go.dedis.ch/kyber/v3"
	dkg "go.dedis.ch/kyber/v3/share/dkg/rabin"
)

//nolint:gocyclo
func GenDistSecret(suite dkg.Suite, nbParticipants int, partSec []kyber.Scalar, partPubs []kyber.Point) []*dkg.DistKeyShare {
	dkgs := make([]*dkg.DistKeyGenerator, nbParticipants)
	for i := 0; i < nbParticipants; i++ {
		dkg, err := dkg.NewDistKeyGenerator(suite, suite, partSec[i], partPubs, nbParticipants/2+1)
		if err != nil {
			panic(err)
		}
		dkgs[i] = dkg
	}
	// full secret sharing exchange
	// 1. broadcast deals
	resps := make([]*dkg.Response, 0, nbParticipants*nbParticipants)
	for _, dkg := range dkgs {
		deals, err := dkg.Deals()
		if err != nil {
			panic(err)
		}
		for i, d := range deals {
			resp, err := dkgs[i].ProcessDeal(d)
			if err != nil {
				panic(err)
			}
			if !resp.Response.Approved {
				panic("wrong approval")
			}
			resps = append(resps, resp)
		}
	}
	// 2. Broadcast responses
	for _, resp := range resps {
		for h, dkg := range dkgs {
			// ignore all messages from ourself
			if resp.Response.Index == uint32(h) {
				continue
			}
			j, err := dkg.ProcessResponse(resp)
			if err != nil || j != nil {
				panic("wrongProcessResponse")
			}
		}
	}
	// 4. Broadcast secret commitment
	for i, dkg := range dkgs {
		scs, err := dkg.SecretCommits()
		if err != nil {
			panic("wrong SecretCommits")
		}
		for j, dkg2 := range dkgs {
			if i == j {
				continue
			}
			cc, err := dkg2.ProcessSecretCommits(scs)
			if err != nil || cc != nil {
				panic("wrong ProcessSecretCommits")
			}
		}
	}

	// 5. reveal shares
	dkss := make([]*dkg.DistKeyShare, len(dkgs))
	for i, dkg := range dkgs {
		dks, err := dkg.DistKeyShare()
		if err != nil {
			panic(err)
		}
		dkss[i] = dks
	}
	return dkss
}

func GenPair(suite dkg.Suite) (kyber.Scalar, kyber.Point) {
	sc := suite.Scalar().Pick(suite.RandomStream())
	return sc, suite.Point().Mul(sc, nil)
}
