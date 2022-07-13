// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package privtangle

import (
	"encoding/hex"
	"fmt"
)

var configFileContentTemplate = `
{
   "restAPI":{
      "bindAddress":"0.0.0.0:14265",
      "jwtAuth":{
         "salt":"HORNET"
      },
      "publicRoutes":[
         "/health",
         "/mqtt/v1",
         "/mqtt",
         "/api/info",
         "/api/routes",
         "/api/core/v2/*",
         "/api/v2/*",
         "/api/plugins/*",
         "/api/indexer/*",
         "/api/indexer/v1/*"
      ],
      "protectedRoutes":[
         
      ],
      "pow":{
         "enabled":true
      },
      "limits":{
         "bodyLength":"1M",
         "maxResults":1000
      }
   },
   "dashboard":{
      "bindAddress":"localhost:8081",
      "dev":false,
      "auth":{
         "sessionTimeout":"72h",
         "username":"admin",
         "passwordHash":"a5c5c6949e5259b6f74b08019da0b54b056473d2ed4712d8590682e6bd46876b",
         "passwordSalt":"b5769c198c45b84bf502ed0dde3b698eb885a527dca5bd5b0cd015992157cc79"
      }
   },
   "db":{
      "engine":"pebble",
      "path":"privatedb",
      "autoRevalidation":false
   },
   "snapshots":{
      "depth":50,
      "interval":200,
      "fullPath":"snapshots/full_snapshot.bin",
      "deltaPath":"snapshots/delta_snapshot.bin",
      "deltaSizeThresholdPercentage":50.0,
      "downloadURLs":[
         
      ]
   },
   "pruning":{
      "milestones":{
         "enabled":false,
         "maxMilestonesToKeep":60480
      },
      "size":{
         "enabled":true,
         "targetSize":"30GB",
         "thresholdPercentage":10.0,
         "cooldownTime":"5m"
      },
      "pruneReceipts":false
   },
   "protocol":{
      "targetNetworkName": "private_tangle_wasp_cluster",
      "baseToken":{
         "name":"Iota",
         "tickerSymbol":"MIOTA",
         "unit":"MIOTA",
         "decimals":6,
         "subunit":"IOTA",
         "useMetricPrefix":false
      },
      "milestonePublicKeyCount":2,
      "publicKeyRanges":[
         {
            "key":"%s",
            "start":0,
            "end":0
         },
         {
            "key":"%s",
            "start":0,
            "end":0
         }
      ]
   },
   "node":{
      "alias":"HORNET private-tangle node",
      "profile":"auto",
      "disablePlugins":[
         
      ],
      "enablePlugins":[
         "Indexer",
         "Spammer",
         "Debug",
         "Prometheus"
      ]
   },
   "p2p":{
      "bindMultiAddresses":[
         "/ip4/127.0.0.1/tcp/15600"
      ],
      "connectionManager":{
         "highWatermark":10,
         "lowWatermark":5
      },
      "gossip":{
         "unknownPeersLimit":4,
         "streamReadTimeout":"1m0s",
         "streamWriteTimeout":"10s"
      },
      "db":{
         "path":"p2pstore"
      },
      "reconnectInterval":"30s",
      "autopeering":{
         "bindAddress":"0.0.0.0:14626",
         "entryNodes":[
            
         ],
         "entryNodesPreferIPv6":false,
         "runAsEntryNode":false
      }
   },
   "logger":{
      "level":"info",
      "disableCaller":true,
      "encoding":"console",
      "outputPaths":[
         "hornet.log"
      ]
   },
   "spammer":{
      "message":"We are all made of stardust.",
      "tag":"HORNET Spammer",
      "tagSemiLazy":"HORNET Spammer Semi-Lazy",
      "cpuMaxUsage":0.8,
      "mpsRateLimit":0.0,
      "workers":0,
      "autostart":true
   },
   "mqtt":{
      "bindAddress":"localhost:1883",
      "wsPort":1888,
      "workerCount":100
   },
   "profiling":{
      "bindAddress":"localhost:6060"
   },
   "prometheus":{
      "bindAddress":"localhost:9311",
      "fileServiceDiscovery":{
         "enabled":false,
         "path":"target.json",
         "target":"localhost:9311"
      },
      "databaseMetrics":true,
      "nodeMetrics":true,
      "gossipMetrics":true,
      "cachesMetrics":true,
      "restAPIMetrics":true,
      "migrationMetrics":true,
      "coordinatorMetrics":true,
      "mqttBrokerMetrics":true,
      "debugMetrics":false,
      "goMetrics":false,
      "processMetrics":false,
      "promhttpMetrics":false
   },
   "debug":{
      "whiteFlagParentsSolidTimeout":"2s"
   }
}
`

const protocolParameters = `
{
   "version": 2,
   "networkName":"private_tangle_wasp_cluster",
   "bech32HRP":"atoi",
   "minPoWScore":1,
   "belowMaxDepth": 15,
   "rentStructure": {
       "vByteCost": 600,
       "vByteFactorData": 1,
       "vByteFactorKey": 10
   },
   "tokenSupply":"2779530283277761"
}
`

func (pt *PrivTangle) configFileContent() string {
	return fmt.Sprintf(configFileContentTemplate,
		hex.EncodeToString(pt.CooKeyPair1.GetPublicKey().AsBytes()),
		hex.EncodeToString(pt.CooKeyPair2.GetPublicKey().AsBytes()),
	)
}
