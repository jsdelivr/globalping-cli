# Contributing guide

Hi! We're really excited that you're interested in contributing! Before submitting your contribution, please read through the following guide.

## General guidelines

- Bug fixes and changes discussed in the existing issues are always welcome.
- For new ideas, please open an issue to discuss them before sending a PR.
- Make sure your PR passes `go test ./...` and has [appropriate commit messages](https://github.com/jsdelivr/globalping-cli/commits/master).

## Project setup

Install golangci-lint:

```shell
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b bin v1.52.2
```

Install mockgen:

```shell
GOBIN=$(pwd)/bin go install go.uber.org/mock/mockgen@latest
```

Run golangci-lint:

```shell
bin/golangci-lint run
```

### Testing

Run tests:

```shell
go test ./...
```

To regenerate the mocks;

```shell
mocks/gen_mocks.sh
```
