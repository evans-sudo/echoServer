package main

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

func TestEcho(t *testing.T) {
	// Use net.Pipe to create a pair of connected endpoints.
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	// Run the echo function in a goroutine.
	go echo(serverConn)

	// Write a test message from the client side.
	testMsg := []byte("Hello, Echo Server!")
	_, err := clientConn.Write(testMsg)
	if err != nil {
		t.Fatalf("Failed to write to connection: %v", err)
	}

	// Give the echo goroutine some time to process.
	time.Sleep(100 * time.Millisecond)

	// Read the echoed message.
	echoed := make([]byte, len(testMsg))
	_, err = io.ReadFull(clientConn, echoed)
	if err != nil {
		t.Fatalf("Failed to read echoed message: %v", err)
	}

	// Verify that the echoed message matches the original message.
	if !bytes.Equal(testMsg, echoed) {
		t.Errorf("Expected echoed message %q, got %q", testMsg, echoed)
	}
}
