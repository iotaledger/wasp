package dashboard

import (
	"testing"

	"github.com/iotaledger/wasp/packages/webapi/testutil"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	e, d := mockDashboard()

	html := testutil.CallHTMLRequestHandler(t, e, d.handleConfig, "/", nil)

	t.Log(html)

	// match this:
	// <div class="card fluid">
	//	 <h2 class="section">Node configuration</h2>
	//	 <dl>
	//     <dt><tt>foo</tt></dt>
	//     <dd><tt>bar</tt></dd>
	//	 </dl>
	// </div>

	dt := html.Find("dl dt tt")
	require.Equal(t, 1, dt.Length())
	require.Equal(t, "foo", dt.First().Text())

	dd := html.Find("dl dd tt")
	require.Equal(t, 1, dd.Length())
	require.Equal(t, "bar", dd.First().Text())
}
