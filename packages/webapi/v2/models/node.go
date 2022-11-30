package models

type NodeOwnerCertificateRequest struct {
	NodePubKey   string `swagger:"desc(Node pub key. (base64))"`
	OwnerAddress string `swagger:"desc(Node owner address. (bech32))"`
}

type NodeOwnerCertificateResponse struct {
	Certificate string `swagger:"desc(Certificate stating the ownership. (base64))"`
}

type InfoResponse struct {
	Version   string `swagger:"desc(The version of the node)"`
	PublicKey string `swagger:"desc(The public key of the node)"`
	NetID     string `swagger:"desc(The net id of the node)"`
}
