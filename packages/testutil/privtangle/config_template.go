// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package privtangle

import (
	"encoding/hex"
	"fmt"
)

var configFileContentTemplate = `
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
    "level": "info",
    "disableCaller": true,
    "disableStacktrace": false,
    "stacktraceLevel": "panic",
    "encoding": "console",
    "outputPaths": [
      "hornet.log"
    ],
    "disableEvents": true
  },
  "node": {
    "profile": "auto",
    "alias": "HORNET private-tangle node"
  },
  "protocol": {
    "targetNetworkName": "private_tangle_wasp_cluster",
    "milestonePublicKeyCount": 2,
    "baseToken": {
      "name": "IOTA",
      "tickerSymbol": "IOTA",
      "unit": "IOTA",
      "subunit": "",
      "decimals": 0,
      "useMetricPrefix": false
    },
    "publicKeyRanges": [
      {
        "key": "%s",
        "start": 0,
        "end": 0
      },
      {
        "key": "%s",
        "start": 0,
        "end": 0
      }
    ]
  },
  "db": {
    "engine": "pebble",
    "path": "privatedb",
    "autoRevalidation": false,
    "checkLedgerStateOnStartup": false
  },
  "pow": {
    "refreshTipsInterval": "5s"
  },
  "p2p": {
    "bindMultiAddresses": [
      "/ip4/0.0.0.0/tcp/15600",
      "/ip6/::/tcp/15600"
    ],
    "connectionManager": {
      "highWatermark": 10,
      "lowWatermark": 5
    },
    "identityPrivateKey": "",
    "db": {
      "path": "p2pstore"
    },
    "reconnectInterval": "30s",
    "gossip": {
      "unknownPeersLimit": 4,
      "streamReadTimeout": "1m",
      "streamWriteTimeout": "10s"
    },
    "autopeering": {
      "enabled": false,
      "bindAddress": "0.0.0.0:14626",
      "entryNodes": [],
      "entryNodesPreferIPv6": false,
      "runAsEntryNode": false
    }
  },
  "requests": {
    "discardOlderThan": "15s",
    "pendingReEnqueueInterval": "5s"
  },
  "tangle": {
    "milestoneTimeout": "30s",
    "maxDeltaBlockYoungestConeRootIndexToCMI": 80,
    "maxDeltaBlockOldestConeRootIndexToCMI": 130,
    "whiteFlagParentsSolidTimeout": "2s"
  },
  "snapshots": {
    "enabled": false,
    "depth": 50,
    "interval": 200,
    "fullPath": "snapshots/full_snapshot.bin",
    "deltaPath": "snapshots/delta_snapshot.bin",
    "deltaSizeThresholdPercentage": 50,
    "deltaSizeThresholdMinSize": "50M",
    "downloadURLs": []
  },
  "pruning": {
    "milestones": {
      "enabled": false,
      "maxMilestonesToKeep": 60480
    },
    "size": {
      "enabled": true,
      "targetSize": "30GB",
      "thresholdPercentage": 10,
      "cooldownTime": "5m"
    },
    "pruneReceipts": false
  },
  "profiling": {
    "enabled": false,
    "bindAddress": "localhost:6060"
  },
  "restAPI": {
    "enabled": true,
    "bindAddress": "0.0.0.0:14265",
    "publicRoutes": [
      "/*"
    ],
    "protectedRoutes": [],
    "debugRequestLoggerEnabled": false,
    "jwtAuth": {
      "salt": "HORNET"
    },
    "pow": {
      "enabled": true,
      "workerCount": 1
    },
    "limits": {
      "maxBodyLength": "1M",
      "maxResults": 1000
    }
  },
  "warpsync": {
    "enabled": true,
    "advancementRange": 150
  },
  "tipsel": {
    "enabled": true,
    "nonLazy": {
      "retentionRulesTipsLimit": 100,
      "maxReferencedTipAge": "3s",
      "maxChildren": 30
    },
    "semiLazy": {
      "retentionRulesTipsLimit": 20,
      "maxReferencedTipAge": "3s",
      "maxChildren": 2
    }
  },
  "receipts": {
    "enabled": false,
    "backup": {
      "enabled": false,
      "path": "receipts"
    },
    "validator": {
      "validate": false,
      "ignoreSoftErrors": false,
      "api": {
        "address": "http://localhost:14266",
        "timeout": "5s"
      },
      "coordinator": {
        "address": "UDYXTZBE9GZGPM9SSQV9LTZNDLJIZMPUVVXYXFYVBLIEUHLSEWFTKZZLXYRHHWVQV9MNNX9KZC9D9UZWZ",
        "merkleTreeDepth": 24
      }
    }
  },
  "prometheus": {
    "enabled": false,
    "bindAddress": "localhost:9311",
    "fileServiceDiscovery": {
      "enabled": false,
      "path": "target.json",
      "target": "localhost:9311"
    },
    "databaseMetrics": true,
    "nodeMetrics": true,
    "gossipMetrics": true,
    "cachesMetrics": true,
    "restAPIMetrics": true,
    "inxMetrics": true,
    "migrationMetrics": true,
    "debugMetrics": false,
    "goMetrics": true,
    "processMetrics": true,
    "promhttpMetrics": true
  },
  "inx": {
    "enabled": true,
    "bindAddress": "localhost:9029",
    "pow": {
      "workerCount": 0
    }
  },
  "debug": {
    "enabled": false
  }
}
`

const protocolParameters = `
{
  "version": 2,
  "networkName": "private_tangle_wasp_cluster",
  "bech32HRP": "atoi",
  "minPoWScore": 100,
  "belowMaxDepth": 150,
  "rentStructure": {
    "vByteCost": 600,
    "vByteFactorData": 1,
    "vByteFactorKey": 10
  },
  "tokenSupply": "2779530283277761"
}
`

func (pt *PrivTangle) configFileContent() string {
	return fmt.Sprintf(configFileContentTemplate,
		hex.EncodeToString(pt.CooKeyPair1.GetPublicKey().AsBytes()),
		hex.EncodeToString(pt.CooKeyPair2.GetPublicKey().AsBytes()),
	)
}
