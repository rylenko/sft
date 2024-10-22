package server

import (
	"fmt"
	"net"
	"time"

	"github.com/rylenko/sft/internal/receiver"
)

func Launch(
		r receiver.Measurable,
		dir string,
		port int,
		contentLenLimit,
		nameBytesLenLimit int64,
		measureMinDelay time.Duration) error {
	// Try to listen on the specified port.
	ln, err := net.Listen("tcp", ":" + string(port))
	if err != nil {
		return fmt.Errorf("listen tcp on port %d: %v", port, err)
	}
	defer ln.Close()

	for {
		// Try to accept a new connection to handle file.
		conn, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("accept new connection on port %d: %v", port, err)
		}

		// Receive a file via accepted connection, print measures in another routine.
		measureChan := make(chan receiver.Measure)
		go func(conn net.Conn, measureChan chan<- receiver.Measure) {
			defer conn.Close()
			r.Receive(conn, measureChan)
		}(conn, measureChan)
		go printMeasures(measureChan)
	}

	return nil
}

func printMeasures(measureChan <-chan receiver.Measure) {
	for measure := range measureChan {
		fmt.Printf(
			"Instant speed: %f bytes/sec; Average speed: %f bytes/sec;\n",
			measure.InstantSpeed,
			measure.AverageSpeed)
	}
}
