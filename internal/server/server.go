package server

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/rylenko/sft/internal/receiver"
)

func Launch(
		r receiver.Measurable, port int, printMeasureDelay time.Duration) error {
	// Try to listen on the specified port.
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("listen tcp on port %d: %v", port, err)
	}
	defer ln.Close()

	// Handle connections using created listener.
	for {
		// Try to accept a new connection to handle file.
		conn, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("accept new connection on port %d: %v", port, err)
		}

		// Handle accepted connection.
		go handleConnection(conn, r, printMeasureDelay)
	}

	return nil
}

func handleConnection(
		conn net.Conn, r receiver.Measurable, printMeasureDelay time.Duration) {
	defer conn.Close()

	measure := receiver.NewSyncMeasure()

	// Create channel to control measure printing goroutine.
	stopPrintMeasure := make(chan struct{})
	defer close(stopPrintMeasure)

	// Print measures in another goroutine.
	go printMeasure(
		measure, conn.RemoteAddr(), printMeasureDelay, stopPrintMeasure)

	// Try to receive a file using accepted connection.
	if err := r.Receive(conn, measure); err != nil {
		fmt.Fprintf(os.Stderr, "failed to receive a file: %v", err)
	}
}

// Prints measure for the passed address until stop.
func printMeasure(
		measure receiver.Measure,
		addr net.Addr,
		delay time.Duration,
		stop <-chan struct{}) {
	addrString := addr.String()

	// Create ticker for measure commits.
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	// Function to commit and print measure.
	commitAndPrint := func() {
		commit := measure.Commit()
		fmt.Printf(
			"[%s] Instant speed: %d bytes/sec | Average speed: %d bytes/sec;\n",
			addrString,
			int64(commit.InstantSpeed),
			int64(commit.AverageSpeed))
	}

	for {
		select {
		case <-ticker.C:
			commitAndPrint()
		case <-stop:
			commitAndPrint()
			return
		}
	}
}
