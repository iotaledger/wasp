// +build ignore

package fadashboard

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/vm/examples/fairauction/faclient"
	"github.com/iotaledger/wasp/tools/wasp-cli/dashboard"
	"github.com/iotaledger/wasp/tools/wasp-cli/sc"
	"github.com/iotaledger/wasp/tools/wasp-cli/sc/fa"
	"github.com/labstack/echo/v4"
)

type fadashboard struct{}

func Dashboard() dashboard.SCDashboard {
	return &fadashboard{}
}

func (d *fadashboard) Config() *sc.Config {
	return fa.Config
}

func (d *fadashboard) AddEndpoints(e *echo.Echo) {
	e.GET(fa.Config.Href(), handleFA)
}

func (d *fadashboard) AddTemplates(r dashboard.Renderer) {
	r[fa.Config.ShortName] = dashboard.MakeTemplate(
		dashboard.TplWs,
		dashboard.TplSCInfo,
		dashboard.TplInstallConfig,
		tplFairAuction,
	)
}

func handleFA(c echo.Context) error {
	status, err := fa.Client().FetchStatus()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, fa.Config.ShortName, &FATemplateParams{
		BaseTemplateParams: dashboard.BaseParams(c, fa.Config.Href()),
		Config:             fa.Config,
		Status:             status,
	})
}

type FATemplateParams struct {
	dashboard.BaseTemplateParams
	Config *sc.Config
	Status *faclient.Status
}

const tplFairAuction = `
{{define "title"}}{{.Config.Name}}{{end}}

{{define "body"}}
	<h2>{{.Config.Name}}</h1>
	{{template "sc-info" .}}

	<div>
		<h3>Auctions</h3>
		<div>
			{{range $color, $auction := .Status.Auctions}}
				<details>
                    <summary>For sale <code>{{$auction.NumTokens}}</code> token(s) of color <code>{{$color}}</code></summary>
					<p>Description: {{trim $auction.Description}}</p>
					<p>Lookup into <i>TokenRegistry</i>: <a href="/tr/{{$color}}"><code>{{$color}}</code></a></p>
					<p>Owner: {{template "address" $auction.AuctionOwner}}</p>
					<p>Started at: <code>{{formatTimestamp $auction.WhenStarted}}</code></p>
					<p>Duration: <code>{{$auction.DurationMinutes}} minutes</code></p>
					<p>Due: <code id="due-{{$color}}"></code></p>
					<p>Deposit: <code>{{$auction.TotalDeposit}}</code></p>
					<p>Minimum bid: <code>{{$auction.MinimumBid}} IOTAs</code></p>
					<p>Owner margin: <code>{{$auction.OwnerMargin}} promilles</code></p>
					{{if gt (len $auction.Bids) 0}}
						<p>This auction has <code>{{len $auction.Bids}}</code> bids totalling <code>{{$auction.SumOfBids}} IOTAs</code></p>
						{{$winner := $auction.WinningBid}}
						{{if $winner}}
							<p>Current winning bid: <code>{{$winner.Total}} IOTAs</code> by {{template "address" $winner.Bidder}}</p>
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
	<p>Status fetched at: <code>{{formatTimestamp .Status.FetchedAt}}</code></p>
	<div>
		<h3>CLI usage</h3>
		{{template "install-config" .}}
		<details>
			<summary>3. Mint a new color</summary>
			<p>See instructions in <a href="/tr">TokenRegistry</a>.</p>
		</details>
		<details>
			<summary>4. Start an auction</summary>
			<p><code>{{waspClientCmd}} fa start-auction <i>description</i> <i>color</i> <i>amount-tokens</i> <i>minimum-bid</i> <i>duration-in-minutes</i></code>
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
				var diff = (due - new Date())/1000;
				console.log(due, diff);
				if (diff > 0) {
                    var days = Math.floor(diff / 86400);
	                diff -= days * 86400;

                    var hours = Math.floor(diff / 3600) % 24;
                    diff -= hours * 3600;

                    var minutes = Math.floor(diff / 60) % 60;
                    diff -= minutes * 60;

                    var seconds = Math.round(diff % 60);  
                   
                    var disp = "in ";
                    if (days != 0){
                       disp += days.toString() + " days ";
                    }
                    if (hours != 0 || days != 0){
                       disp += hours.toString() + " hours ";
                    }
                    disp += minutes.toString() + " minutes ";
                    disp += seconds.toString() + " seconds";

                    countdown.innerText = disp;
					
                	//var date = new Date(0);
					//date.setSeconds(diff / 1000);
					// countdown.innerText = date.toISOString();
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
