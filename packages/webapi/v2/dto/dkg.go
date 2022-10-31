package dto

// DKSharesInfo stands for the DKShare representation, returned by the GET and POST methods.
type DKSharesInfo struct {
	Address      string   `json:"address" swagger:"desc(New generated shared address.)"`
	SharedPubKey []byte   `json:"sharedPubKey" swagger:"desc(Shared public key.)"`
	PubKeyShares [][]byte `json:"pubKeyShares" swagger:"desc(Public key shares for all the peers.)"`
	PeerPubKeys  [][]byte `json:"peerPubKeys" swagger:"desc(Public keys of the nodes sharing the key.)"`
	Threshold    uint16   `json:"threshold"`
	PeerIndex    *uint16  `json:"peerIndex" swagger:"desc(Index of the node returning the share, if it is a member of the sharing group.)"`
}
