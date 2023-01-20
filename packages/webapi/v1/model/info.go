package model

type InfoResponse struct {
	Version       string `json:"version" swagger:"desc(Wasp version)"`
	NetworkID     string `json:"networkId" swagger:"desc('hostname:port'; uniquely identifies the node)"`
	PublisherPort int    `json:"publisherPort" swagger:"desc(Nanomsg port that exposes publisher messages)"`
}
