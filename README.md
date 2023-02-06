<h1 align="center"> <a href="https://www.jsdelivr.com/globalping"><img width="28" alt="Globalping icon" src="https://user-images.githubusercontent.com/1834071/216975126-01529980-a87e-478c-8ab3-bf7d927a1986.png"></a> Globalping CLI </h1>

<p align="center">Access a global network of probes without leaving your console. Powered by the Globalping community!</p>

<p align="center"><img height="350px" src="https://user-images.githubusercontent.com/1834071/217010016-9da38f12-906a-47cf-adca-18017588efe5.png">
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

```
curl -s https://packagecloud.io/install/repositories/jsdelivr/globalping/script.deb.sh | sudo bash
apt install globalping
```

### CentOS/Fedora/Rocky Linux/AlmaLinux (rpm)

```
curl -s https://packagecloud.io/install/repositories/jsdelivr/globalping/script.rpm.sh | sudo bash
dnf install globalping
```

[Manual installation instructions](https://packagecloud.io/jsdelivr/globalping/install#manual)

### MacOS - Homebrew

```
brew tap jsdelivr/globalping-cli
brew install globalping
```

### Windows - Chocolatey

[coming soon]

### Windows - Winget

[coming soon]

### Binary installation 

Every new release is compiled into binaries ready to run on most operating systems and attached as assets on GitHub. You can download and run the binaries directly on your system, but note that you will have to repeat this process for every new release. [Explore the available versions](https://github.com/jsdelivr/globalping-cli/releases).


## Getting Started with Globalping CLI

Once the Globalping CLI is installed, you can verify that it is working by running:

```bash
globalping --help
```


