---
keywords:
- Smart Contracts
- Contribute
- Pull Request
- Linting
- Go-lang
- golangci-lint
description:  How to contribute to IOTA Smart Contracts. Creating a PR, setting up golangci-lint.  
image: /img/logo/WASP_logo_dark.png
---

# Contributing

If you want to contribute to this repository, consider posting a [bug report](https://github.com/iotaledger/wasp/issues/new-issue), feature request or a [pull request](https://github.com/iotaledger/wasp/pulls/).

You can also join our [Discord server](https://discord.iota.org/) and ping us
in `#smartcontracts-dev`.

## Creating a Pull Request

Please base your work on the `develop` branch.

Before creating the Pull Request ensure that:

- all the tests pass:

    ```bash
    go test -tags rocksdb,builtin_static ./...
    ```

- there are no linting violations (instructions on how to setup linting below):

    ```bash
    golangci-lint run
    ```

### Lint Setup

1. Install golintci:

    https://golangci-lint.run/usage/install/#local-installation

2. Dev setup:

    https://golangci-lint.run/usage/integrations/#editor-integration

    **VSCode**:

    ```json
    // required:
    "go.lintTool": "golangci-lint",
    // recommended:
    "go.lintOnSave": "package"
    "go.lintFlags": ["--fix"],
    "editor.formatOnSave": true,
    ```

    **GoLand**:

    - [Install golintci plugin](https://plugins.jetbrains.com/plugin/12496-go-linter)

        ![Install golintci plugin](../static/img/contributing/golintci-goland-1.png)

    - Configure path for golangci

        ![Configure path for golangci](../static/img/contributing/golintci-goland-2.png)

    - Add a golangci file watcher with custom command (I recommend using --fix)

        ![Add a golangci file watcher with custom command](../static/img/contributing/golintci-goland-3.png)

    **Other editors**: please look into the [`golangci` official documentation](https://github.com/golangci/golangci-lint).

3. Ignoring false positives:

    https://golangci-lint.run/usage/false-positives/

    ```go
    //nolint
    ```

    for specific rules:

    ```go
    //nolint:golint,unused
    ```
