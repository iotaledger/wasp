package dashboard

import (
	"testing"

	"github.com/iotaledger/wasp/packages/webapi/testutil"
	"github.com/stretchr/testify/require"
)

func TestDashboardPeering(t *testing.T) {
	e, d := mockDashboard()

	html := testutil.CallHTMLRequestHandler(t, e, d.handlePeering, "/peering", nil)

	// test that 3 peers are listed in the table

	tbody := html.Find("table tbody tr")
	require.Equal(t, 3, tbody.Length())
}
