package fadashboard

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/vm/examples/fairauction/faclient"
	"github.com/iotaledger/wasp/tools/wwallet/dashboard"
	"github.com/iotaledger/wasp/tools/wwallet/sc"
	"github.com/iotaledger/wasp/tools/wwallet/sc/fa"
	"github.com/labstack/echo"
)

type fadashboard struct{}

func Dashboard() dashboard.SCDashboard {
	return &fadashboard{}
}

func (d *fadashboard) Config() *sc.Config {
	return fa.Config
}

const href = "/fairauction"

func (d *fadashboard) AddEndpoints(e *echo.Echo) {
	e.GET(href, handleFA)
}

func (d *fadashboard) AddTemplates(r dashboard.Renderer) {
	r["fairauction"] = dashboard.MakeTemplate(
		dashboard.TplWs,
		dashboard.TplSCInfo,
		dashboard.TplInstallConfig,
		tplFairAuction,
	)
}

func (d *fadashboard) AddNavPages(p []dashboard.NavPage) []dashboard.NavPage {
	return append(p, dashboard.NavPage{Title: "FairAuction", Href: href})
}

func handleFA(c echo.Context) error {
	status, err := fa.Client().FetchStatus()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "fairauction", &FATemplateParams{
		BaseTemplateParams: dashboard.BaseParams(c, href),
		SC:                 fa.Config,
		Status:             status,
	})
}

type FATemplateParams struct {
	dashboard.BaseTemplateParams
	SC     *sc.Config
	Status *faclient.Status
}

const tplFairAuction = `
{{define "title"}}FairAuction{{end}}

{{define "body"}}
	<h2>FairAuction</h2>
	{{template "sc-info" .}}

	<div>
		<h3>Auctions</h3>
		<div>
			{{range $color, $auction := .Status.Auctions}}
				<details>
					<summary>{{$auction.Description}}</summary>
					<p>For sale: <code>{{$auction.NumTokens}}</code> tokens of color <a href="/tokenregistry/{{$color}}"><code>{{$color}}</code></a></p>
					<p>Owner: <code>{{$auction.AuctionOwner}}</code></p>
					<p>Started at: <code>{{formatTimestamp $auction.WhenStarted}}</code></p>
					<p>Duration: <code>{{$auction.DurationMinutes}} minutes</code></p>
					<p>Due: <code id="due-{{$color}}"></code></p>
					<p>Deposit: <code>{{$auction.TotalDeposit}}</code></p>
					<p>Minimum bid: <code>{{$auction.MinimumBid}} IOTAs</code></p>
					<p>Owner margin: <code>{{$auction.OwnerMargin}} promilles</code></p>
					{{if gt (len $auction.Bids) 0}}
						<p>This auction has <code>{{len $auction.Bids}}</code> bids totalling <code>{{$auction.SumOfBids}} IOTAs</code></p>
						{{$winner := $auction.WinningBid}}
						{{if ne $winner nil}}
							<p>Current winning bid: <code>{{$winner.Total}} IOTAs</code> by <code>{{$winner.Bidder}}</code></p>
						{{end}}
					{{else}}
						<p>This auction has no bids yet.</p>
					{{end}}
				</details>
			{{else}}
				There are no active auctions.
			{{end}}
		</div>
	</div>
	<hr/>
	<p>Status fetched at: <code>{{.Status.FetchedAt}}</code></p>
	<div>
		<h3>CLI usage</h3>
		{{template "install-config" .}}
		<details>
			<summary>3. Mint a new color</summary>
			<p>See instructions in <a href="/tokenregistry">TokenRegistry</a>.</p>
		</details>
		<details>
			<summary>4. Start an auction</summary>
			<p><code>{{waspClientCmd}} fa start-auction <i>description</i> <i>color</i> <i>amount-tokens</i> <i>minimum-bid</i> <i>duration</i></code>
			<br/>(e.g.: <code>{{waspClientCmd}} fa start-auction "My awesome token" gHw2r... 1 100 10</code>)</p>
		</details>
		<details>
			<summary>5. Place a bid</summary>
			<p><code>{{waspClientCmd}} fa place-bid <i>color</i> <i>amount-iotas</i></code>
			<br/>(e.g.: <code>{{waspClientCmd}} fa place-bid gHw2r... 110</code>)</p>
		</details>
	</div>

	<script>
		function setupAuctionCountdown(color, due) {
			const countdown = document.getElementById("due-" + color);

			function update() {
				const diff = due - new Date();
				console.log(due, diff);
				if (diff > 0) {
					var date = new Date(0);
					date.setSeconds(diff / 1000);
					countdown.innerText = date.toISOString().substr(11, 8);
				} else {
					countdown.innerText = "";
				}
			}

			update()
			setInterval(update, 1000);
		}
		{{range $color, $auction := .Status.Auctions}}
			setupAuctionCountdown("{{$color}}", new Date({{formatTimestamp $auction.Due}}));
		{{end}}
	</script>

	{{template "ws" .}}
{{end}}
`
