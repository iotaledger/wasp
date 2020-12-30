package client

import (
	"net/http"
)

const (
	InfoRoute = "info"
)

type InfoResponse struct {
	Version       string `swagger:"desc(Wasp version)"`
	NetworkId     string `swagger:"desc('hostname:port'; uniquely identifies the node)"`
	PublisherPort int    `swagger:"desc(Nanomsg port that exposes publisher messages)"`
}

// Info gets the info of the node.
func (c *WaspClient) Info() (*InfoResponse, error) {
	res := &InfoResponse{}
	if err := c.do(http.MethodGet, InfoRoute, nil, res); err != nil {
		return nil, err
	}
	return res, nil
}
