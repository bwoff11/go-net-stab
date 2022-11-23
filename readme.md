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
To be completed