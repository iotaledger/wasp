// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package templates

type WaspConfigParams struct {
	APIPort                      int
	DashboardPort                int
	PeeringPort                  int
	NanomsgPort                  int
	L1APIAddress                 string
	ProfilingPort                int
	MetricsPort                  int
	OffledgerBroadcastUpToNPeers int
	OwnerAddress                 string
}

const WaspConfig = `
{
  "database": {
    "inMemory": true,
    "directory": "waspdb"
  },
  "logger": {
    "level": "info",
    "disableCaller": false,
    "disableStacktrace": true,
    "encoding": "console",
    "outputPaths": [
      "stdout",
      "wasp.log"
    ],
    "disableEvents": true
  },
  "network": {
    "bindAddress": "0.0.0.0",
    "externalAddress": "auto"
  },
  "node": {
    "disablePlugins": [],
    "enablePlugins": [],
    "ownerAddresses": ["{{.OwnerAddress}}"]
  },
  "webapi": {
    "auth": {
      "scheme": "none"
    },
    "bindAddress": "0.0.0.0:{{.APIPort}}"
  },
  "dashboard": {
    "auth": {
      "scheme": "none"
    },
    "bindAddress": "0.0.0.0:{{.DashboardPort}}"
  },
  "peering":{
    "port": {{.PeeringPort}},
    "netid": "127.0.0.1:{{.PeeringPort}}"
  },
  "L1APIAddress": "{{.L1APIAddress}}",
  "nanomsg":{
    "port": {{.NanomsgPort}}
  },
  "offledger":{
    "broadcastUpToNPeers": {{.OffledgerBroadcastUpToNPeers}}
  },
  "profiling":{
    "bindAddress": "0.0.0.0:{{.ProfilingPort}}",
    "writeProfiles": true,
    "enabled": false
  },
  "metrics": {
    "bindAddress": "0.0.0.0:{{.MetricsPort}}",
    "enabled": false
  },
  "wal": {
    "directory": "wal",
    "enabled": true
  },
  "debug": {
    "rawblocksEnabled": true,
    "rawblocksDirectory": "blocks"
  }
}
`
