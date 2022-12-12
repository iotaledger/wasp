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
    "checkForUpdates": true,
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
    "disableCaller": true,
    "disableStacktrace": false,
    "stacktraceLevel": "panic",
    "encoding": "console",
    "outputPaths": [
      "stdout",
      "wasp.log"
    ],
    "disableEvents": true
  },
  "inx": {
    "address": "{{.L1INXAddress}}",
    "maxConnectionAttempts": 30,
    "targetNetworkName": ""
  },
  "db": {
    "engine": "rocksdb",
    "chainState": {
      "path": "waspdb/chains/data"
    },
    "debugSkipHealthCheck": false
  },
  "p2p": {
    "identity": {
      "privateKey": "",
      "filePath": "waspdb/identity/identity.key"
    },
    "db": {
      "path": "waspdb/p2pstore"
    }
  },
  "registries": {
    "chains": {
      "filePath": "waspdb/chains/chain_registry.json"
    },
    "dkShares": {
      "path": "waspdb/dkshares"
    },
    "trustedPeers": {
      "filePath": "waspdb/trusted_peers.json"
    },
    "consensusState": {
      "path": "waspdb/chains/consensus"
    }
  },
  "peering": {
    "netID": "0.0.0.0:{{.PeeringPort}}",
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
    "enabled": true,
    "bindAddress": "0.0.0.0:{{.ProfilingPort}}"
  },
  "prometheus": {
    "enabled": true,
    "bindAddress": "0.0.0.0:{{.MetricsPort}}",
    "nodeMetrics": true,
    "nodeConnMetrics": true,
    "blockWALMetrics": true,
    "restAPIMetrics": true,
    "goMetrics": true,
    "processMetrics": true,
    "promhttpMetrics": true
  },
  "webapi": {
    "enabled": true,
    "nodeOwnerAddresses": ["{{.OwnerAddress}}"],
    "bindAddress": "0.0.0.0:{{.APIPort}}",
    "debugRequestLoggerEnabled": false,
    "auth": {
      "scheme": "none",
      "jwt": {
        "duration": "24h"
      },
      "basic": {
        "username": "wasp"
      },
      "ip": {
        "whitelist": [
          "0.0.0.0"
        ]
      }
    }
  },
  "nanomsg": {
    "enabled": true,
    "port": {{.NanomsgPort}}
  },
  "dashboard": {
    "enabled": true,
    "bindAddress": "0.0.0.0:{{.DashboardPort}}",
    "exploreAddressURL": "",
    "debugRequestLoggerEnabled": false,
    "auth": {
      "scheme": "none",
      "jwt": {
        "duration": "24h"
      },
      "basic": {
        "username": "wasp"
      },
      "ip": {
        "whitelist": [
          "0.0.0.0"
        ]
      }
    }
  }
}`
