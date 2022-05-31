package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/authentication/shared"
)

func (c *WaspClient) Login(username, password string) (string, error) {
	loginRequest := shared.LoginRequest{
		Username: username,
		Password: password,
	}

	loginResponse := shared.LoginResponse{}

	err := c.do(http.MethodPost, shared.AuthRoute(), &loginRequest, &loginResponse)
	if err != nil {
		return "", err
	}

	return loginResponse.JWT, nil
}

func (c *WaspClient) AuthInfo() (*shared.AuthInfoModel, error) {
	authInfoResponse := shared.AuthInfoModel{}

	err := c.do(http.MethodGet, shared.AuthInfoRoute(), nil, &authInfoResponse)
	if err != nil {
		return nil, err
	}

	return &authInfoResponse, nil
}
