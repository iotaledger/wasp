package dashboard

import (
	"fmt"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/parameters"
)

func args(args ...interface{}) []interface{} {
	return args
}

func hashref(hash hashing.HashValue) *hashing.HashValue {
	return &hash
}

func chainIDref(chID isc.ChainID) *isc.ChainID {
	return &chID
}

func assetID(aID []byte) []byte {
	return aID
}

func trim(max int, s string) string {
	if len(s) > max {
		s = s[:max] + "…"
	}
	// escape unprintable chars
	s = fmt.Sprintf("%q", s)
	// remove quotes
	return s[1 : len(s)-1]
}

func incUint32(n uint32) uint32 {
	return n + 1
}

func decUint32(n uint32) uint32 {
	return n - 1
}

func keyToString(k kv.Key) string {
	return string(k)
}

func bytesToString(b []byte) string {
	return string(b)
}

func anythingToString(i interface{}) string {
	if i == nil {
		return ""
	}
	return fmt.Sprintf("%v", i)
}

func formatTimestamp(ts interface{}) string {
	t, ok := ts.(time.Time)
	if !ok {
		t = time.Unix(0, ts.(int64))
	}
	return t.UTC().Format(time.RFC3339)
}

func formatTimestampOrNever(t time.Time) string {
	timestampNever := time.Time{}
	if t == timestampNever {
		return "NEVER"
	}
	return formatTimestamp(t)
}

func (d *Dashboard) exploreAddressURL(address iotago.Address) string {
	return d.wasp.ExploreAddressBaseURL() + "/" + d.addressToString(address)
}

func (d *Dashboard) addressToString(a iotago.Address) string {
	return a.Bech32(parameters.L1.Protocol.Bech32HRP)
}

func (d *Dashboard) agentIDToString(a isc.AgentID) string {
	return a.String()
}

func (d *Dashboard) getETHAddress(a isc.AgentID) string {
	if !d.isETHAddress(a) {
		return ""
	}

	ethAgent, _ := a.(*isc.EthereumAddressAgentID)

	return ethAgent.EthAddress().String()
}

func (d *Dashboard) isETHAddress(a isc.AgentID) bool {
	_, ok := a.(*isc.EthereumAddressAgentID)

	return ok
}

func (d *Dashboard) isValidAddress(a isc.AgentID) bool {
	addr := d.addressFromAgentID(a)

	if addr != nil {
		return true
	}

	return d.isETHAddress(a)
}

func (d *Dashboard) addressFromAgentID(a isc.AgentID) iotago.Address {
	addr, _ := isc.AddressFromAgentID(a)
	return addr
}
