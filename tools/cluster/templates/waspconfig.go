// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package templates

type ModifyNodesConfigFn = func(nodeIndex int, configParams WaspConfigParams) WaspConfigParams

type WaspConfigParams struct {
	APIPort                      int
	DashboardPort                int
	PeeringPort                  int
	NanomsgPort                  int
	L1INXAddress                 string
	ProfilingPort                int
	MetricsPort                  int
	OffledgerBroadcastUpToNPeers int
	OwnerAddress                 string
}

const WaspConfig = `
{
  "app": {
    "checkForUpdates": false,
    "shutdown": {
      "stopGracePeriod": "5m",
      "log": {
        "enabled": true,
        "filePath": "shutdown.log"
      }
    }
  },
  "logger": {
    "level": "debug",
    "disableCaller": false,
    "disableStacktrace": false,
    "encoding": "console",
    "outputPaths": [
      "wasp.log"
    ]
  },
  "inx": {
    "address": "{{.L1INXAddress}}",
    "maxConnectionAttempts": 30
  },
  "database": {
    "inMemory": false,
    "directory": "waspdb"
  },
  "registry": {
    "useText": false,
    "fileName": "chain-registry.json"
  },
  "peering": {
    "netID": "127.0.0.1:{{.PeeringPort}}",
    "port": {{.PeeringPort}}
  },
  "chains": {
    "broadcastUpToNPeers": 2,
    "broadcastInterval": "5s",
    "apiCacheTTL": "5m",
    "pullMissingRequestsFromCommittee": true
  },
  "rawBlocks": {
    "enabled": false,
    "directory": "blocks"
  },
  "profiling": {
    "bindAddress": "0.0.0.0:{{.ProfilingPort}}",
    "writeProfiles": false,
    "enabled": true
  },
  "wal": {
    "enabled": true,
    "directory": "wal"
  },
  "metrics": {
    "bindAddress": "0.0.0.0:{{.MetricsPort}}",
    "enabled": false
  },
  "webapi": {
    "enabled": true,
    "nodeOwnerAddresses": ["{{.OwnerAddress}}"],
    "bindAddress": "0.0.0.0:{{.APIPort}}",
    "auth": {
      "scheme": "none"
    }
  },
  "nanomsg": {
    "enabled": true,
    "port": {{.NanomsgPort}}
  },
  "dashboard": {
    "enabled": true,
    "bindAddress":  "0.0.0.0:{{.DashboardPort}}",
    "exploreAddressURL": "",
    "auth": {
      "scheme": "none"
    }
  },
  "users": {
    "users": {
      "wasp": {
        "Password": "wasp",
        "Permissions": [
          "dashboard",
          "api",
          "chain.read",
          "chain.write"
        ]
      }
    }
  },
  "offledger":{
    "broadcastUpToNPeers": {{.OffledgerBroadcastUpToNPeers}}
  },
  "debug": {
    "rawblocksEnabled": false,
    "rawblocksDirectory": "blocks"
  }
}
`
