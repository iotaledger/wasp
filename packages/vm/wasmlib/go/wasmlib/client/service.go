package client

import (
	"github.com/mr-tron/base58"
)

type ServiceClient struct{}

type Response map[string][]byte

type Service struct{}

func (s *Service) Init(client ServiceClient, chainId string, scHname string, eventHandlers map[string]func([]string)) {
}

func (s *Service) CallView(viewName string, args map[string][]byte) Response {
	return nil
}

func (s *Service) PostRequest(hFuncName string, args map[string][]byte) {
}

/////////////////////////////////////////////////////////////////

func Base58Decode(s string) []byte {
	res, err := base58.Decode(s)
	if err != nil {
		panic("invalid base58 encoding")
	}
	return res
}

func Base58Encode(b []byte) string {
	return base58.Encode(b)
}
