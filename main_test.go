package main


import (
    "crypto/tls"
    "io/ioutil"
    "net/http"
    "testing"
    "time"
)

func TestEchoServer(t *testing.T) {
    go main() // Start the echo server in a goroutine

    time.Sleep(2 * time.Second) // Give the server a moment to start

    // Test the echo functionality
    conn, err := tls.Dial("tcp", "localhost:20080", &tls.Config{InsecureSkipVerify: true})
    if err != nil {
        t.Fatalf("Failed to connect to server: %v", err)
    }
    defer conn.Close()

    message := "Hello, Echo Server!"
    _, err = conn.Write([]byte(message))
    if err != nil {
        t.Fatalf("Failed to send message: %v", err)
    }

    buffer := make([]byte, len(message))
    _, err = conn.Read(buffer)
    if err != nil {
        t.Fatalf("Failed to read message: %v", err)
    }

    if string(buffer) != message {
        t.Fatalf("Expected %q but got %q", message, string(buffer))
    }

    // Test the metrics endpoint
    resp, err := http.Get("http://localhost:9090/metrics")
    if err != nil {
        t.Fatalf("Failed to get metrics: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Fatalf("Expected status code 200 but got %d", resp.StatusCode)
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        t.Fatalf("Failed to read response body: %v", err)
    }

    if len(body) == 0 {
        t.Fatalf("Expected non-empty response body")
    }
}
