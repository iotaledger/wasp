package models

// DKSharesPostRequest is a POST request for creating new DKShare.
type DKSharesPostRequest struct {
	PeerPubKeys []string `json:"peerPubKeys" swagger:"desc(Optional, base64 encoded public keys of the peers generating the DKS.)"`
	Threshold   uint16   `json:"threshold" swagger:"desc(Should be =< len(PeerPubKeys))"`
	TimeoutMS   uint32   `json:"timeoutMS" swagger:"desc(Timeout in milliseconds.)"`
}
