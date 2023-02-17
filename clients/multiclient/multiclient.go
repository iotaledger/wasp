package multiclient

import (
	"time"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/util/multicall"
)

type ClientResolver func(apiHost string) *apiclient.APIClient

// MultiClient allows to send webapi requests in parallel to multiple wasp nodes
type MultiClient struct {
	nodes []*apiclient.APIClient

	Timeout time.Duration
}

// New creates a new instance of MultiClient
func New(resolver ClientResolver, hosts []string) *MultiClient {
	m := &MultiClient{
		nodes: make([]*apiclient.APIClient, len(hosts)),
	}

	for i, host := range hosts {
		m.nodes[i] = resolver(host)
	}

	m.Timeout = 30 * time.Second
	return m
}

func (m *MultiClient) Len() int {
	return len(m.nodes)
}

// Do executes a callback once for each node in parallel, then wraps all error results into a single one
func (m *MultiClient) Do(f func(int, *apiclient.APIClient) error) error {
	return m.DoWithQuorum(f, len(m.nodes))
}

// Do executes a callback once for each node in parallel, then wraps all error results into a single one
func (m *MultiClient) DoWithQuorum(f func(int, *apiclient.APIClient) error, quorum int) error {
	funs := make([]func() error, len(m.nodes))
	for i := range m.nodes {
		j := i // duplicate variable for closure
		funs[j] = func() error { return f(j, m.nodes[j]) }
	}
	errs := multicall.MultiCall(funs, m.Timeout)
	return multicall.WrapErrorsWithQuorum(errs, quorum)
}
