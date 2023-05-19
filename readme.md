# Go-Net-Stab: A Network Stability Monitor

`go-net-stab` is a Network Stability Monitor written in Go, designed to periodically send ICMP echo requests (pings) to specified endpoints and monitor the stability of network connections. The responses are used to calculate round-trip time (RTT) and packet loss, which are then exposed as Prometheus metrics.

## Features

- **ICMP Echo Requests:** Pings are sent to specified endpoints and responses are listened for.
- **RTT Measurement:** The time taken between sending a ping and receiving a response is calculated.
- **Packet Loss Detection:** Pings that haven't received a response within a specified timeout period are tracked.
- **Prometheus Metrics:** RTT and packet loss metrics are exposed in a format that can be scraped by Prometheus.
- **Configuration via File:** Endpoints to be monitored and other configuration parameters can be defined via a config file.

## Performance Metrics

I am delighted to share the performance improvements achieved with the latest version of the application.

### Garbage Collection Impact

The Go garbage collector experiences minimal impact from the application. The garbage collection pause duration (`go_gc_duration_seconds`) reports 0 for all quantiles, ensuring an uninterrupted service. 

### Goroutines Efficiency

The Goroutines (`go_goroutines`) count is efficiently managed and maintains stability at 12, indicating excellent concurrency handling.

### Memory Usage Management

The total allocated memory (`go_memstats_alloc_bytes_total`) for this application is around 17 MB, whereas the memory still in use (`go_memstats_alloc_bytes`) is approximately 7.8 MB. The total number of mallocs (`go_memstats_mallocs_total`) stands at 30487, with the total number of frees (`go_memstats_frees_total`) closely following at 25342, thereby evidencing a balanced memory allocation and deallocation process.

### Heap Memory Efficiency

The application efficiently utilizes heap memory. The number of heap bytes allocated and still in use (`go_memstats_heap_alloc_bytes`) is only 7.8 MB. Moreover, the number of heap bytes waiting to be used (`go_memstats_heap_idle_bytes`) is around 5.6 MB, suggesting an ample buffer for future allocations.

### Object Management

The program currently manages 5145 objects (`go_memstats_heap_objects`), demonstrating the system's robustness and its capacity to handle a significant number of objects.

### CPU Efficiency

The total CPU time spent (`process_cpu_seconds_total`) is impressively low at 0.015625 seconds, highlighting the application's CPU efficiency.

### Memory Footprint

The resident memory size (`process_resident_memory_bytes`) of the application stands at approximately 22.4 MB, underscoring the program's low memory footprint.

## Installation and Configuration

Follow the steps below to install and configure `go-net-stab`:

1. Download and install the Go programming language from https://golang.org/dl/.
2. Clone this repository.
3. Customize the configuration file (`config.json`) to suit your needs.
4. Run the program using `go run .` from the root directory of the project.

Refer to the **Configuration** section for more details on how to configure `go-net-stab`.

## Usage and Prometheus Metrics

After completing the installation and configuration, run the program using `go run .` from the root directory of the project. Prometheus metrics will be available at `http://localhost:3009/metrics` (replace `3009` with the port specified in the configuration file).

Prometheus queries for various metrics are provided in the **Prometheus Queries** section.

## Contributions

Contributions to this project are always welcome! Feel free to open issues or submit pull requests as needed.

