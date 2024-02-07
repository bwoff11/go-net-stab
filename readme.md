# Go-Net-Stab: A Network Stability Monitor

`go-net-stab` is a Network Stability Monitor written in Go, designed to periodically send ICMP echo requests (pings) to specified endpoints and monitor the stability of network connections. The responses are used to calculate round-trip time (RTT) and packet loss, which are then exposed as Prometheus metrics.

## Features

- **ICMP Echo Requests:** Pings are sent to specified endpoints and responses are listened for.
- **RTT Measurement:** The time taken between sending a ping and receiving a response is calculated.
- **Packet Loss Detection:** Pings that haven't received a response within a specified timeout period are tracked.
- **Prometheus Metrics:** RTT and packet loss metrics are exposed in a format that can be scraped by Prometheus.
- **Configuration via File:** Endpoints to be monitored and other configuration parameters can be defined via a config file.

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

