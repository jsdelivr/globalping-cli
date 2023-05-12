<h1 align="center"> <a href="https://www.jsdelivr.com/globalping"><img width="28" alt="Globalping icon" src="https://user-images.githubusercontent.com/1834071/216975126-01529980-a87e-478c-8ab3-bf7d927a1986.png"></a> Globalping CLI </h1>

<p align="center">Access a global network of probes without leaving your console. Powered by the Globalping community!</p>

<p align="center"><img width="80%" src="https://user-images.githubusercontent.com/1834071/217010016-9da38f12-906a-47cf-adca-18017588efe5.png">
</p>

- The official command-line interface for the [Globalping](https://github.com/jsdelivr/globalping) network.
- Run networking commands from any location in the world
- Supported commands: ping, mtr, traceroute, dns resolve, HTTP
- Real-time results right in your command line
- Human friendly format and output
- Cross-platform. Linux, MacOS, Windows are all supported
- Auto-updates via RPM/DEB/Chocolatey repos
- [Check our website for online tools, our Slack app and more!](https://www.jsdelivr.com/globalping)

## Installation - Quick start

Simply run these commands to install the repo and CLI! This way you will get all future updates by simply running an update using your package manager.

### Ubuntu/Debian (deb)

```shell
curl -s https://packagecloud.io/install/repositories/jsdelivr/globalping/script.deb.sh | sudo bash
apt install globalping
```

### CentOS/Fedora/Rocky Linux/AlmaLinux (rpm)

```shell
curl -s https://packagecloud.io/install/repositories/jsdelivr/globalping/script.rpm.sh | sudo bash
dnf install globalping
```

[Manual installation instructions](https://packagecloud.io/jsdelivr/globalping/install#manual)

### MacOS - Homebrew

```shell
brew tap jsdelivr/globalping
brew install globalping
```

### Windows - [Chocolatey](https://community.chocolatey.org/packages/globalping)

```shell
choco install globalping
```

### Binary installation

Every new release is compiled into binaries ready to run on most operating systems and attached as assets on GitHub. You can download and run the binaries directly on your system, but note that you will have to repeat this process for every new release. [Explore the available versions](https://github.com/jsdelivr/globalping-cli/releases).


## Getting Started with Globalping CLI

Once the Globalping CLI is installed, you can verify that it is working by running:

```bash
globalping --help

Globalping is a platform that allows anyone to run networking commands such as ping, traceroute, dig and mtr on probes distributed all around the world.
The CLI tool allows you to interact with the API in a simple and human-friendly way to debug networking issues like anycast routing and script automated tests and benchmarks.

Usage:
  globalping [command]

Measurement Commands:
  dns           Resolve a DNS record similarly to dig
  http          Perform a HEAD or GET request to a host
  mtr           Run an MTR test, similar to traceroute
  ping          Run a ping test
  traceroute    Run a traceroute test

Additional Commands:
  completion    Generate the autocompletion script for the specified shell
  help          Help about any command
  install-probe Join the community powered Globalping platform by running a Docker container.
  version       Print the version number of Globalping CLI

Flags:
  -C, --ci            Disable realtime terminal updates and color suitable for CI and scripting (default false)
  -F, --from string   Comma-separated list of location values to match against. For example the partial or full name of a continent, region (e.g eastern europe), country, US state, city or network (default "world"). (default "world")
  -h, --help          help for globalping
  -J, --json          Output results in JSON format (default false)
      --latency       Output only the stats of a measurement (default false). Only applies to the dns, http and ping commands
  -L, --limit int     Limit the number of probes to use (default 1)

Use "globalping [command] --help" for more information about a command.
```

## Development setup
Install golangci-lint
```shell
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b bin v1.52.2
```

Install mockgen
```shell
GOBIN=$(pwd)/bin go install github.com/golang/mock/mockgen@v1.6.0
```

Run golangci-lint
```shell
bin/golangci-lint run
```

Run tests
```shell
go test ./...
```

To regenerate the mocks
```shell
mocks/gen_mocks.sh
```

