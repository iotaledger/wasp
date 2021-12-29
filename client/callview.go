package client

import (
	"net/http"
	"strings"
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

const (
	retryTimeoutOnOptimisticReadFail = 500 * time.Millisecond
	defaultOptimisticReadTimeout     = 1100 * time.Millisecond
)

func (c *WaspClient) CallView(chainID *iscp.ChainID, hContract iscp.Hname, functionName string, args dict.Dict, optimisticReadTimeout ...time.Duration) (dict.Dict, error) {
	deadline := time.Now().Add(defaultOptimisticReadTimeout)
	if len(optimisticReadTimeout) > 0 {
		deadline = time.Now().Add(optimisticReadTimeout[0])
	}
	arguments := args
	if arguments == nil {
		arguments = dict.Dict(nil)
	}
	var res dict.Dict
	var err error
	for {
		err = c.do(http.MethodGet, routes.CallView(chainID.Hex(), hContract.String(), functionName), arguments, &res)
		switch {
		case err == nil:
			return res, err
		case strings.Contains(err.Error(), "virtual state has been invalidated"):
			if time.Now().After(deadline) {
				return nil, coreutil.ErrorStateInvalidated
			}
			time.Sleep(retryTimeoutOnOptimisticReadFail)
		default:
			return nil, err
		}
	}
}
