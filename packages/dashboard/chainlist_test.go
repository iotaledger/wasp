package dashboard

import (
	"testing"

	"github.com/iotaledger/wasp/packages/webapi/testutil"
	"github.com/stretchr/testify/require"
)

func TestDashboardChainList(t *testing.T) {
	e, d := mockDashboard()

	html := testutil.CallHTMLRequestHandler(t, e, d.handleChainList, "/chains", nil)

	// test that a chain is listed in the table

	tbody := html.Find(`table tbody tr td[data-label="Description"]`)
	require.Equal(t, "mock chain", tbody.Text())
}
