# Contributing

## Creating a Pull Request

Please base your work on the `develop` branch.

Before creating the Pull Request ensure that:

- all the tests pass:

    ```bash
    go test -tags rocksdb ./...
    ```

- there are no linting violations (instructions on how to setup linting below):

    ```bash
    golangci-lint run
    ```

## Lint Setup

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

    - install golintci plugin: https://plugins.jetbrains.com/plugin/12496-go-linter

        ![install plugin](./golintci-goland-1.png)

    - configure path for golangci

        ![configue plugin](./golintci-goland-2.png)

    - add a golangci file watcher with custom command (I recommend using --fix)

        ![watcher plugin](./golintci-goland-3.png)

    **Other editors**: please look into the `golangci` official documentation.

3. Ignoring false positives:

    https://golangci-lint.run/usage/false-positives/

    ```go
    //nolint
    ```

    for specific rules:

    ```go
    //nolint:golint,unused
    ```
