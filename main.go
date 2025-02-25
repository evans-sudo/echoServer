package main

import (
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/juju/ratelimit"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	port        = flag.String("port", "20080", "Port to listen on")
	bufferSize  = flag.Int("bufferSize", 512, "Size of the buffer")
	certFile    = flag.String("certFile", "server.crt", "TLS certificate file")
	keyFile     = flag.String("keyFile", "server.key", "TLS key file")
	maxConn     = flag.Int("maxConn", 100, "Maximum number of concurrent connections")
	rateLimit   = flag.Int64("rateLimit", 10, "Rate limit in requests per second")
	username    = flag.String("username", "admin", "Username for basic authentication")
	password    = flag.String("password", "password", "Password for basic authentication")
	metricsAddr = flag.String("metricsAddr", ":9090", "Address for metrics endpoint")
)

var (
	activeConnections int
	connMutex         sync.Mutex
	connLimit         = make(chan struct{}, *maxConn)
	rateLimiter       = ratelimit.NewBucketWithRate(float64(*rateLimit), int64(*rateLimit))
)

var (
	totalConnections = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "total_connections",
		Help: "Total number of connections",
	})
	activeConnectionsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "active_connections",
		Help: "Number of active connections",
	})
)

func init() {
	prometheus.MustRegister(totalConnections)
	prometheus.MustRegister(activeConnectionsGauge)
}

func echo(conn net.Conn) {
	defer conn.Close()
	connMutex.Lock()
	activeConnections++
	activeConnectionsGauge.Inc()
	connMutex.Unlock()
	defer func() {
		connMutex.Lock()
		activeConnections--
		activeConnectionsGauge.Dec()
		connMutex.Unlock()
	}()

	b := make([]byte, *bufferSize)
	for {
		size, err := conn.Read(b[0:])
		if err == io.EOF {
			log.Println("Client disconnected")
			break
		}
		if err != nil {
			log.Println("Unexpected error:", err)
			break
		}
		log.Printf("Received %d bytes: %s\n", size, string(b))
		log.Println("Writing data")
		if _, err := conn.Write(b[0:size]); err != nil {
			log.Fatalln("Unable to write data:", err)
		}
	}
}

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalln("Unable to bind to port:", err)
	}
	defer listener.Close()

	tlsConfig := &tls.Config{
		Certificates: make([]tls.Certificate, 1),
	}
	tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		log.Fatalln("Unable to load TLS certificate and key:", err)
	}

	tlsListener := tls.NewListener(listener, tlsConfig)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Println("Starting metrics server on", *metricsAddr)
		if err := http.ListenAndServe(*metricsAddr, nil); err != nil {
			log.Fatalln("Unable to start metrics server:", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		listener.Close()
		os.Exit(0)
	}()

	log.Println("Listening on 0.0.0.0:" + *port)
	for {
		conn, err := tlsListener.Accept()
		if err != nil {
			log.Println("Unable to accept connection:", err)
			continue
		}

		connMutex.Lock()
		if activeConnections >= *maxConn {
			connMutex.Unlock()
			conn.Close()
			log.Println("Connection limit reached, rejecting connection")
			continue
		}
		connMutex.Unlock()

		totalConnections.Inc()
		connLimit <- struct{}{}
		go func() {
			defer func() { <-connLimit }()
			echo(conn)
		}()
	}
}
