# go-net-stab
Golang Network Stability Monitor (or go-net-stab) lets you easily monitor the network quality and stability to multiple endpoints.

## Quick Start
1. Download and install the Go programming language from https://golang.org/dl/
2. Clone this repository
3. Customize the configuration file (config.json) to your needs
4. Run the program with `sudo go run .`

### Prometheus Queries

```bash
# Min latency
min_over_time(ping_rtt_milliseconds[1m])

# Max latency
min_over_time(ping_rtt_milliseconds[1m])

# Avg Latency
avg_over_time(ping_rtt_milliseconds[1m])

# Jitter
max_over_time(ping_rtt_milliseconds[1m]) - min_over_time(ping_rtt_milliseconds[1m])

# Packet loss - last 5 minutes
rate(ping_lost_packet_total[$__rate_interval]) / rate(ping_sent_packet_total[$__rate_interval])
```