package models

// DKSharesPostRequest is a POST request for creating new DKShare.
type DKSharesPostRequest struct {
	PeerPublicKeys []string `json:"peerPublicKeys" swagger:"desc(Optional, Hex encoded public keys of the peers generating the DKS. (Hex))"`
	Threshold      uint16   `json:"threshold" swagger:"desc(Should be =< len(PeerPublicKeys))"`
	TimeoutMS      uint32   `json:"timeoutMS" swagger:"desc(Timeout in milliseconds.)"`
}

// DKSharesInfo stands for the DKShare representation, returned by the GET and POST methods.
type DKSharesInfo struct {
	Address         string   `json:"address" swagger:"desc(New generated shared address.)"`
	SharedPublicKey string   `json:"sharedPubKey" swagger:"desc(Shared public key. (Hex))"`
	PublicKeyShares []string `json:"pubKeyShares" swagger:"desc(Public key shares for all the peers. (Hex))"`
	PeerPublicKeys  []string `json:"peerPublicKeys" swagger:"desc(Public keys of the nodes sharing the key. (Hex))"`
	Threshold       uint16   `json:"threshold"`
	PeerIndex       *uint16  `json:"peerIndex" swagger:"desc(Index of the node returning the share, if it is a member of the sharing group.)"`
}
