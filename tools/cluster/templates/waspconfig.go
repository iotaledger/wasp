// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package templates

import "github.com/iotaledger/wasp/packages/cryptolib"

type ModifyNodesConfigFn = func(nodeIndex int, configParams WaspConfigParams) WaspConfigParams

type WaspConfigParams struct {
	APIPort                      int
	PeeringPort                  int
	L1INXAddress                 string
	ProfilingPort                int
	MetricsPort                  int
	OffledgerBroadcastUpToNPeers int // TODO this is unused, should it be removed?
	ValidatorKeyPair             *cryptolib.KeyPair
	ValidatorAddress             string // bech32 encoded address of ValidatorKeyPair
	PruningMinStatesToKeep       int
	AuthScheme                   string
}

var WaspConfig = `
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
    "encodingConfig": {
      "timeEncoder": "rfc3339nano"
    },
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
    "debugSkipHealthCheck": true
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
    "peeringURL": "localhost:{{.PeeringPort}}",
    "port": {{.PeeringPort}}
  },
  "chains": {
    "broadcastUpToNPeers": 2,
    "broadcastInterval": "5s",
    "apiCacheTTL": "5m",
    "pullMissingRequestsFromCommittee": true,
    "deriveAliasOutputByQuorum": true,
    "pipeliningLimit": -1,
    "consensusDelay": "50ms"
  },
  "stateManager": {
    "blockCacheMaxSize": 1000,
    "blockCacheBlocksInCacheDuration": "1h",
    "blockCacheBlockCleaningPeriod": "1m",
    "stateManagerGetBlockRetry": "3s",
    "stateManagerRequestCleaningPeriod": "1s",
    "stateManagerTimerTickPeriod": "1s",
    "pruningMinStatesToKeep": {{.PruningMinStatesToKeep}},
    "pruningMaxStatesToDelete": 1000
  },
  "validator": {
    "address": "{{.ValidatorAddress}}"
  },
  "wal": {
    "enabled": true,
    "path": "waspdb/wal"
  },
  "webapi": {
    "enabled": true,
    "bindAddress": "0.0.0.0:{{.APIPort}}",
    "auth": {
      "scheme": "{{.AuthScheme}}",
      "jwt": {
        "duration": "24h"
      }
    },
    "limits": {
      "timeout": "30s",
      "readTimeout": "10s",
      "writeTimeout": "1m",
      "maxBodyLength": "2M",
      "maxTopicSubscriptionsPerClient": 0,
      "confirmedStateLagThreshold": 2
    },
    "debugRequestLoggerEnabled": false
  },
  "profiling": {
    "enabled": false,
    "bindAddress": "0.0.0.0:{{.ProfilingPort}}" 
  },
  "profilingRecorder": {
    "enabled": false
  },
  "prometheus": {
    "enabled": true,
    "bindAddress": "0.0.0.0:{{.MetricsPort}}"
  }
}
`
