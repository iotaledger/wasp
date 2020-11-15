package dkg

import (
	"bytes"
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
)

// Coordinator runs at the requestor node and coordinates the
// entire DKG procedure.
type Coordinator struct {
	dkgID    string            // Unique ID of the DKG procedure instance.
	coordKey kyber.Scalar      // Private key of the corrdinator.
	coordPub kyber.Point       // Public key of the coordinator.
	peerLocs []string          // Addresses of the peers.
	peerPubs []kyber.Point     // Public keys of the peers participating in the DKG.
	network  CoordNodeProvider // Network to use for communication.
}

// GenerateDistributedKey is called from the client node to initiate the DKG
// procedure on a set of nodes. The client is not required to have an instance
// of the DkgNode, but may have it (be one of the peers sharing the secret).
// This function works synchronously, so the user should run it async if needed.
func GenerateDistributedKey(
	coordKey kyber.Scalar,
	coordPub kyber.Point,
	peerLocs []string,
	peerPubs []kyber.Point,
	timeout time.Duration,
	suite *edwards25519.SuiteEd25519,
	network CoordNodeProvider,
) (*Coordinator, error) {
	var err error
	var dkgID string = address.Random().String()
	c := Coordinator{
		dkgID:    dkgID,
		coordKey: coordKey,
		coordPub: coordPub,
		peerLocs: peerLocs,
		peerPubs: peerPubs,
		network:  network,
	}
	//
	// Initialize the peers.
	var peerPubsBytes [][]byte
	if peerPubsBytes, err = pubsToBytes(peerPubs); err != nil {
		return nil, err
	}
	var coordPubBytes []byte
	if coordPubBytes, err = pubToBytes(coordPub); err != nil {
		return nil, err
	}
	initReq := InitReq{
		PeerLocs:  peerLocs,
		PeerPubs:  peerPubsBytes,
		CoordPub:  coordPubBytes,
		TimeoutMS: timeout.Milliseconds(),
	}
	if err = c.network.DkgInit(peerLocs, dkgID, &initReq); err != nil {
		return nil, err
	}
	//
	// Perform the DKG steps, each step in parallel, all steps sequentially.
	// Step numbering (R) is according to <https://github.com/dedis/kyber/blob/master/share/dkg/rabin/dkg.go>.
	if err = c.network.DkgStep(peerLocs, dkgID, &StepReq{Step: "1-R2.1-SendDeals"}); err != nil {
		return nil, err
	}
	if err = c.network.DkgStep(peerLocs, dkgID, &StepReq{Step: "2-R2.2-SendResponses"}); err != nil {
		return nil, err
	}
	if err = c.network.DkgStep(peerLocs, dkgID, &StepReq{Step: "3-R2.3-SendJustifications"}); err != nil {
		return nil, err
	}
	if err = c.network.DkgStep(peerLocs, dkgID, &StepReq{Step: "4-R4-SendSecretCommits"}); err != nil {
		return nil, err
	}
	if err = c.network.DkgStep(peerLocs, dkgID, &StepReq{Step: "5-R5-SendComplaintCommits"}); err != nil {
		return nil, err
	}
	//
	// Now get the public keys.
	// This also performs the "6-R6-SendReconstructCommits" step implicitly.
	var pubKeyResponses []*PubKeyResp
	if pubKeyResponses, err = c.network.DkgPubKey(peerLocs, dkgID); err != nil {
		return nil, err
	}
	pubKeyBytes := pubKeyResponses[0].PubKey
	for i := range pubKeyResponses {
		if !bytes.Equal(pubKeyResponses[i].PubKey, pubKeyBytes) {
			return nil, fmt.Errorf("nodes generated different public keys")
		}
	}
	sharedPublic := suite.Point()
	sharedPublic.UnmarshalBinary(pubKeyBytes)
	fmt.Printf("COORD: Shared public key: %v\n", sharedPublic)
	//
	// Commit the keys to persistent storage.
	if err = c.network.DkgStep(peerLocs, dkgID, &StepReq{Step: "7-CommitAndTerminate"}); err != nil {
		return nil, err
	}
	return &c, nil
}
