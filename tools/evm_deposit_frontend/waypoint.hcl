# The name of your project. A project typically maps 1:1 to a VCS repository.
# This name must be unique for your Waypoint server. If you're running in
# local mode, this must be unique to your machine.
project = "isc"

# Labels can be specified for organizational purposes.
labels = { "team" = "isc" }

variable "ghcr" {
    type = object({
        username = string
        password = string
        server_address = string
    })
}

app "evm-deposit" {
    build {
        use "docker" {
            disable_entrypoint = true
            buildkit   = true
            dockerfile = "./Dockerfile"
        }

        registry {
            use "docker" {
                image = "ghcr.io/luke-thorne/evm_deposit_frontend"
                tag = gitrefpretty()
                encoded_auth = base64encode(jsonencode(var.ghcr))
            }
        }
    }

    deploy {
        use "nomad-jobspec" {
            // Templated to perhaps bring in the artifact from a previous
            // build/registry, entrypoint env vars, etc.
            jobspec = templatefile("${path.app}/wasp-evm-deposit.nomad.tpl", { 
                artifact = artifact
                auth = var.ghcr
                workspace = workspace.name
            })
        }
    }
}