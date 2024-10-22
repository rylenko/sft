package receiver

import "net"

type Measurable interface {
	Receive(conn net.Conn, measureChan chan<- Measure) error
}
