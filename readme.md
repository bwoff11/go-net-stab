# go-net-stab

The `go-net-stab` is a Network Stability Monitor written in Go. It monitors the stability of network connections by periodically sending ICMP echo requests, also known as "pings", to specified endpoints. The responses are used to calculate round-trip time (RTT) and packet loss, which are exposed as Prometheus metrics. 

## Features

- **ICMP Echo Requests (Pings):** Sends pings to the specified endpoints and listens for responses.
- **Round-Trip Time (RTT) Measurements:** Calculates the time between sending a ping and receiving a response.
- **Packet Loss Detection:** Keeps track of pings that have not received a response within a specified timeout period.
- **Prometheus Metrics:** Exposes RTT and packet loss metrics in a format that can be scraped by Prometheus.
- **Configuration via File:** Allows defining the endpoints to be monitored and other configuration parameters via a config file.

# Performance Metrics

I'm proud to highlight the improved performance of the latest version of this application. 

## Garbage Collection Impact

The application currently exerts a very low impact on the Go garbage collector. The garbage collection pause duration (`go_gc_duration_seconds`) has reported 0 for all quantiles, ensuring an uninterrupted service. 

## Goroutines Efficiency

The count of Goroutines (`go_goroutines`) is efficiently managed and remains stable at 12, showcasing excellent handling of concurrency.

## Memory Usage Management

Memory usage in this application is managed effectively. The total allocated memory (`go_memstats_alloc_bytes_total`) is around 17 MB, while the memory still in use (`go_memstats_alloc_bytes`) is approximately 7.8 MB. With the total number of mallocs (`go_memstats_mallocs_total`) at 30487 and the total number of frees (`go_memstats_frees_total`) close at 25342, a balanced memory allocation and deallocation process is evident.

## Heap Memory Efficiency

The application uses heap memory efficiently. The number of heap bytes allocated and still in use (`go_memstats_heap_alloc_bytes`) is only 7.8 MB. Additionally, the number of heap bytes waiting to be used (`go_memstats_heap_idle_bytes`) is around 5.6 MB, suggesting a good buffer for future allocations.

## Object Management

The program is currently managing 5145 objects (`go_memstats_heap_objects`), showing robustness and capacity to handle a substantial amount of objects.

## CPU Efficiency

The total CPU time spent (`process_cpu_seconds_total`) is incredibly low at 0.015625 seconds, demonstrating the application's CPU efficiency.

## Memory Footprint

The resident memory size (`process_resident_memory_bytes`) of the application is approximately 22.4 MB, showcasing a low memory footprint.

## Installation

1. Download and install the Go programming language from https://golang.org/dl/
2. Clone this repository.
3. Customize the configuration file (config.json) according to your needs.
4. Run the program with `go run .` from the root directory of the project.

## Configuration

The configuration file is located at one of the following locations (in order):
- `/etc/go-net-stab/`
- `$HOME/.config/go-net-stab/`
- `$HOME/.go-net-stab`
- `.` (the current directory)

The configuration file (config.json) has the following structure:

```json
{
  "Interval": 1000,
  "Timeout": 10000,
  "Port": "3009",
  "Localhost": "localhost",
  "Endpoints": [
    {
      "Hostname": "example1.com",
      "Address": "93.184.216.34",
      "Location": "Example City"
    },
    {
      "Hostname": "example2.com",
      "Address": "203.0.113.195",
      "Location": "Another Example City"
    }
  ]
}
```

Where:
- `Interval` is the time in milliseconds between each ping.
- `Timeout` is the time in milliseconds after which a ping is considered lost if no response has been received.
- `Port` is the port where the Prometheus metrics are exposed.
- `Localhost` is the hostname of the local machine.
- `Endpoints` is a list of the endpoints to ping. Each endpoint has a `Hostname`, `Address` (IP address), and `Location` (for informational purposes).

## Prometheus Metrics

The following metrics are exposed on the `/metrics` endpoint:
- `ping_sent_packet_total`: The total number of ICMP echo requests sent.
- `ping_lost_packet_total`: The total number of ICMP echo requests that did not receive a response within the timeout period.
- `ping_rtt_milliseconds`: The round-trip time of the ICMP echo requests.

Each of these metrics is labeled with the source hostname, destination hostname, destination address, and destination location.

## Usage

After the installation and configuration are complete, run the program with `go run .` from the root directory of the project. The Prometheus metrics will be available at `http://localhost:3009/metrics` (replace `3009` with the port specified in the configuration file).

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

## Contribution

Feel free to contribute to this project by opening issues or submitting pull requests. All contributions are welcome.
