package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/iotaledger/wasp/packages/authentication/shared"
)

// Login is a soon-to-be deprecated login function. It will be replaced by a generated client function.
func Login(client *http.Client, username, password string) (string, error) {
	loginRequest := shared.LoginRequest{
		Username: username,
		Password: password,
	}

	requestBytes, err := json.Marshal(loginRequest)

	if err != nil {
		return "", err
	}

	request, err := http.NewRequest(http.MethodPost, "/auth", bytes.NewBuffer(requestBytes))
	if err != nil {
		return "", err
	}

	result, err := client.Do(request)
	if err != nil {
		return "", err
	}

	resBody, err := io.ReadAll(result.Body)
	if err != nil {
		return "", err
	}

	var loginResponse shared.LoginResponse
	err = json.Unmarshal(resBody, &loginResponse)
	if err != nil {
		return "", err
	}

	return loginResponse.JWT, nil
}
