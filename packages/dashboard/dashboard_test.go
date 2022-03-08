package dashboard

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
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
	// For colored.Color we need to use String() instead of .Base58() ¯\_(ツ)_/¯
	require.NotContains(t, strings.ToLower(html.Text()), "["+strings.Repeat("0 ", colored.ColorLength-1)+"0]")
}

func TestDashboardConfig(t *testing.T) {
	env := initDashboardTest(t)

	html := testutil.CallHTMLRequestHandler(t, env.echo, env.dashboard.handleConfig, "/config", nil)

	dt := html.Find("dl dt code")
	require.Equal(t, 1, dt.Length())
	require.Equal(t, "foo", dt.First().Text())

	dd := html.Find("dl dd code")
	require.Equal(t, 1, dd.Length())
	require.Equal(t, "bar", dd.First().Text())
}

func TestDashboardPeering(t *testing.T) {
	env := initDashboardTest(t)
	html := testutil.CallHTMLRequestHandler(t, env.echo, env.dashboard.handlePeering, "/peering", nil)
	require.Equal(t, 5, html.Find("table tbody tr").Length()) // 3 in peer list and 2 in trusted list.
}

func TestDashboardChainList(t *testing.T) {
	env := initDashboardTest(t)
	env.newChain()
	env.newChain()
	html := testutil.CallHTMLRequestHandler(t, env.echo, env.dashboard.handleChainList, "/chains", nil)
	require.Equal(t, 2, html.Find(`table tbody tr`).Length())
	checkProperConversionsToString(t, html)
}

func TestDashboardChainView(t *testing.T) {
	env := initDashboardTest(t)
	ch := env.newChain()
	html := testutil.CallHTMLRequestHandler(t, env.echo, env.dashboard.handleChain, "/chain/:chainid", map[string]string{
		"chainid": ch.ChainID.Base58(),
	})
	checkProperConversionsToString(t, html)
}

func TestDashboardChainAccount(t *testing.T) {
	env := initDashboardTest(t)
	ch := env.newChain()
	html := testutil.CallHTMLRequestHandler(t, env.echo, env.dashboard.handleChainAccount, "/chain/:chainid/account/:agentid", map[string]string{
		"chainid": ch.ChainID.Base58(),
		"agentid": strings.Replace(iscp.NewRandomAgentID().String(), "/", ":", 1),
	})
	checkProperConversionsToString(t, html)
	require.Regexp(t, "^A/", html.Find(".value-agentid").Text())
}

func TestDashboardChainBlob(t *testing.T) {
	env := initDashboardTest(t)
	ch := env.newChain()
	html := testutil.CallHTMLRequestHandler(t, env.echo, env.dashboard.handleChainBlob, "/chain/:chainid/blob/:hash", map[string]string{
		"chainid": ch.ChainID.Base58(),
		"hash":    hashing.RandomHash(nil).Base58(),
	})
	checkProperConversionsToString(t, html)
}

func TestDashboardChainBlock(t *testing.T) {
	env := initDashboardTest(t)
	ch := env.newChain()

	for _, index := range []string{"0", "1"} {
		html := testutil.CallHTMLRequestHandler(t, env.echo, env.dashboard.handleChainBlock, "/chain/:chainid/block/:index", map[string]string{
			"chainid": ch.ChainID.Base58(),
			"index":   index,
		})
		checkProperConversionsToString(t, html)
	}
}

func TestDashboardChainContract(t *testing.T) {
	env := initDashboardTest(t)
	ch := env.newChain()
	html := testutil.CallHTMLRequestHandler(t, env.echo, env.dashboard.handleChainContract, "/chain/:chainid/contract/:hname", map[string]string{
		"chainid": ch.ChainID.Base58(),
		"hname":   iscp.Hname(0).String(),
	})
	checkProperConversionsToString(t, html)
}
