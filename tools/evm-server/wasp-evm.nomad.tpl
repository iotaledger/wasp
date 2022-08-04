variable "wasp_cli_config" {
	default = <<EOH
{
  "chain": "evmchain",
  "chains": {
    "evmchain": "${chainid}"
  },
  "goshimmer": {
    "api": "https://api.goshimmer.sc.iota.org"
  },
  "wallet": {
    "seed": "${wallet_seed}"
  },
  "wasp": {
    "0": {
      "api": "https://api.wasp.sc.iota.org"
    }
  }
}
EOH
}

job "isc-evm-server" {
	datacenters = ["hcloud"]

	update {
		max_parallel      = 1
		health_check      = "checks"
		min_healthy_time  = "5s"
		healthy_deadline  = "1m"
		progress_deadline = "3m"
		auto_revert       = true
		auto_promote      = true
		canary            = 1
		stagger           = "5s"
	}

	group "node" {
		ephemeral_disk {
			migrate = true
			sticky = true
		}

		network {
			mode = "host"

			port "evm" {
				host_network = "private"
			}

			port "explorer" {
				host_network = "private"
				to = "80"
			}
		}

		task "wasp-cli" {
			driver = "docker"

			config {
				network_mode = "host"
				image = "${artifact.image}:${artifact.tag}"
				entrypoint = [""]
				command = "wasp-cli"
				args = [
					"-c=$${NOMAD_TASK_DIR}/wasp-cli.json",
					"chain",
					"evm",
					"jsonrpc",
					"-d",
					"--chainid",
					"1074",
					"-l=0.0.0.0:$${NOMAD_PORT_evm}",
				]
				ports = [
					"evm",
				]

				auth {
					username = "${auth.username}"
					password = "${auth.password}"
					server_address = "${auth.server_address}"
				}
			}

			service {
				tags = ["wasp-cli", "evm"]
				port  = "evm"
				check {
					type     = "http"
					port     = "evm"
					path     = "/"
					interval = "5s"
					timeout  = "2s"
				}
			}

			template {
				data = var.wasp_cli_config
				destination = "/local/wasp-cli.json"
				perms = "444"
			}

			resources {
				memory = 256
				cpu = 256
			}
		}

		task "explorer" {
			driver = "docker"

			config {
				network_mode = "bridge"
				image = "alethio/ethereum-lite-explorer"
				ports = [
					"explorer",
				]
			}

			env {
				APP_NODE_URL = "https://evm.wasp.sc.iota.org"
			}

			service {
				tags = ["wasp-cli", "explorer"]
				port  = "explorer"
				check {
					type     = "http"
					port     = "explorer"
					path     = "/"
					interval = "5s"
					timeout  = "2s"
				}
			}

			resources {
				memory = 256
				cpu = 256
			}
		}
	}
}