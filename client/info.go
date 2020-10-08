package client

import (
	"net/http"
)

const (
	InfoRoute = "info"
)

type InfoResponse struct {
	Version       string
	NetworkId     string
	PublisherPort int
}

// Info gets the info of the node.
func (c *WaspClient) Info() (*InfoResponse, error) {
	res := &InfoResponse{}
	if err := c.do(http.MethodGet, InfoRoute, nil, res); err != nil {
		return nil, err
	}
	return res, nil
}
