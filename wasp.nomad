job "wasp" {
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

    count = 1

    network {
      mode = "host"

      port "dashboard" {
        to = 80
      }
      port "api" {
      }
      port "nanomsg" {
      }
      port "peering" {
      }
      port "metrics" {
      }
      port "profiling" {
      }
    }

    service {
      name = JOB
      tags = ["${NOMAD_NAMESPACE}", "api"]
      port = "api"
      check {
        type     = "http"
        port     = "api"
        path     = "/health"
        interval = "5s"
        timeout  = "2s"
      }

      # connect {
      #   sidecar_service {
      #     proxy {
      #       upstreams {
      #         destination_name = "elastic-agent-gelf"
      #         local_bind_port  = 8080
      #       }
      #     }
      #   }
      # }
    }
    service {
      name = JOB
      tags = ["${NOMAD_NAMESPACE}", "dashboard"]
      port = "dashboard"
    }
    service {
      name = JOB
      tags = ["${NOMAD_NAMESPACE}", "nanomsg"]
      port = "nanomsg"
    }
    service {
      name = JOB
      tags = ["${NOMAD_NAMESPACE}", "peering"]
      port = "peering"
    }
    service {
      name = JOB
      tags = ["${NOMAD_NAMESPACE}", "metrics"]
      port = "metrics"
    }

    task "dashboard" {
      driver = "docker"

      lifecycle {
        hook    = "poststart"
        sidecar = true
      }

      template {
        data        = <<EOH
WASP_API_URL="https://api.sc.testnet.shimmer.network"
L1_API_URL="https://api.hornet.sc.testnet.shimmer.network"
EOH
        destination = "${NOMAD_TASK_DIR}/dashboard.env"
        change_mode = "restart"
        env         = true
      }


      config {
        image = "iotaledger/wasp-dashboard:latest"
        ports = ["dashboard"]
      }
    }

    task "wasp" {
      driver = "docker"

      config {
        image      = "iotaledger/wasp:${VERSION}"
        entrypoint = ["/app/wasp", "-c", "${NOMAD_TASK_DIR}/config.json"]
        ports = [
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

        # logging {
        #   type = "gelf"
        #   config {
        #     gelf-address = "udp://${NOMAD_UPSTREAM_ADDR_elastic_agent_gelf}"
        #     # tag          = "wasp"
        #     # labels       = "node"
        #   }
        # }
      }

      template {
        data        = <<EOH
{{- with nomadVar (printf "nomad/jobs/%s" (env "NOMAD_JOB_NAME")) -}}
VERSION={{ .version }}
{{ end -}}
EOH
        destination = "${NOMAD_TASK_DIR}/node.env"
        change_mode = "restart"
        env         = true
      }

      template {
        data        = <<EOF
{{- with nomadVar (printf "nomad/jobs/%s" (env "NOMAD_JOB_NAME")) -}}
{
  "app": {
    "checkForUpdates": true,
    "shutdown": {
      "stopGracePeriod": "5m",
      "log": {
        "enabled": true,
        "filePath": "{{ env "NOMAD_TASK_DIR" }}/shutdown.log"
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
      "{{ env "NOMAD_TASK_DIR" }}/wasp.log"
    ],
    "disableEvents": true
  },
  "inx": {
    "address": "{{ range service "inx.tangle-testnet" }}{{ .Address }}:{{ .Port }}{{ end }}",
    "maxConnectionAttempts": 30
  },
  "db": {
    "engine": "rocksdb",
    "chainState": {
      "path": "{{ env "NOMAD_TASK_DIR" }}/chains/data"
    },
    "debugSkipHealthCheck": false
  },
  "p2p": {
    "identity": {
      "privateKey": "",
      "filePath": "{{ env "NOMAD_TASK_DIR" }}/identity/identity.key"
    },
    "db": {
      "path": "{{ env "NOMAD_TASK_DIR" }}/p2pstore"
    }
  },
  "registries": {
    "chains": {
      "filePath": "{{ env "NOMAD_TASK_DIR" }}/chains/chain_registry.json"
    },
    "dkShares": {
      "path": "{{ env "NOMAD_TASK_DIR" }}/dkshares"
    },
    "trustedPeers": {
      "filePath": "{{ env "NOMAD_TASK_DIR" }}/trusted_peers.json"
    },
    "consensusState": {
      "path": "{{ env "NOMAD_TASK_DIR" }}/chains/consensus"
    }
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
        "whitelist": {{ .admin_whitelist }}
      }
    }
  },
  "nanomsg": {
    "enabled": true,
    "port": {{ env "NOMAD_PORT_nanomsg" }}
  },
  "dashboard": {
    "enabled": true,
    "bindAddress": "0.0.0.0:{{ env "NOMAD_PORT_dashboard" }}",
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
        "whitelist": {{ .admin_whitelist }}
      }
    }
  },
  "peering":{
      "port": {{ env "NOMAD_PORT_peering" }},
      "peeringURL": "{{ env "NOMAD_ADDR_peering" }}"
  },
  "profiling":{
    "enabled": false,
    "bindAddress": "{{ env "NOMAD_ADDR_profiling" }}"
  },
  "inx": {
    "address": "{{ range service "inx.tangle-testnet" }}{{ .Address }}:{{ .Port }}{{ end }}",
    "maxConnectionAttempts": 30
  },
  "wal": {
    "directory": "{{ env "NOMAD_TASK_DIR" }}/wal",
    "enabled": true
  },
  "debug": {
    "rawblocksEnabled": false,
    "rawblocksDirectory": "{{ env "NOMAD_TASK_DIR" }}/blocks"
  }
}
{{- end -}}
EOF
        destination = "${NOMAD_TASK_DIR}/config.json"
        perms       = "444"
      }

      resources {
        memory = 4000
        cpu    = 3000
      }
    }
  }

  // group "access" {
  //   ephemeral_disk {
  //     migrate = true
  //     sticky  = true
  //   }

  //   count = 0

  //   network {
  //     mode = "host"

  //     port "dashboard" {
  //       host_network = "private"
  //     }
  //     port "api" {
  //       host_network = "private"
  //     }
  //     port "nanomsg" {
  //       host_network = "private"
  //     }
  //     port "peering" {
  //       host_network = "private"
  //     }
  //     port "metrics" {
  //       host_network = "private"
  //     }
  //     port "profiling" {
  //       host_network = "private"
  //     }
  //     port "dlv" {
  //       static = 40000
  //       to = 40000
  //     }
  //   }

  //   task "wasp" {
  //     driver = "docker"

  //     config {
  //      network_mode = "host"
  //       image        = "${artifact.image}:${artifact.tag}"
  //       entrypoint   = ["wasp", "-c", "/local/config.json"]
  //       ports = [
  //         "dashboard",
  //         "api",
  //         "nanomsg",
  //         "peering",
  //         "metrics",
  //         "profiling"
  //       ]


  //       labels = {
  //         "co.elastic.metrics/raw" = "[{\"metricsets\":[\"collector\"],\"module\":\"prometheus\",\"period\":\"10s\",\"metrics_path\":\"/metrics\",\"hosts\":[\"$${NOMAD_ADDR_metrics}\"]}]"
  //         "wasp"                   = "access"
  //       }

  //       // logging {
  //       //   type = "gelf"
  //       //   config {
  //       //     gelf-address          = "tcp://elastic-logstash-beats-logstash.service.consul:12201"
  //       //     tag                   = "wasp"
  //       //     labels                = "wasp"
  //       //   }
  //       // }

  //       auth {
  //         username       = "${auth.username}"
  //         password       = "${auth.password}"
  //         server_address = "${auth.server_address}"
  //       }
  //     }

  //     service {
  //       tags = ["wasp", "api"]
  //       port = "api"
  //       check {
  //         type     = "http"
  //         port     = "api"
  //         path     = "info"
  //         interval = "5s"
  //         timeout  = "2s"
  //       }
  //     }
  //     service {
  //       tags = ["wasp", "dashboard"]
  //       port = "dashboard"
  //     }
  //     service {
  //       tags = ["wasp", "nanomsg"]
  //       port = "nanomsg"
  //     }
  //     service {
  //       tags = ["wasp", "peering"]
  //       port = "peering"
  //     }
  //     service {
  //       tags = ["wasp", "metrics"]
  //       port = "metrics"
  //     }
  //     service {
  //       tags = ["wasp", "profiling"]
  //       port = "profiling"
  //     }
  //     service {
  //       tags = ["wasp", "dlv"]
  //       port = "dlv"
  //     }

  //     template {
  //       data        = var.wasp_config
  //       destination = "/local/config.json"
  //       perms       = "777"
  //     }

  //     resources {
  //       memory = 4000
  //       cpu    = 3000
  //     }
  //   }
  // }
}
