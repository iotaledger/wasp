package models

// DKSharesPostRequest is a POST request for creating new DKShare.
type DKSharesPostRequest struct {
	PeerIdentities []string `json:"peerIdentities" swagger:"desc(Optional, Hex encoded public keys of the peers generating the DKS. (Hex))"`
	Threshold      uint16   `json:"threshold" swagger:"desc(Should be =< len(PeerPublicIdentities))"`
	TimeoutMS      uint32   `json:"timeoutMS" swagger:"desc(Timeout in milliseconds.)"`
}

// DKSharesInfo stands for the DKShare representation, returned by the GET and POST methods.
type DKSharesInfo struct {
	Address         string   `json:"address" swagger:"desc(New generated shared address.)"`
	PeerIdentities  []string `json:"peerIdentities" swagger:"desc(Identities of the nodes sharing the key. (Hex))"`
	PeerIndex       *uint16  `json:"peerIndex" swagger:"desc(Index of the node returning the share, if it is a member of the sharing group.)"`
	PublicKey       string   `json:"publicKey" swagger:"desc(Used public key. (Hex))"`
	PublicKeyShares []string `json:"publicKeyShares" swagger:"desc(Public key shares for all the peers. (Hex))"`
	Threshold       uint16   `json:"threshold"`
}
