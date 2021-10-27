job "fairroulette" {
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

	group "server" {
		network {
			mode = "host"

			port "http" {
				host_network = "private"
				to = "80"
			}
		}

		task "worker" {
			driver = "docker"

			config {
				network_mode = "bridge"
				image = "${artifact.image}:${artifact.tag}"
				ports = [
					"http",
				]

				auth {
					username = "${auth.username}"
					password = "${auth.password}"
					server_address = "${auth.server_address}"
				}
			}

			service {
				tags = ["fairroulette", "http"]
				port  = "http"
				check {
					type     = "http"
					port     = "http"
					path     = "/"
					interval = "5s"
					timeout  = "2s"
				}
			}

			resources {
				memory = 128
				cpu = 128
			}
		}
	}
}