// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package privtangle

import (
	"encoding/hex"
	"fmt"
)

var configFileContentTemplate = `
{
	"restAPI": {
		"bindAddress": "0.0.0.0:14265",
		"jwtAuth": {
			"salt": "HORNET"
		},
		"publicRoutes": [
			"/health",
			"/mqtt",
			"/api/v2/*",
			"/api/plugins/*"
		],
		"protectedRoutes": [],
		"powEnabled": true,
		"powWorkerCount": 1,
		"limits": {
			"bodyLength": "1M",
			"maxResults": 1000
		}
	},
	"dashboard": {
		"bindAddress": "localhost:8081",
		"dev": false,
		"auth": {
			"sessionTimeout": "72h",
			"username": "admin",
			"passwordHash": "a5c5c6949e5259b6f74b08019da0b54b056473d2ed4712d8590682e6bd46876b",
			"passwordSalt": "b5769c198c45b84bf502ed0dde3b698eb885a527dca5bd5b0cd015992157cc79"
		}
	},
	"db": {
		"engine": "rocksdb",
		"path": "privatedb",
		"autoRevalidation": false
	},
	"snapshots": {
		"depth": 50,
		"interval": 200,
		"fullPath": "snapshots/full_snapshot.bin",
		"deltaPath": "snapshots/delta_snapshot.bin",
		"deltaSizeThresholdPercentage": 50.0,
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
			"thresholdPercentage": 10.0,
			"cooldownTime": "5m"
		},
		"pruneReceipts": false
	},
	"protocol": {
		"networkName": "private_tangle_wasp_cluster",
		"bech32HRP": "atoi",
		"minPoWScore": 100.0,
		"vByteCost": 0,
		"vByteFactorData": 1,
		"vByteFactorKey": 10,
		"milestonePublicKeyCount": 2,
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
	"pow": {
		"refreshTipsInterval": "5s"
	},
	"requests": {
		"discardOlderThan": "15s",
		"pendingReEnqueueInterval": "5s"
	},
	"coordinator": {
		"stateFilePath": "coordinator.state",
		"interval": "500ms",
		"powWorkerCount": 0,
		"checkpoints": {
			"maxTrackedMessages": 10000
		},
		"tipsel": {
			"minHeaviestBranchUnreferencedMessagesThreshold": 20,
			"maxHeaviestBranchTipsPerCheckpoint": 10,
			"randomTipsPerCheckpoint": 3,
			"heaviestBranchSelectionTimeout": "100ms"
		},
		"signing": {
			"provider": "local",
			"remoteAddress": "localhost:12345",
			"retryAmount": 10,
			"retryTimeout": "2s"
		},
		"quorum": {
			"enabled": false,
			"groups": {
				"hornet": [
					{
						"alias": "test01",
						"baseURL": "http://localhost:14265",
						"userName": "",
						"password": ""
					}
				]
			},
			"timeout": "2s"
		}
	},
	"migrator": {
		"stateFilePath": "migrator.state",
		"receiptMaxEntries": 110,
		"queryCooldownPeriod": "5s"
	},
	"receipts": {
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
	"tangle": {
		"milestoneTimeout": "30s"
	},
	"tipsel": {
		"maxDeltaMsgYoungestConeRootIndexToCMI": 2,
		"maxDeltaMsgOldestConeRootIndexToCMI": 7,
		"belowMaxDepth": 15,
		"nonLazy": {
			"retentionRulesTipsLimit": 100,
			"maxReferencedTipAge": "3s",
			"maxChildren": 30,
			"spammerTipsThreshold": 0
		},
		"semiLazy": {
			"retentionRulesTipsLimit": 20,
			"maxReferencedTipAge": "3s",
			"maxChildren": 2,
			"spammerTipsThreshold": 30
		}
	},
	"node": {
		"alias": "HORNET private-tangle node",
		"profile": "auto",
		"disablePlugins": [],
		"enablePlugins": [
			"Indexer",
			"Spammer",
			"Debug",
			"Prometheus"
		]
	},
	"p2p": {
		"bindMultiAddresses": [
			"/ip4/127.0.0.1/tcp/15600"
		],
		"connectionManager": {
			"highWatermark": 10,
			"lowWatermark": 5
		},
		"gossip": {
			"unknownPeersLimit": 4,
			"streamReadTimeout": "1m0s",
			"streamWriteTimeout": "10s"
		},
		"db": {
			"path": "p2pstore"
		},
		"reconnectInterval": "30s",
		"autopeering": {
			"bindAddress": "0.0.0.0:14626",
			"entryNodes": [],
			"entryNodesPreferIPv6": false,
			"runAsEntryNode": false
		}
	},
	"logger": {
		"level": "debug",
		"disableCaller": true,
		"encoding": "console",
		"outputPaths": [
			"hornet.log"
		]
	},
	"warpsync": {
		"advancementRange": 150
	},
	"spammer": {
		"message": "IOTA - A new dawn",
		"index": "HORNET Spammer",
		"indexSemiLazy": "HORNET Spammer Semi-Lazy",
		"cpuMaxUsage": 0.8,
		"mpsRateLimit": 5.0,
		"workers": 0,
		"autostart": true
	},
	"faucet": {
		"amount": 10000000,
		"smallAmount": 1000000,
		"maxAddressBalance": 20000000,
		"maxOutputCount": 127,
		"tagMessage": "HORNET FAUCET",
		"batchTimeout": "2s",
		"powWorkerCount": 0,
		"website": {
			"bindAddress": "localhost:8091",
			"enabled": true
		}
	},
	"mqtt": {
		"bindAddress": "localhost:1883",
		"wsPort": 1888,
		"workerCount": 100
	},
	"profiling": {
		"bindAddress": "localhost:6060"
	},
	"prometheus": {
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
		"migrationMetrics": true,
		"coordinatorMetrics": true,
		"mqttBrokerMetrics": true,
		"debugMetrics": false,
		"goMetrics": false,
		"processMetrics": false,
		"promhttpMetrics": false
	},
	"debug": {
		"whiteFlagParentsSolidTimeout": "2s"
	}
}
`

func (pt *PrivTangle) configFileContent() string {
	return fmt.Sprintf(configFileContentTemplate,
		hex.EncodeToString(pt.CooKeyPair1.GetPublicKey().AsBytes()),
		hex.EncodeToString(pt.CooKeyPair2.GetPublicKey().AsBytes()),
	)
}
