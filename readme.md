# go-net-stab
Golang Network Stability Monitor (or go-net-stab) aims to make it incredibly simple to monitor the stability of an enterprise network. It works by heavily utilizing the net/icmp package and goroutines to ping a list of hosts and report back the results which can then be scraped by Prometheus.

## Quickstart
```py
git clone https://github.com/bwoff11/go-net-stab.git
cd go-net-stab
# Edit the config.yml file to include the hosts you want to monitor
go run .
```

## How it Works
Each component of the program is divided into an internal package. The main package's only responsibility is to start each package's primary function.

### Config

Responsible for reading the config.YAML file and parsing it into a struct. The config file is used to specify the hosts to ping and the interval at which to ping them, among other things.

### Listener

The Listener is separate from the pingers, given that having multiple threads trying to read and parse layers three packets causes problems. Once it reads the packet, it parses it and hands it off to the registry to process.

### Pingers

Pingers are generated from the hosts in the config file and are responsible for transmitting the pings on the interval specified in the config.

### Registry

Probably the most significant package, the registry is responsible for ingesting transmitted packet data, receiving incoming packets from the Listener, and marking the state of each packet (e.g. completed/lost).

### Reporting

The reporting package is responsible for exposing the metrics for the other packages to submit data to, as well as hosting a simple web server for Prometheus to scrape the metrics from.