variable "wasp_config" {
  default = <<EOH
{
  "app": {
    "checkForUpdates": true,
    "shutdown": {
      "stopGracePeriod": "5m",
      "log": {
        "enabled": true,
        "filePath": "{{ env "NOMAD_TASK_DIR" }}/waspdb/shutdown.log"
      }
    }
  },
  "logger": {
    "level": "debug",
    "disableCaller": false,
    "disableStacktrace": false,
    "stacktraceLevel": "panic",
    "encoding": "console",
    "outputPaths": [
      "stdout",
      "{{ env "NOMAD_TASK_DIR" }}/waspdb/wasp.log"
    ],
    "disableEvents": true
  },
  "inx": {
    "address": "{{ range service "inx.tangle-testnet-hornet" }}{{ .Address }}:{{ .Port }}{{ end }}",
    "maxConnectionAttempts": 30,
    "targetNetworkName": ""
  },
  "db": {
    "engine": "rocksdb",
    "chainState": {
      "path": "{{ env "NOMAD_TASK_DIR" }}/waspdb/chains/data"
    },
    "debugSkipHealthCheck": false
  },
  "p2p": {
    "identity": {
      "privateKey": "",
      "filePath": "{{ env "NOMAD_TASK_DIR" }}/waspdb/identity/identity.key"
    },
    "db": {
      "path": "{{ env "NOMAD_TASK_DIR" }}/waspdb/p2pstore"
    }
  },
  "registries": {
    "chains": {
      "filePath": "{{ env "NOMAD_TASK_DIR" }}/waspdb/chains/chain_registry.json"
    },
    "dkShares": {
      "path": "{{ env "NOMAD_TASK_DIR" }}/waspdb/dkshares"
    },
    "trustedPeers": {
      "filePath": "{{ env "NOMAD_TASK_DIR" }}/waspdb/trusted_peers.json"
    },
    "consensusState": {
      "path": "{{ env "NOMAD_TASK_DIR" }}/waspdb/chains/consensus"
    }
  },
  "peering": {
    "peeringURL": "{{ env "NOMAD_ADDR_peering" }}",
    "port": {{ env "NOMAD_PORT_peering" }}
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
    "enabled": false,
    "bindAddress": "{{ env "NOMAD_ADDR_profiling" }}"
  },
  "prometheus": {
    "enabled": true,
    "bindAddress": "0.0.0.0:{{ env "NOMAD_PORT_metrics" }}",
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
    "nodeOwnerAddresses": [],
    "bindAddress": "0.0.0.0:{{ env "NOMAD_PORT_api" }}",
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
        "whitelist": ${adminWhitelist}
      }
    }
  },
  "nanomsg": {
    "enabled": true,
    "port": {{ env "NOMAD_PORT_nanomsg" }}
  },
  "dashboard": {
    "enabled": true,
    "bindAddress": "0.0.0.0:{{ env "NOMAD_PORT_dashboard" }}"
    "exploreAddressURL": "",
    "debugRequestLoggerEnabled": false,
    "auth": {
      "scheme": "basic",
      "jwt": {
        "duration": "24h"
      },
      "basic": {
        "username": "wasp"
      },
      "ip": {
        "whitelist": ${adminWhitelist}
      }
    }
  }
}
EOH
}

job "isc-${workspace}" {
  datacenters = ["hcloud"]

  // update {
  //   max_parallel      = 1
  //   health_check      = "task_states"
  //   min_healthy_time  = "1s"
  //   healthy_deadline  = "30s"
  //   progress_deadline = "5m"
  //   auto_revert       = true
  //   auto_promote      = true
  //   canary            = 1
  //   stagger           = "15s"
  // }

  group "node" {
    ephemeral_disk {
      migrate = true
      sticky  = true
    }

    count = 0

    network {
      mode = "host"

      port "dashboard" {
        host_network = "private"
      }
      port "api" {
        host_network = "private"
      }
      port "nanomsg" {
        host_network = "private"
      }
      port "peering" {
        host_network = "private"
      }
      port "metrics" {
        host_network = "private"
      }
      port "profiling" {
        host_network = "private"
      }
    }

    task "wasp" {
      driver = "docker"

      config {
        network_mode = "host"
        image        = "${artifact.image}:${artifact.tag}"
        entrypoint   = ["wasp", "-c", "/local/config.json"]
        ports = [
          "dashboard",
          "api",
          "nanomsg",
          "peering",
          "metrics",
          "profiling"
        ]

        labels = {
          "co.elastic.metrics/raw" = "[{\"metricsets\":[\"collector\"],\"module\":\"prometheus\",\"period\":\"10s\",\"metrics_path\":\"/metrics\",\"hosts\":[\"$${NOMAD_ADDR_metrics}\"]}]"
          "wasp"                   = "node"
        }

        // logging {
        //   type = "gelf"
        //   config {
        //     gelf-address          = "tcp://elastic-logstash-beats-logstash.service.consul:12201"
        //     tag                   = "wasp"
        //     labels                = "wasp"
        //   }
        // }

        auth {
          username       = "${auth.username}"
          password       = "${auth.password}"
          server_address = "${auth.server_address}"
        }
      }

      service {
        tags = ["wasp", "api"]
        port = "api"
        check {
          type     = "http"
          port     = "api"
          path     = "info"
          interval = "5s"
          timeout  = "2s"
        }
      }
      service {
        tags = ["wasp", "dashboard"]
        port = "dashboard"
      }
      service {
        tags = ["wasp", "nanomsg"]
        port = "nanomsg"
      }
      service {
        tags = ["wasp", "peering"]
        port = "peering"
      }
      service {
        tags = ["wasp", "metrics"]
        port = "metrics"
      }

      template {
        data        = var.wasp_config
        destination = "/local/config.json"
        perms       = "777"
      }

      resources {
        memory = 4000
        cpu    = 3000
      }
    }
  }

  group "access" {
    ephemeral_disk {
      migrate = true
      sticky  = true
    }

    count = 4

    network {
      mode = "host"

      port "dashboard" {
        host_network = "private"
      }
      port "api" {
        host_network = "private"
      }
      port "nanomsg" {
        host_network = "private"
      }
      port "peering" {
        host_network = "private"
      }
      port "metrics" {
        host_network = "private"
      }
      port "profiling" {
        host_network = "private"
      }
      port "dlv" {
        static = 40000
        to = 40000
      }
    }

    task "wasp" {
      driver = "docker"

      config {
       network_mode = "host"
        image        = "${artifact.image}:${artifact.tag}"
        entrypoint   = ["wasp", "-c", "/local/config.json"]
        ports = [
          "dashboard",
          "api",
          "nanomsg",
          "peering",
          "metrics",
          "profiling"
        ]


        labels = {
          "co.elastic.metrics/raw" = "[{\"metricsets\":[\"collector\"],\"module\":\"prometheus\",\"period\":\"10s\",\"metrics_path\":\"/metrics\",\"hosts\":[\"$${NOMAD_ADDR_metrics}\"]}]"
          "wasp"                   = "access"
        }

        // logging {
        //   type = "gelf"
        //   config {
        //     gelf-address          = "tcp://elastic-logstash-beats-logstash.service.consul:12201"
        //     tag                   = "wasp"
        //     labels                = "wasp"
        //   }
        // }

        auth {
          username       = "${auth.username}"
          password       = "${auth.password}"
          server_address = "${auth.server_address}"
        }
      }

      service {
        tags = ["wasp", "api"]
        port = "api"
        check {
          type     = "http"
          port     = "api"
          path     = "info"
          interval = "5s"
          timeout  = "2s"
        }
      }
      service {
        tags = ["wasp", "dashboard"]
        port = "dashboard"
      }
      service {
        tags = ["wasp", "nanomsg"]
        port = "nanomsg"
      }
      service {
        tags = ["wasp", "peering"]
        port = "peering"
      }
      service {
        tags = ["wasp", "metrics"]
        port = "metrics"
      }
      service {
        tags = ["wasp", "profiling"]
        port = "profiling"
      }
      service {
        tags = ["wasp", "dlv"]
        port = "dlv"
      }

      template {
        data        = var.wasp_config
        destination = "/local/config.json"
        perms       = "777"
      }

      resources {
        memory = 4000
        cpu    = 3000
      }
    }
  }
}
