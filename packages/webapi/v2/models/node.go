package models

type NodeOwnerCertificateRequest struct {
	NodePubKey   string `swagger:"desc(Node pub key. (base64))"`
	OwnerAddress string `swagger:"desc(Node owner address. (bech32))"`
}

type NodeOwnerCertificateResponse struct {
	Certificate string `swagger:"desc(Certificate stating the ownership. (base64))"`
}
