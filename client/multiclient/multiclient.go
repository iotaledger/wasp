package multiclient

import (
	"net/http"
	"time"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/util/multicall"
)

// MultiClient allows to send webapi requests in parallel to multiple wasp nodes
type MultiClient struct {
	nodes []*client.WaspClient

	Timeout time.Duration
}

// New creates a new instance of MultiClient
func New(hosts []string, httpClient ...func() http.Client) *MultiClient {
	m := &MultiClient{
		nodes: make([]*client.WaspClient, len(hosts)),
	}
	for i, host := range hosts {
		if len(httpClient) > 0 {
			m.nodes[i] = client.NewWaspClient(host, httpClient[0]())
		} else {
			m.nodes[i] = client.NewWaspClient(host)
		}
	}
	m.Timeout = 30 * time.Second
	return m
}

func (m *MultiClient) WithToken(token string) *MultiClient {
	for _, v := range m.nodes {
		v.WithToken(token)
	}

	return m
}

func (m *MultiClient) WithLogFunc(logFunc func(msg string, args ...interface{})) *MultiClient {
	for i, node := range m.nodes {
		m.nodes[i] = node.WithLogFunc(logFunc)
	}
	return m
}

func (m *MultiClient) Len() int {
	return len(m.nodes)
}

// Do executes a callback once for each node in parallel, then wraps all error results into a single one
func (m *MultiClient) Do(f func(int, *client.WaspClient) error) error {
	return m.DoWithQuorum(f, len(m.nodes))
}

// Do executes a callback once for each node in parallel, then wraps all error results into a single one
func (m *MultiClient) DoWithQuorum(f func(int, *client.WaspClient) error, quorum int) error {
	funs := make([]func() error, len(m.nodes))
	for i := range m.nodes {
		j := i // duplicate variable for closure
		funs[j] = func() error { return f(j, m.nodes[j]) }
	}
	errs := multicall.MultiCall(funs, m.Timeout)
	return multicall.WrapErrorsWithQuorum(errs, quorum)
}
