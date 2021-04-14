package dashboard

import (
	"strings"
	"testing"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/webapi/testutil"
	"github.com/stretchr/testify/require"
)

func TestDashboardConfig(t *testing.T) {
	e, d := mockDashboard()

	html := testutil.CallHTMLRequestHandler(t, e, d.handleConfig, "/", nil)

	dt := html.Find("dl dt tt")
	require.Equal(t, 1, dt.Length())
	require.Equal(t, "foo", dt.First().Text())

	dd := html.Find("dl dd tt")
	require.Equal(t, 1, dd.Length())
	require.Equal(t, "bar", dd.First().Text())
}

func TestDashboardPeering(t *testing.T) {
	e, d := mockDashboard()
	html := testutil.CallHTMLRequestHandler(t, e, d.handlePeering, "/peering", nil)
	require.Equal(t, 3, html.Find("table tbody tr").Length())
}

func TestDashboardChainList(t *testing.T) {
	e, d := mockDashboard()
	html := testutil.CallHTMLRequestHandler(t, e, d.handleChainList, "/chains", nil)
	require.Equal(t, "mock chain", html.Find(`table tbody tr td[data-label="Description"]`).Text())
}

func TestDashboardChainView(t *testing.T) {
	e, d := mockDashboard()
	html := testutil.CallHTMLRequestHandler(t, e, d.handleChain, "/chain/:chainid", map[string]string{
		"chainid": coretypes.RandomChainID().Base58(),
	})

	// make sure we are using .Base58()
	require.NotContains(t, html.Text(), "OutputID {")
	require.NotContains(t, html.Text(), "Address {")
}

func TestDashboardChainAccount(t *testing.T) {
	e, d := mockDashboard()
	html := testutil.CallHTMLRequestHandler(t, e, d.handleChainAccount, "/chain/:chainid/account/:agentid", map[string]string{
		"chainid": coretypes.RandomChainID().Base58(),
		"agentid": strings.Replace(coretypes.NewRandomAgentID().String(), "/", ":", 1),
	})

	// make sure we are using .Base58()
	require.NotContains(t, html.Text(), "OutputID {")
	require.NotContains(t, html.Text(), "Address {")
}
