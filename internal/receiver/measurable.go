package receiver

import "net"

type Measurable interface {
	Receive(conn net.Conn, measure Measure) error
}
