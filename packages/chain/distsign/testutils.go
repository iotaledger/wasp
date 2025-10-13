package distsign

import (
	"fortio.org/safecast"
	"go.dedis.ch/kyber/v3"
	dkg "go.dedis.ch/kyber/v3/share/dkg/rabin"
)

//nolint:gocyclo
func GenDistSecret(suite dkg.Suite, nbParticipants int, partSec []kyber.Scalar, partPubs []kyber.Point) []*dkg.DistKeyShare {
	distKeyGenerators := make([]*dkg.DistKeyGenerator, nbParticipants)
	for i := 0; i < nbParticipants; i++ {
		generator, err := dkg.NewDistKeyGenerator(suite, suite, partSec[i], partPubs, nbParticipants/2+1)
		if err != nil {
			panic(err)
		}
		distKeyGenerators[i] = generator
	}
	// full secret sharing exchange
	// 1. broadcast deals
	resps := make([]*dkg.Response, 0, nbParticipants*nbParticipants)
	for _, generator := range distKeyGenerators {
		deals, err := generator.Deals()
		if err != nil {
			panic(err)
		}
		for i, d := range deals {
			resp, err := distKeyGenerators[i].ProcessDeal(d)
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
		for h, generator := range distKeyGenerators {
			// ignore all messages from ourself
			if resp.Response.Index == safecast.MustConvert[uint32](h) {
				continue
			}
			j, err := generator.ProcessResponse(resp)
			if err != nil || j != nil {
				panic("wrongProcessResponse")
			}
		}
	}
	// 4. Broadcast secret commitment
	for i, generator := range distKeyGenerators {
		scs, err := generator.SecretCommits()
		if err != nil {
			panic("wrong SecretCommits")
		}
		for j, generator2 := range distKeyGenerators {
			if i == j {
				continue
			}
			cc, err := generator2.ProcessSecretCommits(scs)
			if err != nil || cc != nil {
				panic("wrong ProcessSecretCommits")
			}
		}
	}

	// 5. reveal shares
	dkss := make([]*dkg.DistKeyShare, len(distKeyGenerators))
	for i, generator := range distKeyGenerators {
		dks, err := generator.DistKeyShare()
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
