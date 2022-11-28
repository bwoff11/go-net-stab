# go-net-stab
Golang Network Stability Monitor (or go-net-stab) lets you easily monitor the network quality and stability to multiple endpoints.


### Prometheus Queries

```bash
# Min latency - last 5 minutes
min_over_time(ping_rtt_milliseconds[5m])

# Max latency - last 5 minutes
max_over_time(ping_rtt_milliseconds[5m])

# Average latency - last 5 minutes
avg_over_time(ping_rtt_milliseconds[5m])

# Jitter - last 5 minutes
abs(min_over_time(ping_rtt_milliseconds[1m]) - max_over_time(ping_rtt_milliseconds[1m]))

# Packet loss - last 5 minutes
rate(ping_lost_packet_total{}[$__interval]) / rate(ping_sent_packet_total{}[$__interval])
```