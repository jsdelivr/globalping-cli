<h1 align="center"> <a href="https://www.jsdelivr.com/globalping"><img width="28" alt="Globalping icon" src="https://user-images.githubusercontent.com/1834071/216975126-01529980-a87e-478c-8ab3-bf7d927a1986.png"></a> Globalping CLI </h1>

<p align="center">Access a global network of probes without leaving your console. Powered by the Globalping community!</p>

<p align="center"><img width="80%" src="https://user-images.githubusercontent.com/1834071/217010016-9da38f12-906a-47cf-adca-18017588efe5.png">
</p>

- The official command-line interface for the [Globalping](https://github.com/jsdelivr/globalping) network.
- Run networking commands from any location in the world
- Supported commands: ping, mtr, traceroute, dns resolve, HTTP
- Includes detailed timings and latency metrics with every test
- Real-time results right in your command line
- Human friendly format and output
- Cross-platform. Linux, MacOS, Windows are all supported
- Auto-updated via all automated installation methods
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

### Windows - WinGet

```shell
winget install globalping
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

### Run your first test

The following command will show a real-time result of ping from probes that have the parameters Comcast and Seattle.
You can use the + symbol as a filter to select probes more precisely.  

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

To select multiple locations you can use a comma as a delimiter. You can mix and match the location types without issues. 
We also set a limit of 4 probes to get 1 answer per location, otherwise the default limit of 1 would result in a random result from one of the four locations.

Last we use the `--latency` parameter to only get the summary latency data instead of the full raw ping output. 

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

The `--share` parameter will add a link to end of any test to view the results online. 
Note that these links expire after a few weeks depending on the type of user. GitHub Sponsors get their tests stored for longer.

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

Most commands have their own unique parameters, explore them to run and automate your network tests in powerful ways.

If you get stuck or want to provide your feedback please open a new issue.

## Development setup

Please refer to [CONTRIBUTING.md](CONTRIBUTING.md) for more information.
