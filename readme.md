# Go TLS Echo Server with Rate Limiting and Prometheus Metrics

A secure TLS echo server written in Go that demonstrates advanced networking concepts, including:

## Features

- **TLS Encryption**: All client-server communications are encrypted using TLS.
- **Connection Limiting**: Enforces a maximum number of concurrent connections using channels and mutexes.
- **Prometheus Metrics**: Exposes runtime metrics (total connections and active connections) on a dedicated HTTP endpoint for monitoring.
- **Graceful Shutdown**: Handles system signals to gracefully shut down the server.
- **Rate Limiting**: Prevents the server from being overwhelmed by too many requests.
- **Basic Authentication**: Restricts access to the server using username and password.

## Prerequisites

- Go 1.16 or later
- Prometheus (if you wish to scrape the metrics)

## Installation

1. **Clone the Repository**:
   ```bash
   cd go-tls-echo-server
   ```

2. **Build the Server**:
   ```bash
   go build -o echo-server
   ```

## Usage

Run the server with default parameters:

```bash
./echo-server
```

You can also specify flags:

```bash
./echo-server -port=20080 -bufferSize=512 -certFile=server.crt -keyFile=server.key -maxConn=100 -rateLimit=10 -metricsAddr=":9090"
```

## Metrics

The server exposes Prometheus metrics on the /metrics endpoint (by default at http://localhost:9090/metrics). Monitor total connections and active connections in real time.

## Graceful Shutdown

The server listens for SIGINT and SIGTERM signals. Use Ctrl+C to gracefully shut down the server.

## Testing

Run the test cases with:

```bash
go test -v
```

This will execute the provided test for the echo functionality, ensuring the server echoes messages correctly.