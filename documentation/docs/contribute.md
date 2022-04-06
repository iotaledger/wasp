---
description:  How to contribute to IOTA Smart Contracts. How to create better pull requests by running tests and the linter locally.
image: /img/logo/WASP_logo_dark.png
keywords:
- smart contracts
- Contribute
- pull request
- linting
- Go-lang
- golangci-lint
- how to
---

# Contributing

If you want to contribute to this repository, consider posting a [bug report](https://github.com/iotaledger/wasp/issues/new-issue), feature request, or a [pull request](https://github.com/iotaledger/wasp/pulls/).

You can talk to us directly on our [Discord server](https://discord.iota.org/), in the `#smartcontracts-dev` channel.

## Creating a Pull Request

Please base your work on the `develop` branch.

Before creating a pull request ensure that all tests pass locally, and that the linter reports no violations.

## Running Tests

To run tests locally, execute one of the following commands:

```shell
go test -tags rocksdb,builtin_static ./...
```

or, as an alternative:

```shell
make test
```

The commands above only trigger lightweight tests. If you have introduced major changes, consider running the whole test suite instead. That it will much longer (and will time out after an hour).

::: warn

Check that the `database.inMemory` parameter is set to `false` (default) before running the complete test suite. See the [source code](https://github.com/iotaledger/wasp/blob/develop/tools/cluster/tests/spam_test.go) for details.

:::

```shell
make test-full
```

## Running the Linter

### Setup

#### Step 1: Install golintci

See the [provider instructions](https://golangci-lint.run/usage/install/#local-installation) on how to install golintci.

#### Step 2: Set Up Your Environment

See the [provider instructions](https://golangci-lint.run/usage/integrations/#editor-integration) on how to integrate golintci into your source code editor. You can also find our [recommended settings](#appendix-recommended-settings) for VS Code and GoLand at the bottom of this article.

### Usage

To run the linter locally, execute:

```shell
golangci-lint run
```

or

```shell
make lint
```

The linter will also automatically run every time you run:

```shell
make
```

### False Positives

You can [disable](https://golangci-lint.run/usage/false-positives/) false positives by placing a special comment directly above the "violating" element:

```go
//nolint
func foobar() *string {
    // ...
}
```

To be sure that linter will not ignore actual issues in the future, try to suppress only relevant warnings over an element:

```go
//nolint:golint,unused
func neverUsedFoobar() *string {
    // ...
}
```

## Appendix: Recommended Settings

### Visual Studio Code

Adjust your VS Code settings as follows:

```json
// required:
"go.lintTool": "golangci-lint",
// recommended:
"go.lintOnSave": "package"
"go.lintFlags": ["--fix"],
"editor.formatOnSave": true,
```

### GoLand

1. Install the [golintci](https://plugins.jetbrains.com/plugin/12496-go-linter) plugin.

![A screenshot that shows how to install golintci in GoLand.](/img/contributing/golintci-goland-1.png "Click to see the full-sized image.")

2. Configure path for golangci.

![A screenshot that shows how to configure path for golangci in GoLand.](/img/contributing/golintci-goland-2.png "Click to see the full-sized image.")

3. Add a golangci file watcher with a custom command. We recommend you to use it with the `--fix` parameter.

![A screenshot that shows how to add a golangci file watcher in GoLand.](/img/contributing/golintci-goland-3.png "Click to see the full-sized image.")