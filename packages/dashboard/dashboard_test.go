package dashboard

import (
	"github.com/iotaledger/wasp/packages/webapi/v1/testutil"
	"strconv"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func checkProperConversionsToString(t *testing.T, html *goquery.Document) {
	// make sure we are not using the default String() implementation
	// for things like OutputID, ChainID, Address, etc
	require.NotContains(t, strings.ToLower(html.Text()), "outputid {")
	require.NotContains(t, strings.ToLower(html.Text()), "{alias")
	require.NotContains(t, strings.ToLower(html.Text()), "address {")
	require.NotContains(t, strings.ToLower(html.Text()), "$/")
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
		"chainid": ch.ChainID.String(),
	})
	checkProperConversionsToString(t, html)
}

func TestDashboardChainAccount(t *testing.T) {
	env := initDashboardTest(t)
	ch := env.newChain()
	html := testutil.CallHTMLRequestHandler(t, env.echo, env.dashboard.handleChainAccount, "/chain/:chainid/account/:agentid", map[string]string{
		"chainid": ch.ChainID.String(),
		"agentid": isc.NewRandomAgentID().String(),
	})
	checkProperConversionsToString(t, html)
	require.Regexp(t, "@", html.Find(".value-agentid").Text())
}

func TestDashboardChainBlob(t *testing.T) {
	env := initDashboardTest(t)
	ch := env.newChain()
	html := testutil.CallHTMLRequestHandler(t, env.echo, env.dashboard.handleChainBlob, "/chain/:chainid/blob/:hash", map[string]string{
		"chainid": ch.ChainID.String(),
		"hash":    hashing.RandomHash(nil).Hex(),
	})
	checkProperConversionsToString(t, html)
}

func TestDashboardChainBlock(t *testing.T) {
	env := initDashboardTest(t)
	ch := env.newChain()

	for _, index := range []string{"0", "1"} {
		html := testutil.CallHTMLRequestHandler(t, env.echo, env.dashboard.handleChainBlock, "/chain/:chainid/block/:index", map[string]string{
			"chainid": ch.ChainID.String(),
			"index":   index,
		})
		checkProperConversionsToString(t, html)
	}

	// post a request that fails
	_, err := ch.PostRequestSync(solo.NewCallParams("", ""), nil)
	require.NotEmpty(t, err.Error())
	receipt := ch.LastReceipt()
	html := testutil.CallHTMLRequestHandler(t, env.echo, env.dashboard.handleChainBlock, "/chain/:chainid/block/:index", map[string]string{
		"chainid": ch.ChainID.String(),
		"index":   strconv.Itoa(int(receipt.BlockIndex)),
	})
	checkProperConversionsToString(t, html)
	t.Log(ch.ResolveVMError(receipt.Error).Error())
	require.Equal(t, ch.ResolveVMError(receipt.Error).Error(), html.Find("#receipt-error-0").Text())
}

func TestDashboardChainContract(t *testing.T) {
	env := initDashboardTest(t)
	ch := env.newChain()
	html := testutil.CallHTMLRequestHandler(t, env.echo, env.dashboard.handleChainContract, "/chain/:chainid/contract/:hname", map[string]string{
		"chainid": ch.ChainID.String(),
		"hname":   accounts.Contract.Hname().String(),
	})
	checkProperConversionsToString(t, html)
}
