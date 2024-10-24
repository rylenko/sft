package sender

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
)

type Status byte

const (
	statusOK Status = iota
	statusTooLong
)

type Named struct {}

func (sender *Named) Send(conn net.Conn, file *os.File) error {
	// Try to get file stat to send file name and file size.
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("get file stat: %v", err)
	}

	// Try to send file name to the connection.
	if err := sender.sendName(conn, info.Name()); err != nil {
		return fmt.Errorf("send file name: %v", err)
	}

	// Try to send file content length to the connection.
	if err := sender.sendLen(conn, info.Size()); err != nil {
		return fmt.Errorf("send file size: %v", err)
	}

	// Try to send file content to the connection.
	if _, err := io.Copy(conn, file); err != nil {
		return fmt.Errorf("send file content: %v", err)
	}

	return nil
}

// Function to send file name length or file content length.
func (sender *Named) sendLen(conn net.Conn, length int64) error {
	// Try to send a length.
	err := binary.Write(conn, binary.LittleEndian, length)
	if err != nil {
		return fmt.Errorf("write length %d: %v", length, err)
	}

	var status Status

	// Try to receive receiver's status about length.
	if err := binary.Read(conn, binary.LittleEndian, &status); err != nil {
		return fmt.Errorf("read status: %v", err)
	}

	// Check receiver's status about length.
	if status == statusTooLong {
		return fmt.Errorf("too long (%d) for a receiver", length)
	}
	if status != statusOK {
		return fmt.Errorf("unknown name status from a receiver: %d", status)
	}

	return nil
}

func (sender *Named) sendName(conn net.Conn, name string) error {
	// Get file name bytes.
	nameBytes := []byte(name)

	// Try to send name bytes length.
	if err := sender.sendLen(conn, int64(len(nameBytes))); err != nil {
		return fmt.Errorf("send name bytes length %d: %v", len(nameBytes), err)
	}

	// Try to write name bytes to the connection.
	if _, err := conn.Write(nameBytes); err != nil {
		return fmt.Errorf("write name bytes: %v", err)
	}

	return nil
}

func NewNamed() *Named {
	return &Named{}
}
