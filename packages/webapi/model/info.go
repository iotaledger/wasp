package model

type InfoResponse struct {
	Version       string `swagger:"desc(Wasp version)"`
	NetworkId     string `swagger:"desc('hostname:port'; uniquely identifies the node)"`
	PublisherPort int    `swagger:"desc(Nanomsg port that exposes publisher messages)"`
}
