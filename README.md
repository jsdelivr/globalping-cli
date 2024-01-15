<h1 align="center"> <a href="https://www.jsdelivr.com/globalping"><img width="28" alt="Globalping icon" src="https://user-images.githubusercontent.com/1834071/216975126-01529980-a87e-478c-8ab3-bf7d927a1986.png"></a> Globalping CLI </h1>

<p align="center">Access a global network of probes without leaving your console. Benchmark your internet infrastructure, automate uptime and latency monitoring with scripts, or optimize your anycast network – from any location and free of charge. Powered by the Globalping community!</p>

<p align="center"><img width="80%" src="https://user-images.githubusercontent.com/1834071/217010016-9da38f12-906a-47cf-adca-18017588efe5.png">
</p>

## Key features

- The official command-line interface for the [Globalping](https://github.com/jsdelivr/globalping) network
- Run networking commands from any location in the world
- Real-time results for all supported commands: ping, mtr, traceroute, DNS resolve, HTTP
- Includes detailed timings and latency metrics for every test
- Human-friendly format and output
- Supports Linux, MacOS, and Windows
- Auto-updated via all automated installation methods
- Explore additional [Globalping integrations](https://www.jsdelivr.com/globalping/integrations), including our online tools, Slack app, and more

## Installation

Install the repository and Globalping CLI using the relevant package manager command from below. This way, you can get future updates by simply running an update with your package manager.

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

### Windows - WinGet

```shell
winget install globalping
```

### Binary installation

Every new release is compiled into binaries ready to run on most operating systems and provided as assets on GitHub. You can download and execute these binaries directly on your system.

> [!IMPORTANT]
> Opting for this installation method means you'll have to repeat this manual process to update the CLI to a newer release!

[Explore the available versions](https://github.com/jsdelivr/globalping-cli/releases).

## Updating

If you've installed the Globalping CLI via a package manager, you only need to run the manager's update command to get the latest Globalping CLI version.

## Getting started with Globalping CLI

After installing, verify the Globalping CLI is working by running:

`globalping --help`

The result shows how to use the CLI and which commands and flags are available:

```bash
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
  -F, --from string   Comma-separated list of location values to match against or a measurement ID
                      For example, the partial or full name of a continent, region (e.g eastern europe), country, US state, city or network
                      Or use [@1 | first, @2 ... @-2, @-1 | last | previous] to run with the probes from previous measurements. (default "world")
  -h, --help          help for globalping
  -J, --json          Output results in JSON format (default false)
      --latency       Output only the stats of a measurement (default false). Only applies to the dns, http and ping commands
  -L, --limit int     Limit the number of probes to use (default 1)

Use "globalping [command] --help" for more information about a command.
```

### Run your first tests

Globalping relies on a community-hosted probe network, enabling you to run network tests from any location with an active probe. The following examples show you through some tests, exploring how to define locations, set limits, and use some command flags.

#### Filter locations

For example, if you want to run ping from a probe in Seattle that is also part of the Comcast network, run the following:

```bash
globalping ping google.com from Comcast+Seattle
> NA, US, (WA), Seattle, ASN:7922, Comcast Cable Communications, LLC
PING  (142.250.217.78) 56(84) bytes of data.
64 bytes from sea09s29-in-f14.1e100.net (142.250.217.78): icmp_seq=1 ttl=58 time=14.0 ms
64 bytes from sea09s29-in-f14.1e100.net (142.250.217.78): icmp_seq=2 ttl=58 time=14.5 ms
64 bytes from sea09s29-in-f14.1e100.net (142.250.217.78): icmp_seq=3 ttl=58 time=15.9 ms

---  ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 402ms
rtt min/avg/max/mdev = 13.985/14.779/15.886/0.807 ms
```

You can use the `+` symbol as a filter to select the desired location of the probes more precisely.

> [!TIP]
> You can mix and match any location type, including countries, continents, cities, US states, regions, ASNs, ISP names, eyeball or data center tags, and cloud region names.

#### Define multiple locations and basic flags

With the following command, we execute four ping commands at four different locations and obtain the summarized latency metrics for each test as a result:

```bash
globalping ping google.com from Amazon,Germany,USA,Dallas --limit 4 --latency
> AS, KR, Seoul, ASN:16509, Amazon.com, Inc. (aws-ap-northeast-2)
Min: 33.163 ms
Max: 33.256 ms
Avg: 33.22 ms

> EU, DE, Frankfurt, ASN:16276, OVH SAS
Min: 1.221 ms
Max: 1.291 ms
Avg: 1.264 ms

> NA, US, (IL), Chicago, ASN:174, Cogent Communications
Min: 112.405 ms
Max: 112.686 ms
Avg: 112.528 ms

> NA, US, (TX), Dallas, ASN:393336, Catalyst Host LLC
Min: 1.579 ms
Max: 1.588 ms
Avg: 1.584 ms
```

You can select multiple locations for running a command by using a comma `,` as a delimiter. When doing so, make sure to also specify the number of tests to run with the `--limit` flag.
For example, to run ping from four different locations (as we did in the example above), add `--limit 4` to make sure you get one test result per location. Otherwise, the default limit of 1 will be selected, resulting in a random result from one of the four locations.

Finally, you can use the `--latency` parameter to only get the summarized latency data instead of the full raw output.

> [!TIP]
> We recommend reading our [tips and best practices](https://github.com/jsdelivr/globalping#best-practices-and-tips) to learn more about defining locations effectively!

#### Share results online

Include a link at the bottom of your results using the `--share` flag to view and share the test results online.

> [!IMPORTANT]
> Shareable links and the respective saved measurement results expire after a few weeks, depending on the user type. GitHub Sponsors, for example, enjoy extended result storage.

```bash
 globalping dns google.com from gcp-asia-south1 --share
> AS, IN, Mumbai, ASN:396982, Google LLC (gcp-asia-south1)
; <<>> DiG 9.16.37-Debian <<>> -t A google.com -p 53 -4 +timeout=3 +tries=2 +nocookie +nsid
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 23733
;; flags: qr rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 512
;; QUESTION SECTION:
;google.com.                    IN      A

;; ANSWER SECTION:
google.com.             300     IN      A       142.250.183.206

;; Query time: 3 msec
;; SERVER: x.x.x.x#53(x.x.x.x)
;; WHEN: Mon Jul 10 10:38:00 UTC 2023
;; MSG SIZE  rcvd: 55
> View the results online: https://www.jsdelivr.com/globalping?measurement=xrfXUEAOGfzwfHFz
```

#### Reselect probes

You can select the same probes used in a previous measurement by passing the measurement ID to the `--from` flag.

```bash
globalping dns google.com from rvasVvKnj48cxNjC
> AS, IN, Mumbai, ASN:396982, Google LLC (gcp-asia-south1)
; <<>> DiG 9.16.42-Debian <<>> -t A google.com -p 53 -4 +timeout=3 +tries=2 +nocookie +nosplit +nsid
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 42607
;; flags: qr rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 512
;; QUESTION SECTION:
;google.com.                    IN      A

;; ANSWER SECTION:
google.com.             300     IN      A       142.250.199.174

;; Query time: 5 msec
;; SERVER: x.x.x.x#53(x.x.x.x)
;; WHEN: Mon Dec 18 10:01:00 UTC 2023
;; MSG SIZE  rcvd: 55
```

#### Reselect probes from measurements in the current session

Use `[@1 | first, @2 ... @-2, @-1 | last | previous]` to select the probes from previous measurements in the current session.

```bash
globalping ping google.com from USA  --latency
> NA, US, (VA), Ashburn, ASN:213230, Hetzner Online GmbH
Min: 7.314 ms
Max: 7.413 ms
Avg: 7.359 ms

globalping ping google.com from Germany  --latency
> EU, DE, Falkenstein, ASN:24940, Hetzner Online GmbH
Min: 4.87 ms
Max: 4.936 ms
Avg: 4.911 ms

globalping ping google.com from previous --latency
> EU, DE, Falkenstein, ASN:24940, Hetzner Online GmbH
Min: 4.87 ms
Max: 4.936 ms
Avg: 4.911 ms

globalping ping google.com from @-1 --latency
> EU, DE, Falkenstein, ASN:24940, Hetzner Online GmbH
Min: 4.87 ms
Max: 4.936 ms
Avg: 4.911 ms

globalping ping google.com from @-2 --latency
> NA, US, (VA), Ashburn, ASN:213230, Hetzner Online GmbH
Min: 7.314 ms
Max: 7.413 ms
Avg: 7.359 ms

globalping ping google.com from first --latency
> NA, US, (VA), Ashburn, ASN:213230, Hetzner Online GmbH
Min: 7.314 ms
Max: 7.413 ms
Avg: 7.359 ms

globalping ping google.com from @1 --latency
> NA, US, (VA), Ashburn, ASN:213230, Hetzner Online GmbH
Min: 7.314 ms
Max: 7.413 ms
Avg: 7.359 ms
```

#### Continuously ping

Use the `--infinite` flag to continuously ping a host.

```bash
globalping ping google.com from USA --infinite
...
```

#### Learn about available flags

Most commands have shared and unique flags. We recommend that you familiarize yourself with these so that you can run and automate your network tests in powerful ways.

Simply execute the command you want to learn more about with the `--help` flag:

`globalping [command] --help`

## Support and Feedback

If you are stuck or want to give us your feedback, please [open a new issue](https://github.com/jsdelivr/globalping-cli/issues).

## Development

Please refer to [CONTRIBUTING.md](CONTRIBUTING.md) for more information.
