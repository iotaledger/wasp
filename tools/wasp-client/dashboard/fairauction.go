package dashboard

import (
	"html/template"
	"net/http"
	"time"

	"github.com/iotaledger/wasp/packages/vm/examples/fairauction/faclient"
	"github.com/iotaledger/wasp/tools/wasp-client/config"
	"github.com/iotaledger/wasp/tools/wasp-client/config/fa"
	"github.com/labstack/echo"
)

func handleFA(c echo.Context) error {
	status, err := fa.Client().FetchStatus()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "fairauction", &FATemplateParams{
		BaseTemplateParams: baseParams(c, "fairauction"),
		SC:                 fa.Config,
		Status:             status,
	})
}

type FATemplateParams struct {
	BaseTemplateParams
	SC     *config.SCConfig
	Status *faclient.Status
}

func initFATemplate() *template.Template {
	t := template.New("").Funcs(template.FuncMap{
		"formatTimestamp": func(ts int64) string {
			return time.Unix(0, ts).UTC().Format(time.RFC3339)
		},
	})
	t = template.Must(t.Parse(tplBase))
	t = template.Must(t.Parse(tplWs))
	t = template.Must(t.Parse(tplInstallConfig))
	t = template.Must(t.Parse(tplFairAuction))
	return t
}

const tplFairAuction = `
{{define "title"}}FairAuction{{end}}

{{define "body"}}
	<p>SC address: <code>{{.SC.Address}}</code></p>
	<p>Balance: <code>{{.Status.SCBalance}} IOTAs</code></p>

	<div>
		<h2>Auctions</h2>
		<div>
			<ul>
			{{range $color, $auction := .Status.Auctions}}
				<li><div>
					<p>Color: <code>{{$color}}</code></p>
					<p>Owner: <code>{{$auction.AuctionOwner}}</code></p>
					<p>Description: <code>{{$auction.Description}}</code></p>
					<p>Started at: <code>{{formatTimestamp $auction.WhenStarted}}</code></p>
					<p>Duration: <code>{{$auction.DurationMinutes}} minutes</code></p>
					<p>Deposit: <code>{{$auction.TotalDeposit}}</code></p>
					<p>Tokens for sale: <code>{{$auction.NumTokens}}</code></p>
					<p>Minimum bid: <code>{{$auction.MinimumBid}} IOTAs</code></p>
					<p>Owner margin: <code>{{$auction.OwnerMargin}} promilles</code></p>
					<p>Bids:
						<ul>
						{{range $i, $bid := $auction.Bids}}
							<li><div>
								<p>Bidder: <code>{{$bid.Bidder}}</code></p>
								<p>Amount: <code>{{$bid.Total}} IOTAs</code></p>
							</div></li>
						{{end}}
						</ul>
					</p>
				</div></li>
			{{end}}
			</ul>
		</div>
	</div>
	<hr/>
	<p>Status fetched at: <code>{{.Status.FetchedAt}}</code></p>
	<div>
		<h2>CLI usage</h2>
		{{template "install-config" .}}
		<details>
			<summary>3. Mint a new color</summary>
			<p><code>wasp-client wallet mint <i>amount-tokens</i></code>
			(e.g.: <code>wasp-client wallet mint 1</code>)</p>
		</details>
		<details>
			<summary>4. Start an auction</summary>
			<p><code>wasp-client fa start-auction <i>description</i> <i>color</i> <i>amount-tokens</i> <i>minimum-bid</i> <i>duration</i></code>
			(e.g.: <code>wasp-client fa start-auction gHw2r... 1 100 10</code>)</p>
		</details>
		<details>
			<summary>5. Place a bid</summary>
			<p><code>wasp-client fa place-bid <i>color</i> <i>amount-iotas</i></code>
			(e.g.: <code>wasp-client fa place-bid gHw2r... 110</code>)</p>
		</details>
	</div>
	{{template "ws" .}}
{{end}}
`
