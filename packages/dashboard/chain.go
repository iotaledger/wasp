package dashboard

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/labstack/echo/v4"
)

func chainBreadcrumb(e *echo.Echo, chainID coretypes.ChainID) Tab {
	return Tab{
		Path:  e.Reverse("chain"),
		Title: fmt.Sprintf("Chain %.8s…", chainID),
		Href:  e.Reverse("chain", chainID.String()),
	}
}

func initChain(e *echo.Echo, r renderer) {
	route := e.GET("/chain/:chainid", handleChain)
	route.Name = "chain"
	r[route.Path] = makeTemplate(e, tplChain, tplWs)
}

func handleChain(c echo.Context) error {
	chainid, err := coretypes.NewChainIDFromBase58(c.Param("chainid"))
	if err != nil {
		return err
	}

	tab := chainBreadcrumb(c.Echo(), chainid)

	result := &ChainTemplateParams{
		BaseTemplateParams: BaseParams(c, tab),
		ChainID:            chainid,
	}

	result.ChainRecord, err = registry.GetChainRecord(&chainid)
	if err != nil {
		return err
	}

	if result.ChainRecord != nil && result.ChainRecord.Active {
		result.VirtualState, result.Block, _, err = state.LoadSolidState(&chainid)
		if err != nil {
			return err
		}

		chain := chains.GetChain(chainid)

		result.Committee.Size = chain.Size()
		result.Committee.Quorum = chain.Quorum()
		result.Committee.NumPeers = chain.NumPeers()
		result.Committee.HasQuorum = chain.HasQuorum()
		result.Committee.PeerStatus = chain.PeerStatus()
		result.RootInfo, err = fetchRootInfo(chain)
		if err != nil {
			return err
		}

		result.Accounts, err = fetchAccounts(chain)
		if err != nil {
			return err
		}

		result.TotalAssets, err = fetchTotalAssets(chain)
		if err != nil {
			return err
		}

		result.Blobs, err = fetchBlobs(chain)
		if err != nil {
			return err
		}
	}

	return c.Render(http.StatusOK, c.Path(), result)
}

func fetchAccounts(chain chain.Chain) ([]coretypes.AgentID, error) {
	accounts, err := callView(chain, accounts.Interface.Hname(), accounts.FuncAccounts, nil)
	if err != nil {
		return nil, fmt.Errorf("accountsc view call failed: %v", err)
	}

	ret := make([]coretypes.AgentID, 0)
	for k := range accounts {
		agentid, _, err := codec.DecodeAgentID([]byte(k))
		if err != nil {
			return nil, err
		}
		ret = append(ret, agentid)
	}
	return ret, nil
}

func fetchTotalAssets(chain chain.Chain) (map[balance.Color]int64, error) {
	bal, err := callView(chain, accounts.Interface.Hname(), accounts.FuncTotalAssets, nil)
	if err != nil {
		return nil, err
	}
	return accounts.DecodeBalances(bal)
}

func fetchBlobs(chain chain.Chain) (map[hashing.HashValue]uint32, error) {
	ret, err := callView(chain, blob.Interface.Hname(), blob.FuncListBlobs, nil)
	if err != nil {
		return nil, err
	}
	return blob.DecodeDirectory(ret)
}

type ChainTemplateParams struct {
	BaseTemplateParams

	ChainID coretypes.ChainID

	ChainRecord  *registry.ChainRecord
	Block        state.Block
	VirtualState state.VirtualState
	RootInfo     RootInfo
	Accounts     []coretypes.AgentID
	TotalAssets  map[balance.Color]int64
	Blobs        map[hashing.HashValue]uint32
	Committee    struct {
		Size       uint16
		Quorum     uint16
		NumPeers   uint16
		HasQuorum  bool
		PeerStatus []*chain.PeerStatus
	}
}

const tplChain = `
{{define "title"}}Chain details{{end}}

{{define "body"}}
	{{ $chainid := .ChainID }}

	{{if .ChainRecord}}
		{{ $rootinfo := .RootInfo }}
		{{ $desc := trim 50 $rootinfo.Description }}

		<div class="card fluid">
			<h2 class="section">{{if $desc}}{{$desc}}{{else}}Chain <tt>{{$chainid}}</tt>{{end}}</h2>

			<dl>
				<dt>ChainID</dt><dd><tt>{{.ChainRecord.ChainID}}</tt></dd>
				<dt>Chain address</dt><dd>{{template "address" .RootInfo.ChainAddress}}</dd>
				<dt>Chain color</dt><dd><tt>{{.RootInfo.ChainColor}}</tt></dd>
				<dt>Active</dt><dd><tt>{{.ChainRecord.Active}}</tt></dd>
				{{if .ChainRecord.Active}}
					<dt>Owner ID</dt><dd>{{template "agentid" (args .ChainID $rootinfo.OwnerID)}}</dd>
					<dt>Delegated Owner ID</dt><dd>
						{{- if $rootinfo.OwnerIDDelegated -}}
							{{- template "agentid" (args .ChainID $rootinfo.OwnerIDDelegated) -}}
						{{- end -}}
					</dd>
					<dt>Default owner fee</dt><dd><tt>{{$rootinfo.DefaultOwnerFee}} {{$rootinfo.FeeColor}}</tt></dd>
					<dt>Default validator fee</dt><dd><tt>{{$rootinfo.DefaultValidatorFee}} {{$rootinfo.FeeColor}}</tt></dd>
				{{end}}
			</dl>
		</div>

		{{if .ChainRecord.Active}}
			<div class="card fluid">
				<h3 class="section">Contracts</h3>
				<dl>
				{{range $_, $c := $rootinfo.Contracts}}
					<dt><a href="{{ uri "chainContract" $chainid $c.Hname }}"><tt>{{trim 30 $c.Name}}</tt></a></dt>
					<dd><tt>{{trim 50 $c.Description}}</tt></dd>
				{{end}}
				</dl>
			</div>

			<div class="card fluid">
				<h3 class="section">On-chain accounts</h3>
				<table>
					<thead>
						<tr>
							<th>AgentID</th>
						</tr>
					</thead>
					<tbody>
					{{range $_, $agentid := .Accounts}}
						<tr>
							<td>{{template "agentid" (args $chainid $agentid)}}</td>
						</tr>
					{{end}}
					</tbody>
				</table>
				<h4>Total assets</h4>
				{{ template "balances" .TotalAssets }}
			</div>

			<div class="card fluid">
				<h3 class="section">Blobs</h3>
				<table>
					<thead>
						<tr>
							<th style="flex: 2">Hashing</th>
							<th>Size (bytes)</th>
						</tr>
					</thead>
					<tbody>
					{{range $hash, $size := .Blobs}}
						<tr>
							<td style="flex: 2"><a href="{{ uri "chainBlob" $chainid (hashref $hash) }}"><tt>{{ hashref $hash }}</tt></a></td>
							<td>{{ $size }}</td>
						</tr>
					{{end}}
					</tbody>
				</table>
			</div>

			<div class="card fluid">
				<h3 class="section">State</h3>
				<dl>
					<dt>State index</dt><dd><tt>{{.Block.StateIndex}}</tt></dd>
					<dt>State hash</dt><dd><tt>{{.VirtualState.Hashing}}</tt></dd>
					<dt>Last updated</dt><dd><tt>{{formatTimestamp .Block.Timestamp}}</tt> in transaction <tt>{{.Block.StateTransactionID}}</tt></dd>
				</dl>
			</div>

			<div class="card fluid">
				<h3 class="section">Committee</h3>
				<dl>
				<dt>Size</dt>      <dd><tt>{{.Committee.Size}}</tt></dd>
				<dt>Quorum</dt>    <dd><tt>{{.Committee.Quorum}}</tt></dd>
				<dt>NumPeers</dt>  <dd><tt>{{.Committee.NumPeers}}</tt></dd>
				<dt>HasQuorum</dt> <dd><tt>{{.Committee.HasQuorum}}</tt></dd>
				</dl>
				<h4>Peer status</h4>
				<table>
				<thead>
					<tr>
						<th>Index</th>
						<th>ID</th>
						<th>Status</th>
					</tr>
				</thead>
				<tbody>
				{{range $_, $s := .Committee.PeerStatus}}
					<tr>
						<td>{{$s.Index}}</td>
						<td><tt>{{$s.PeeringID}}</tt></td>
						<td>{{if $s.Connected}}up{{else}}down{{end}}</td>
					</tr>
				{{end}}
				</tbody>
				</table>
			</div>
		{{end}}
		{{ template "ws" .ChainID }}
	{{else}}
		<div class="card fluid error">No chain record for ID <td>{{$chainid}}</tt></div>
	{{end}}
{{end}}
`
