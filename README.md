<h1 align="center"> <a href="https://globalping.io"><img width="28" alt="Globalping icon" src="https://user-images.githubusercontent.com/1834071/216975126-01529980-a87e-478c-8ab3-bf7d927a1986.png"></a> Globalping CLI </h1>

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
- Explore additional [Globalping integrations](https://globalping.io/integrations), including our online tools, Slack app, and more

## Table of contents

<!-- TOC -->

- [Installation](#installation)
- [Get started with Globalping CLI](#get-started-with-globalping-cli)
- [Filter locations](#filter-locations)
- [Define multiple locations and basic flags](#define-multiple-locations-and-basic-flags)
- [Share results online](#share-results-online)
- [Authenticate](#authenticate)
- [Reselect probes](#reselect-probes)
- [Reselect probes from measurements in the current session](#reselect-probes-from-measurements-in-the-current-session)
- [Run continuous non-stop measurements](#run-continuous-non-stop-measurements)
- [Get TCP & TLS/SSL details](#get-tcp--tlsssl-details)
- [View your measurement history](#view-your-measurement-history)
- [Learn about available flags](#learn-about-available-flags)
<!-- TOC -->

## Installation

Install the repository and Globalping CLI using the relevant package manager command from below. This way, you can get future updates by simply running an update with your package manager.

#### Ubuntu/Debian (deb)

```shell
curl -s https://packagecloud.io/install/repositories/jsdelivr/globalping/script.deb.sh | sudo bash
apt install globalping
```

#### CentOS/Fedora/Rocky Linux/AlmaLinux (rpm)

```shell
curl -s https://packagecloud.io/install/repositories/jsdelivr/globalping/script.rpm.sh | sudo bash
dnf install globalping
```

[Manual installation instructions](https://packagecloud.io/jsdelivr/globalping/install#manual)

#### MacOS - Homebrew

```shell
brew tap jsdelivr/globalping
brew install globalping
```

#### Windows - [Chocolatey](https://community.chocolatey.org/packages/globalping)

```shell
choco install globalping
```

#### Windows - WinGet

```shell
winget install globalping
```

#### Binary installation

Every new release is compiled into binaries ready to run on most operating systems and provided as assets on GitHub. You can download and execute these binaries directly on your system.

> [!IMPORTANT]
> Opting for this installation method means you'll have to repeat this manual process to update the CLI to a newer release!

[Explore the available versions](https://github.com/jsdelivr/globalping-cli/releases).

## Get started with Globalping CLI

After installing, verify the Globalping CLI is working by running:

`globalping --help`

The result shows how to use the CLI and which commands and flags are available:

```
Usage:
  globalping [command]

Measurement Commands:
  dns           Resolve DNS records, similar to the dig command
  http          Perform a HEAD, GET, or OPTIONS request to a host
  mtr           Run a MTR test, which combines traceroute and ping
  ping          Perform a ping test
  traceroute    Run a traceroute test

Additional Commands:
  auth          Authenticate with the Globalping API
  completion    Generate the autocompletion script for the specified shell
  help          Help about any command
  history       Display the measurement history of your current session
  install-probe Join the Globalping network by running a probe
  limits        Show the current rate limits
  version       Display the version of your installed Globalping CLI

Global Measurement Flags:
  -F, --from string   specify the probe locations as a comma-separated list; you may use:
                       - names of continents, regions, countries, US states, cities, or
                      networks
                       - [@1 | first, @2 ... @-2, @-1 | last | previous] to run with the probes
                      from previous measurements in this session
                       - an ID of a previous measurement to run with its probes
                       (default "world")
  -4, --ipv4          resolve names to IPv4 addresses
  -6, --ipv6          resolve names to IPv6 addresses
  -J, --json          output results in JSON format (default false)
      --latency       output only the latency stats; applicable only to dns, http, and ping
                      commands (default false)
  -L, --limit int     define the number of probes to use (default 1)
      --share         print a link at the end of the results to visualize them online (default
                      false)

Global Flags:
  -C, --ci     disable real-time terminal updates and colors, suitable for CI and scripting
               (default false)
  -h, --help   help for globalping

Use "globalping [command] --help" for more information about a command.
```

Globalping relies on a community-hosted probe network, enabling you to run network tests from any location with an active probe. The following examples show you through some tests, exploring how to define locations, set limits, and use some command flags.

#### Filter locations

For example, if you want to run ping from a probe in Seattle that is also part of the Comcast network, run the following:

```bash
globalping ping google.com from Comcast+Seattle
> Seattle (WA), US, NA, Comcast Cable Communications, LLC (AS33650)
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
> Seoul, KR, AS, Amazon.com, Inc. (AS16509) (aws-ap-northeast-2)
Min: 33.163 ms
Max: 33.256 ms
Avg: 33.22 ms

> Frankfurt, DE, EU, DE, OVH SAS (AS16276)
Min: 1.221 ms
Max: 1.291 ms
Avg: 1.264 ms

> Chicago (IL), US, NA, Cogent Communications (AS174)
Min: 112.405 ms
Max: 112.686 ms
Avg: 112.528 ms

> Dallas (TX), US, NA, Catalyst Host LLC (AS393336)
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
> Mumbai, IN, AS, Google LLC (AS396982) (gcp-asia-south1)
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
> View the results online: https://globalping.io?measurement=xrfXUEAOGfzwfHFz
```

## Authenticate

Register a free [Globalping account](https://dash.globalping.io/), then sign in with the CLI to increase the number of tests you can run every hour. GitHub sponsors and users hosting probes receive additional [credits](https://globalping.io/credits), which allow them to run even more tests.

#### Interactive browser-based flow

By default, the CLI will open your web browser, where you simpy confirm the sign in by clicking a button.

```bash
globalping auth login
Please visit the following URL to authenticate:
https://auth.globalping.io/oauth/authorize...
```

#### Providing a token manually

If you can't use the interactive flow, you can create a token in the [Dashboard](https://dash.globalping.io/tokens) and provide it via `stdin`.

```bash
globalping auth login --with-token
Please enter your token:
```

Alternatively, you may set the environment variable `GLOBALPING_TOKEN`, which will be used automatically when present.

## Advanced features

After learning the basics, you may also be interested in these extra features, which provide additional control over your measurements.

#### Reselect probes

You can select the same probes used in a previous measurement by passing the measurement ID to the `--from` flag.

```bash
globalping dns google.com from rvasVvKnj48cxNjC
> Mumbai, IN, AS, Google LLC (AS396982) (gcp-asia-south1)
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
globalping ping google.com from USA --latency
> Ashburn (VA), US, NA, Hetzner Online GmbH (AS213230)
Min: 7.314 ms
Max: 7.413 ms
Avg: 7.359 ms

globalping ping google.com from Germany --latency
> Falkenstein, DE, EU, Hetzner Online GmbH (AS24940)
Min: 4.87 ms
Max: 4.936 ms
Avg: 4.911 ms

globalping ping google.com from previous --latency
> Falkenstein, DE, EU, Hetzner Online GmbH (AS24940)
Min: 4.87 ms
Max: 4.936 ms
Avg: 4.911 ms

globalping ping google.com from @-1 --latency
> Falkenstein, DE, EU, Hetzner Online GmbH (AS24940)
Min: 4.87 ms
Max: 4.936 ms
Avg: 4.911 ms

globalping ping google.com from @-2 --latency
> Ashburn (VA), US, NA, Hetzner Online GmbH (AS213230)
Min: 7.314 ms
Max: 7.413 ms
Avg: 7.359 ms

globalping ping google.com from first --latency
> Ashburn (VA), US, NA, Hetzner Online GmbH (AS213230)
Min: 7.314 ms
Max: 7.413 ms
Avg: 7.359 ms

globalping ping google.com from @1 --latency
> Ashburn (VA), US, NA, Hetzner Online GmbH (AS213230)
Min: 7.314 ms
Max: 7.413 ms
Avg: 7.359 ms
```

#### Run continuous non-stop measurements

> [!IMPORTANT]
> Currently, this feature is limited to the ping command.

Use the `--infinite` flag to continuously ping a host, just like on Linux or MacOS. Although it appears as a single measurement, the Globalping API combines multiple measurements from the same probes into one output. As a result, the test will stop once you run out of credits.

Here's an example of running an infinite ping from a single probe:

```bash
globalping ping cdn.jsdelivr.net from Europe --infinite
> London, GB, EU, Psychz Networks (AS40676)
PING cdn.jsdelivr.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=59 time=0.54 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=59 time=0.42 ms
^C
```

If you define multiple probes to perform your infinite ping, the CLI output switches to summary mode, letting you compare the result data collected from all probes:

```bash
globalping ping cdn.jsdelivr.net from Europe --limit 5 --infinite
Location                                               | Sent |    Loss |     Last |      Min |      Avg |      Max
London, GB, EU, OVH SAS (AS16276)                      |   22 |   0.00% |  3.33 ms |  3.07 ms |  3.20 ms |  3.33 ms
Falkenstein, DE, EU, Hetzner Online GmbH (AS24940)     |   22 |   0.00% |  5.41 ms |  5.30 ms |  5.78 ms |  13.1 ms
Vienna, AT, EU, EDIS GmbH (AS57169)                    |   22 |   0.00% |  0.47 ms |  0.46 ms |  0.56 ms |  0.88 ms
Stockholm, SE, EU, The Constant Company, LLC (AS20473) |   22 |   0.00% |  1.03 ms |  0.83 ms |  1.15 ms |  4.66 ms
Madrid, ES, EU, EDGOO NETWORKS LLC (AS47787)           |   22 |   0.00% |  0.24 ms |  0.13 ms |  0.26 ms |  0.42 ms
^C
```

> [!TIP]
> Stop the infinite ping by pressing CTRL+C on your keyboard.

#### Get TCP & TLS/SSL details

Use the `--full` option when running an `http` command to include the TCP and TLS/SSL details in the output.

```bash
globalping http jsdelivr.com --full
> Kansas City (MO), US, NA, IONOS SE (AS8560), u-zGato
Resolved address: 172.67.208.113

TLSv1.3/TLS_AES_256_GCM_SHA384
Subject: jsdelivr.com; DNS:jsdelivr.com, DNS:*.jsdelivr.com
Issuer: WE1; Google Trust Services; US
Validity: 2025-02-02T21:13:14Z; 2025-05-03T22:10:40Z
Serial number: 57:E1:7F:F0:C7:25:0A:A5:0D:3B:73:06:FD:C0:72:D4
Fingerprint: 73:A9:13:48:38:E7:42:E6:97:72:5A:87:D8:95:20:BE:3A:32:5F:27:78:40:CA:9C:BE:EC:F8:7F:75:11:43:79
Key type: EC256

HTTP/1.1 301
...
```

> [!TIP]
> Use `globalping http jsdelivr.com --full --method head` to omit the response body.

#### View your measurement history

You can view the history of your current session's measurements by running the `history` command.

```bash
globalping history
1 | 2024-03-27 11:56:46 | ping google.com
> https://globalping.io?measurement=itcR65tYCqbouXib
- | 2024-03-27 11:57:01 | dns google.com from last
> https://globalping.io?measurement=kWc5UBK9A6G4RUYM
2 | 2024-03-27 11:57:20 | traceroute google.com from New York --limit 2
> https://globalping.io?measurement=Yz7A1UifUonZsC3C
3 | 2024-03-27 11:57:37 | mtr google.com from New York --limit 2
> https://globalping.io?measurement=SX1NBgfDKiabM1vZ
4 | 2024-03-27 11:57:52 | http google.com from London,Belgium --limit 2 --method get --ci
> https://globalping.io?measurement=eclwFSYX0zgU10Cs
```

> [!TIP]
> Use this command to get the measurement IDs needed to run a new measurement, which [reuses the probes](#reselect-probes) from a previous one.

#### Learn about available flags

Most commands have shared and unique flags. We recommend that you familiarize yourself with these so that you can run and automate your network tests in powerful ways.

Simply execute the command you want to learn more about with the `--help` flag:

`globalping [command] --help`

## Support and Feedback

If you are stuck or want to give us your feedback, please [open a new issue](https://github.com/jsdelivr/globalping-cli/issues).

## Development

Please refer to [CONTRIBUTING.md](CONTRIBUTING.md) for more information.

## Globalping API package

For the official Globalping API package, visit [https://github.com/jsdelivr/globalping-go](https://github.com/jsdelivr/globalping-go).
