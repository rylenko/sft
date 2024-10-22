package client

import (
	"fmt"
	"net"
	"os"

	"github.com/rylenko/sft/internal/sender"
)

func Launch(s sender.Sender, address, filePath string) error {
	// Try to open file using its path.
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file %s: %v", filePath, err)
	}
	defer file.Close()

	// Try to connect to server using its address.
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("connect via tcp to server %s: %v", address, err)
	}
	defer conn.Close()

	// Try to send file using opened file and established connection.
	if err := s.Send(conn, file); err != nil {
		return fmt.Errorf("send file %s to server %s: %v", filePath, address, err)
	}

	return nil
}
