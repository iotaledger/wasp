package wasmclient

import (
	"errors"
	"time"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/mr-tron/base58"
)

type ServiceClient struct {
	client  *client.WaspClient
	keyPair *ed25519.KeyPair
}

type Response map[string][]byte

type Service struct {
	chainID *iscp.ChainID
	cl      *ServiceClient
	scHname iscp.Hname
}

func (s *Service) Init(cl *ServiceClient, chainID string, scHname uint32, eventHandlers map[string]func([]string)) (err error) {
	s.scHname = iscp.Hname(scHname)
	s.chainID, err = iscp.ChainIDFromString(chainID)
	return err
}

func (s *Service) CallView(viewName string, args *Arguments) (ret Results) {
	ret.res, ret.err = s.cl.client.CallView(s.chainID, s.scHname, viewName, args.args)
	return ret
}

func (s *Service) PostRequest(funcHname uint32, args *Arguments, transfer ...map[string]uint64) Request {
	bal, err := makeBalances(transfer...)
	if err != nil {
		return Request{err: err}
	}
	reqArgs := requestargs.New().AddEncodeSimpleMany(args.args)
	req := request.NewOffLedger(s.chainID, s.scHname, iscp.Hname(funcHname), reqArgs)
	req.WithTransfer(bal)
	req.WithNonce(uint64(time.Now().UnixNano()))
	req.Sign(s.cl.keyPair)
	err = s.cl.client.PostOffLedgerRequest(s.chainID, req)
	if err != nil {
		return Request{err: err}
	}
	id := req.ID()
	return Request{id: &id}
}

func (s *Service) WaitRequest(req Request) error {
	return s.cl.client.WaitUntilRequestProcessed(s.chainID, *req.id, 1*time.Minute)
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

func makeBalances(transfer ...map[string]uint64) (colored.Balances, error) {
	cb := colored.NewBalances()
	switch len(transfer) {
	case 0:
	case 1:
		for color, amount := range transfer[0] {
			if color == colored.IOTA.String() {
				cb.Set(colored.IOTA, amount)
				continue
			}
			c, err := colored.ColorFromBase58EncodedString(color)
			if err != nil {
				return nil, err
			}
			cb.Set(c, amount)
		}
	default:
		return cb, errors.New("makeBalances: too many transfers")
	}
	return cb, nil
}
