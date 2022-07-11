---
description: How to run a Wasp node in using Docker. Build the image, configure it, run it.
image: /img/Banner/banner_wasp_using_docker.png
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

![Wasp Node using Docker](/img/Banner/banner_wasp_using_docker.png)

This page describes the configuration of a single Wasp node in combination with Docker. 

The docker setup comes preconfigured and should work as is, differing setups might require a different configuration.

In this case the following instructions should be read [Running a Node](running-a-node.md).

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

Depending on the use case, it might be required to change the default Hornet configuration, which can be changed in this location inside the [docker_config.json](https://github.com/iotaledger/wasp/blob/develop/docker_config.json) file:

```json
"l1": {
  "apiAddress": "http://hornet:14265",
  "faucetAddress": "http://hornet:8191"
},
```

After the build process has finished, you can start your Wasp node by running:

```shell
docker run wasp-node
```

### Configuration of built images 

After the build process has been completed, it is still possible to inject a different configuration file into a new container by running:

```shell
docker run -v $(pwd)/alternative_docker_config.json:/etc/wasp_config.json wasp-node
```

You can also add further configuration using arguments:

```shell
docker run wasp-node --l1.apiAddress="alt_hornet:14265"
```

To get a list of all available arguments, run the node with the argument '--help'

```shell
docker run wasp-node --help
```
