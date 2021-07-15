package dashboard

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/webapi/testutil"
	"github.com/stretchr/testify/require"
)

func checkProperConversionsToString(t *testing.T, html *goquery.Document) {
	// make sure we are using .Base58() instead of the default String() implementation
	// for things like OutputID, ChainID, Address, etc
	require.NotContains(t, strings.ToLower(html.Text()), "outputid {")
	require.NotContains(t, strings.ToLower(html.Text()), "{alias")
	require.NotContains(t, strings.ToLower(html.Text()), "address {")
	require.NotContains(t, strings.ToLower(html.Text()), "$/")
}

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
	require.Equal(t, 5, html.Find("table tbody tr").Length()) // 3 in peer list and 2 in trusted list.
}

func TestDashboardChainList(t *testing.T) {
	e, d := mockDashboard()
	html := testutil.CallHTMLRequestHandler(t, e, d.handleChainList, "/chains", nil)
	require.Equal(t, "mock chain", html.Find(`table tbody tr td[data-label="Description"]`).Text())
	checkProperConversionsToString(t, html)
}

func TestDashboardChainView(t *testing.T) {
	e, d := mockDashboard()
	html := testutil.CallHTMLRequestHandler(t, e, d.handleChain, "/chain/:chainid", map[string]string{
		"chainid": coretypes.RandomChainID().Base58(),
	})
	checkProperConversionsToString(t, html)
}

func TestDashboardChainAccount(t *testing.T) {
	e, d := mockDashboard()
	html := testutil.CallHTMLRequestHandler(t, e, d.handleChainAccount, "/chain/:chainid/account/:agentid", map[string]string{
		"chainid": coretypes.RandomChainID().Base58(),
		"agentid": strings.Replace(coretypes.NewRandomAgentID().String(), "/", ":", 1),
	})
	checkProperConversionsToString(t, html)
	require.Regexp(t, "^A/", html.Find(".value-agentid").Text())
}

func TestDashboardChainBlob(t *testing.T) {
	e, d := mockDashboard()
	html := testutil.CallHTMLRequestHandler(t, e, d.handleChainBlob, "/chain/:chainid/blob/:hash", map[string]string{
		"chainid": coretypes.RandomChainID().Base58(),
		"hash":    hashing.RandomHash(nil).Base58(),
	})
	checkProperConversionsToString(t, html)
}

func TestDashboardChainContract(t *testing.T) {
	e, d := mockDashboard()
	html := testutil.CallHTMLRequestHandler(t, e, d.handleChainContract, "/chain/:chainid/contract/:hname", map[string]string{
		"chainid": coretypes.RandomChainID().Base58(),
		"hname":   coretypes.Hname(0).String(),
	})
	checkProperConversionsToString(t, html)
}
