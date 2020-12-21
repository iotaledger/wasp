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
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/labstack/echo"
)

const chainRoute = "/chains/:chainid"
const chainTplName = "chain"

func addChainEndpoints(e *echo.Echo) {
	e.GET(chainRoute, func(c echo.Context) error {
		chainid, err := coretypes.NewChainIDFromBase58(c.Param("chainid"))
		if err != nil {
			return err
		}

		result := &ChainTemplateParams{
			BaseTemplateParams: BaseParams(c, chainListRoute),
			ChainID:            chainid,
		}

		result.ChainRecord, err = registry.GetChainRecord(&chainid)
		if err != nil {
			return err
		}

		if result.ChainRecord != nil && result.ChainRecord.Active {
			_, result.Block, _, err = state.LoadSolidState(&chainid)
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

		return c.Render(http.StatusOK, chainTplName, result)
	})
}

func fetchAccounts(chain chain.Chain) ([]coretypes.AgentID, error) {
	accounts, err := callView(chain, accountsc.Interface.Hname(), accountsc.FuncAccounts, nil)
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
	bal, err := callView(chain, accountsc.Interface.Hname(), accountsc.FuncTotalAssets, nil)
	if err != nil {
		return nil, err
	}
	return accountsc.DecodeBalances(bal)
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

	ChainRecord *registry.ChainRecord
	Block       state.Block
	RootInfo    RootInfo
	Accounts    []coretypes.AgentID
	TotalAssets map[balance.Color]int64
	Blobs       map[hashing.HashValue]uint32
	Committee   struct {
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
<div class="container">
<div class="row">
<div class="col-sm">
	{{ $chainid := .ChainID }}

	{{if .ChainRecord}}
		{{ $rootinfo := .RootInfo }}
		{{ $desc := printf "%.50s" $rootinfo.Description }}

		<h2>{{if $desc}}{{$desc}}{{else}}<tt>{{$chainid}}</tt>{{end}}</h2>

		<dl>
			<dt>ChainID</dt><dd><tt>{{.ChainRecord.ChainID}}</tt></dd>
			<dt>Chain address</dt><dd>{{template "address" .ChainRecord.ChainID.Address}}</dd>
			<dt>Chain color</dt><dd><tt>{{.ChainRecord.Color}}</tt></dd>
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
		{{if .ChainRecord.Active}}
			<div class="card fluid">
				<h3>Contracts</h3>
				{{range $_, $c := $rootinfo.Contracts}}
					<div class="card fluid">
						<h4><tt>{{printf "%.30s" $c.Name}}</tt></h4>
						<dl>
							<dt>Description</dt><dd><tt>{{printf "%.50s" $c.Description}}</tt></dd>
							<dt>Program hash</dt><dd><tt>{{$c.ProgramHash.String}}</tt></dd>
							{{if $c.HasCreator}}<dt>Creator</dt><dd>{{ template "agentid" (args $chainid $c.Creator) }}</dd>{{end}}
							<dt>Owner fee</dt><dd>
								{{- if $c.OwnerFee -}}
									<tt>{{- $c.OwnerFee }} {{ $rootinfo.FeeColor -}}</tt>
								{{- else -}}
									<tt>{{- $rootinfo.DefaultOwnerFee }} {{ $rootinfo.FeeColor }}</tt> (chain default)
								{{- end -}}
							</dd>
							<dt>Validator fee</dt><dd>
								{{- if $c.ValidatorFee -}}
									<tt>{{- $c.ValidatorFee }} {{ $rootinfo.FeeColor -}}
								{{- else -}}
									<tt>{{- $rootinfo.DefaultValidatorFee }} {{ $rootinfo.FeeColor }}</tt> (chain default)
								{{- end -}}
							</dd>
						</dl>
					</div>
				{{end}}
			</div>

			<div class="card fluid">
				<h3>On-chain accounts</h3>
				<table style="max-width: 50em">
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
				<h3>Blobs</h3>
				<table style="max-width: 50em">
					<thead>
						<tr>
							<th>Hash</th>
							<th>Size (bytes)</th>
						</tr>
					</thead>
					<tbody>
					{{range $hash, $size := .Blobs}}
						<tr>
							<td><a href="/chains/{{$chainid}}/blob/{{hashref $hash}}"><tt>{{ hashref $hash }}</tt></a></td>
							<td>{{ $size }}</td>
						</tr>
					{{end}}
					</tbody>
				</table>
			</div>

			<div class="card fluid">
				<h3>Block</h3>
				<dl>
				<dt>State index</dt><dd><tt>{{.Block.StateIndex}}</tt></dd>
				<dt>State Transaction ID</dt><dd><tt>{{.Block.StateTransactionID}}</tt></dd>
				<dt>Timestamp</dt><dd><tt>{{formatTimestamp .Block.Timestamp}}</tt></dd>
				<dt>Essence Hash</dt><dd><tt>{{.Block.EssenceHash}}</tt></dd>
				</dl>
				<div>
					<table style="max-width: 50em">
						<caption>Requests</caption>
						<thead>
							<tr>
								<th>RequestID</th>
							</tr>
						</thead>
						<tbody>
						{{range $_, $reqId := .Block.RequestIDs}}
							<tr>
								<td><tt>{{$reqId}}</tt></td>
							</tr>
						{{end}}
						</tbody>
					</table>
				</div>
			</div>

			<div class="card fluid">
				<h3>Committee</h3>
				<dl>
				<dt>Size</dt>      <dd><tt>{{.Committee.Size}}</tt></dd>
				<dt>Quorum</dt>    <dd><tt>{{.Committee.Quorum}}</tt></dd>
				<dt>NumPeers</dt>  <dd><tt>{{.Committee.NumPeers}}</tt></dd>
				<dt>HasQuorum</dt> <dd><tt>{{.Committee.HasQuorum}}</tt></dd>
				</dl>
				<table style="max-width: 50em">
				<caption>Peer status</caption>
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
	{{else}}
		<div class="card fluid error">No chain record for ID <td>{{$chainid}}</tt></div>
	{{end}}
</div>
</div>
</div>
{{end}}
`
