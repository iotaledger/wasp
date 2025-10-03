package distsign

import (
	"fortio.org/safecast"
	"go.dedis.ch/kyber/v3"
	distkeygen "go.dedis.ch/kyber/v3/share/dkg/rabin"
)

//nolint:gocyclo
func GenDistSecret(suite distkeygen.Suite, nbParticipants int, partSec []kyber.Scalar, partPubs []kyber.Point) []*distkeygen.DistKeyShare {
	keyGenerators := make([]*distkeygen.DistKeyGenerator, nbParticipants)
	for i := 0; i < nbParticipants; i++ {
		keyGenerator, err := distkeygen.NewDistKeyGenerator(suite, suite, partSec[i], partPubs, nbParticipants/2+1)
		if err != nil {
			panic(err)
		}
		keyGenerators[i] = keyGenerator
	}
	// full secret sharing exchange
	// 1. broadcast deals
	resps := make([]*distkeygen.Response, 0, nbParticipants*nbParticipants)
	for _, keyGenerator := range keyGenerators {
		deals, err := keyGenerator.Deals()
		if err != nil {
			panic(err)
		}
		for i, d := range deals {
			resp, err := keyGenerators[i].ProcessDeal(d)
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
		for h, keyGenerator := range keyGenerators {
			// ignore all messages from ourself
			if resp.Response.Index == safecast.MustConvert[uint32](h) {
				continue
			}
			j, err := keyGenerator.ProcessResponse(resp)
			if err != nil || j != nil {
				panic("wrongProcessResponse")
			}
		}
	}
	// 4. Broadcast secret commitment
	for i, keyGenerator := range keyGenerators {
		scs, err := keyGenerator.SecretCommits()
		if err != nil {
			panic("wrong SecretCommits")
		}
		for j, keyGenerator2 := range keyGenerators {
			if i == j {
				continue
			}
			cc, err := keyGenerator2.ProcessSecretCommits(scs)
			if err != nil || cc != nil {
				panic("wrong ProcessSecretCommits")
			}
		}
	}

	// 5. reveal shares
	dkss := make([]*distkeygen.DistKeyShare, len(keyGenerators))
	for i, keyGenerator := range keyGenerators {
		dks, err := keyGenerator.DistKeyShare()
		if err != nil {
			panic(err)
		}
		dkss[i] = dks
	}
	return dkss
}

func GenPair(suite distkeygen.Suite) (kyber.Scalar, kyber.Point) {
	sc := suite.Scalar().Pick(suite.RandomStream())
	return sc, suite.Point().Mul(sc, nil)
}
