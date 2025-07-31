package iotaclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

// We can set a certain amount of coin returned from the faucet. However,
// you get this value 5 times with 5 coins. Therefore,
// we need to assert account balances faucetamount * 5
const (
	SingleCoinFundsFromFaucetAmount = 2_000_000_000
	FundsFromFaucetAmount           = SingleCoinFundsFromFaucetAmount * 5
)

func RequestFundsFromFaucet(ctx context.Context, address *iotago.Address, faucetUrl string) error {
	paramJson := fmt.Sprintf(`{"FixedAmountRequest":{"recipient":"%v"}}`, address)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, faucetUrl, bytes.NewBuffer([]byte(paramJson)))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(request)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("post %v response code: %v", faucetUrl, res.Status)
	}
	defer res.Body.Close()

	resByte, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var response struct {
		Task  string `json:"task,omitempty"`
		Error string `json:"error,omitempty"`
	}
	err = json.Unmarshal(resByte, &response)
	if err != nil {
		return err
	}
	if strings.TrimSpace(response.Error) != "" {
		return errors.New(response.Error)
	}

	return nil
}
