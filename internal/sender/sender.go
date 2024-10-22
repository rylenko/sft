package sender

import (
	"net"
	"os"
)

type Sender interface {
	Send(conn net.Conn, file *os.File) error
}
