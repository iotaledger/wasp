package dashboard

import (
	"html/template"
	"net/http"

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
	return makeTemplate(tplWs, tplInstallConfig, tplFairAuction)
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
				<details>
					<summary>{{$auction.Description}}</summary>
					<p>For sale: <code>{{$auction.NumTokens}}</code> tokens of color <code>{{$color}}</code></p>
					<p>Owner: <code>{{$auction.AuctionOwner}}</code></p>
					<p>Started at: <code>{{formatTimestamp $auction.WhenStarted}}</code></p>
					<p>Duration: <code>{{$auction.DurationMinutes}} minutes</code></p>
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
			<p><code>{{waspClientCmd}} wallet mint <i>amount-tokens</i></code>
			<br/>(e.g.: <code>{{waspClientCmd}} wallet mint 1</code>)</p>
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
	{{template "ws" .}}
{{end}}
`
