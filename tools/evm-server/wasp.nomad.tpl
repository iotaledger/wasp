variable "wasp_cli_config" {
	default = <<EOH
{
  "chain": "testchain",
  "chains": {
    "testchain": "${chainid}"
  },
  "goshimmer": {
    "api": "https://goshimmer.sc.iota.org/api"
  },
  "wallet": {
    "seed": "${wallet_seed}"
  },
  "wasp": {
    "0": {
      "api": "https://wasp.sc.iota.org/api"
    }
  }
}
EOH
}

job "iscp-evm-server" {
	datacenters = ["hcloud"]

	update {
		max_parallel      = 1
		health_check      = "checks"
		min_healthy_time  = "15s"
		healthy_deadline  = "1m"
		progress_deadline = "3m"
		auto_revert       = true
		auto_promote      = true
		canary            = 1
		stagger           = "30s"
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
		}

		task "wasp-cli" {
			driver = "docker"

			config {
				network_mode = "host"
				image = "${artifact.image}:${artifact.tag}"
				command = "wasp-cli"
				args = [
					"-c=$${NOMAD_TASK_DIR}/wasp-cli.json",
					"chain",
					"evm",
					"jsonrpc",
					"--chainid=1074",
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
	}
}