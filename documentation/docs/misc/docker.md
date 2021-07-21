# Docker

This page describes the configuration of the Wasp node in combination with Docker.

## Introduction

The dockerfile is separated into several stages which effectively splits Wasp into four small pieces:

* Testing
    * Unit testing
    * Integration testing
* Wasp CLI
* Wasp Node

## Running a Wasp node

Checkout the project, switch to 'develop' and build the main image:

```
$ git clone -b develop https://github.com/iotaledger/wasp.git
$ cd wasp
$ docker build -t wasp-node .
```

The build process will copy the docker_config.json file into the image which will use it when the node gets started. 

By default, the build process will use `-tags rocksdb` as a build argument. This argument can be modified with `--build-arg BUILD_TAGS=<tags>`.

Depending on the use case, Wasp requires a different GoShimmer hostname which can be changed at this part inside the docker_config.json file: 
```
  "nodeconn": {
    "address": "goshimmer:5000"
  },
```

The Wasp node can be started like so:

```
$ docker run wasp-node
```

### Configuration

After the build process has been completed, it is still possible to inject a different configuration file into a new container. 

```
$ docker run -v $(pwd)/alternative_docker_config.json:/run/config.json wasp-node
```

Further configuration is possible using arguments:

```
$ docker run wasp-node --nodeconn.address=alt_goshimmer:5000 
```

To get a list of all available arguments, run the node with the argument '--help'

```
$ docker run wasp-node --help
```

# Wasp CLI

It is possible to create a micro image that just contains the wasp-cli application without any Wasp node related additions.

This might be helpful if it's required to control but not to run a Wasp node.

The image can be created like this:

```
$docker build --target wasp-cli -t wasp-cli . 
```

Like with the Wasp node setup, the container gets started by:

```
$ docker run wasp-cli
```

and can be controlled with further arguments:


```
$ docker run wasp-cli --help
```

# Testing

Wip
