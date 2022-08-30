# The name of your project. A project typically maps 1:1 to a VCS repository.
# This name must be unique for your Waypoint server. If you're running in
# local mode, this must be unique to your machine.
project = "isc"

# Labels can be specified for organizational purposes.
labels = { "team" = "isc" }

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
app "wasp-evm" {
    # Build specifies how an application should be deployed. In this case,
    # we'll build using a Dockerfile and keeping it in a local registry.
    build {
        use "docker" {
            disable_entrypoint = true
            buildkit   = true
            dockerfile = "./Dockerfile"
            build_args = {
                BUILD_TAGS = "rocksdb,builtin_static"
                BUILD_LD_FLAGS = "-X github.com/iotaledger/wasp/packages/wasp.VersionHash=${gitrefhash()}"
            }
        }

        registry {
            use "docker" {
                image = "ghcr.io/luke-thorne/wasp"
                tag = gitrefhash()
                encoded_auth = base64encode(jsonencode(var.ghcr))
            }
        }
    }

    # Deploy to Nomad
    deploy {
        use "nomad-jobspec" {
            // Templated to perhaps bring in the artifact from a previous
            // build/registry, entrypoint env vars, etc.
            jobspec = templatefile("${path.app}/wasp.nomad.tpl", { 
                artifact = artifact
                adminWhitelist = jsonencode(var.adminWhitelist)
                auth = var.ghcr
                workspace = workspace.name
            })
        }
    }
}
