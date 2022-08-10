# The name of your project. A project typically maps 1:1 to a VCS repository.
# This name must be unique for your Waypoint server. If you're running in
# local mode, this must be unique to your machine.
project = "isc"

# Labels can be specified for organizational purposes.
labels = { "team" = "isc" }

variable "wasp_url" {
    default = "api.wasp.sc.iota.org"
    type = string
}

variable "goshimmer_url" {
    default = "api.goshimmer.sc.iota.org"
    type = string
}

variable "chainid" {
    type = string
}

variable "adminWhitelist" {
    type = list(string)
}

variable "ghcr" {
    type = object({
        username = string
        password = string
        server_address = string
    })
}

# An application to deploy.
app "fairroulette" {
    # Build specifies how an application should be deployed. In this case,
    # we'll build using a Dockerfile and keeping it in a github registry.
    build {
        use "docker" {
            disable_entrypoint = true
            buildkit   = true
            build_args = {
                WASP_URL = "https://${var.wasp_url}"
                WASP_WS_URL = "wss://${var.wasp_url}"
                GOSHIMMER_URL = "https://${var.goshimmer_url}"
                CHAIN_ID = var.chainid
            }
        }

        registry {
            use "docker" {
                image = "ghcr.io/luke-thorne/fairroulette"
                tag = gitrefpretty()
                encoded_auth = base64encode(jsonencode(var.ghcr))
            }
        }
    }

    # Deploy to Nomad
    deploy {
        use "nomad-jobspec" {
            // Templated to perhaps bring in the artifact from a previous
            // build/registry, entrypoint env vars, etc.
            jobspec = templatefile("${path.app}/fairroulette.nomad.tpl", { 
                artifact = artifact
                auth = var.ghcr
            })
        }
    }
}
