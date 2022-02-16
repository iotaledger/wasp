---
description: How to run a Wasp node in using Docker. Build the image, configure it, run it.
image: /img/logo/WASP_logo_dark.png
keywords:
  - Smart Contracts
  - Running a node
  - docker
  - image
  - build
  - configure
  - arguments
---

# Docker (Standalone)

This page describes the configuration of a single Wasp node in combination with Docker. If you followed the instructions in [Running a Node](running-a-node.md), you can skip to [Configuring wasp-cli](wasp-cli.md).

## Introduction

## Running a Wasp Node

Checkout the project, switch to 'develop' and build the main image:

```shell
git clone https://github.com/iotaledger/wasp.git
cd wasp
docker build -t wasp-node .
```

The build process will copy the docker_config.json file into the image, which will be used when the node gets started.

By default, the build process will use `-tags rocksdb,builtin_static` as a build argument. This argument can be modified with `--build-arg BUILD_TAGS=<tags>`.

Depending on the use case, Wasp requires a different GoShimmer hostname which can be changed at this part inside the [docker_config.json](https://github.com/iotaledger/wasp/blob/develop/docker_config.json) file:

```json
  "nodeconn": {
    "address": "goshimmer:5000"
  },
```

After the build process has finished, you can start your Wasp node by running:

```shell
docker run wasp-node
```

### Configuration

After the build process has been completed, it is still possible to inject a different configuration file into a new container by running:

```shell
docker run -v $(pwd)/alternative_docker_config.json:/etc/wasp_config.json wasp-node
```

You can also add further configuration using arguments:

```shell
docker run wasp-node --nodeconn.address=alt_goshimmer:5000
```

To get a list of all available arguments, run the node with the argument '--help'

```shell
docker run wasp-node --help
```
