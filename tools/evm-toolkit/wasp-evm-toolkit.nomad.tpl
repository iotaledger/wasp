job "isc-evm-toolkit-${workspace}" {
	datacenters = ["hcloud"]

	group "web" {
		ephemeral_disk {
			migrate = false
			sticky = true
		}

		network {

			port "http" {
				host_network = "private"
				to = 80
			}
		}

		task "worker" {
			driver = "docker"

			config {
				image = "${artifact.image}:${artifact.tag}"
				entrypoint   = ["nginx", "-g", "daemon off;"]
				ports = [
					"http",
				]

				auth {
					username = "${auth.username}"
					password = "${auth.password}"
					server_address = "${auth.server_address}"
				}
			}

			env {
				%{ for k,v in entrypoint.env ~}
				${k} = "${v}"
				%{ endfor ~}

				// Ensure we set PORT for the URL service. This is only necessary
				// if we want the URL service to function.
				PORT = "$${NOMAD_ALLOC_PORT_http}"
			}

			service {
				tags = ["http"]
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
				memory = 256
				cpu = 256
			}
		}
	}
}
