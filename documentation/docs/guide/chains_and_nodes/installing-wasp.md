---
description: Installing Wasp.
image: /img/logo/WASP_logo_dark.png
keywords:
  - Wasp
  - Installing wasp
  - Go-lang
  - Hornet
---

# Installing Wasp

In this page you can find information on how to use Wasp by cloning the repository and building the application. The instructions below will build both the Wasp node and the Wasp CLI to interact with the node from the command line.

Alternativaly, you can run a Wasp node using one of the provided docker setups:

- [Wasp standalone Docker image](docker_standalone.md)
- pre-configured local [Wasp and Hornet node setup using Docker Compose](../development_tools/docker_preconfigured.md).

## Clone the Wasp repository

You can get the source code of the latest Wasp version from the [official repository](https://github.com/iotaledger/wasp).

```shell
git clone https://github.com/iotaledger/wasp
```

## Building/Installing

You can build and install both `wasp` and `wasp-cli` by running the following commands.

### Linux/macOS

```shell
make install
```

### macOS arm64 (M1 Apple Silicon)

[`wasmtime-go`](https://github.com/bytecodealliance/wasmtime-go) hasn't supported macOS on arm64 yet, so you should build your own wasmtime library. You can follow the README in `wasmtime-go` to build the library.
Once a wasmtime library is built, then you can run the following commands.

```shell
go mod edit -replace=github.com/bytecodealliance/wasmtime-go=<wasmtime-go path>
make install
```

### Microsoft Windows

Its recommended to use [WSL](https://docs.microsoft.com/en-us/windows/wsl/install) on windows and follow the [Linux/macOS](#linuxmacos) instructions above.

:::info

The install command will place the applications binaries in `$GOPATH/bin`.
Make sure that directory is part of your `$PATH` environment variable.
If needed, you can include this location in `$PATH` by adding the following line to your `~/.bash_profile`:

```shell
export PATH=$PATH:$(go env GOPATH)/bin
```

To apply changes made to a profile file either restart your terminal application or execute `source ~/.bash_profile`

:::
