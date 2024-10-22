package receiver

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"unsafe"
)

type Status byte

const (
	statusOK Status = iota
	statusTooLong
)

const (
	chunkReadLen int64 = 2048
)

type NamedLimited struct {
	dirPath string
	nameLenLimit int64
	contentLenLimit int64
}

func (receiver *NamedLimited) Receive(conn net.Conn, measure Measure) error {
	// Try to receive file name from the connection.
	name, err := receiver.receiveName(conn, measure)
	if err != nil {
		return fmt.Errorf("receive file name: %v", err)
	}

	// Try to receive file content length from the connection.
	contentLen, err := receiver.receiveLen(
		conn, receiver.contentLenLimit, measure)
	if err != nil {
		return fmt.Errorf("receive file content length: %v", err)
	}

	// Try to receive file content.
	err = receiver.receiveContent(conn, contentLen, name, measure)
	if err != nil {
		return fmt.Errorf("receive content with length %d: %v", contentLen, err)
	}

	return nil
}

func (receiver *NamedLimited) createFile(
		name string, length int64) (*os.File, error) {
	// Build the full path to the file.
	path := filepath.Join(receiver.dirPath, name)

	// Create a file in specified directory using accepted name.
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create file %s: %v", path, err)
	}

	// Try to resize file using accepted content length.
	if err := file.Truncate(length); err != nil {
		return nil, fmt.Errorf("truncate file %s with %d: %v", path, length, err)
	}

	return file, nil
}

func (receiver *NamedLimited) receiveContent(
		conn net.Conn, contentLen int64, name string, measure Measure) error {
	// Try to create a file using accepted content length and name.
	file, err := receiver.createFile(name, contentLen)
	if err != nil {
		return fmt.Errorf(
			"create a file with name %s and length %d: %v", name, contentLen, err)
	}

	// Read content bytes from connection and add readed lengths to measure.
	for contentLen > 0 {
		// Try to copy chunk of bytes from connection to the created file.
		chunkReadedLen, err := io.CopyN(file, conn, chunkReadLen)
		if err != nil && err != io.EOF {
			return fmt.Errorf("copy %d bytes to file %s: %v", chunkReadLen, name, err)
		}
		contentLen -= chunkReadedLen

		// Add readed chunk length to the measure to commit it in the future.
		measure.AddReadedLen(chunkReadedLen)
	}

	return nil
}

func (receiver *NamedLimited) receiveName(
		conn net.Conn, measure Measure) (string, error) {
	// Try to receive name bytes length.
	nameLen, err := receiver.receiveLen(conn, receiver.nameLenLimit, measure)
	if err != nil {
		return "", fmt.Errorf("failed to receive bytes length: %v", err)
	}

	// Try to receive a name bytes.
	nameBytes := make([]byte, nameLen)
	if _, err := io.ReadFull(conn, nameBytes); err != nil {
		return "", fmt.Errorf("failed to receive %d bytes: %v", nameLen, err)
	}
	// Add readed name length to measure.
	measure.AddReadedLen(nameLen)

	return string(nameBytes), nil
}

func (receiver *NamedLimited) receiveLen(
		conn net.Conn, limit int64, measure Measure) (int64, error) {
	// Try to receive a length.
	var length int64
	if err := binary.Read(conn, binary.LittleEndian, &length); err != nil {
		return 0, fmt.Errorf("failed to read: %v", err)
	}
	// Add readed length variable size to measure.
	measure.AddReadedLen(int64(unsafe.Sizeof(length)))

	// Validate readed length using accepted limit.
	if length > limit {
		// Try to write too long status to the connection.
		err := binary.Write(conn, binary.LittleEndian, statusTooLong)
		if err != nil {
			return 0, fmt.Errorf("failed to write too long status: %v", err)
		}

		return 0, fmt.Errorf(
			"too long (%d), limit is (%d)", length, receiver.nameLenLimit)
	}

	// Try to write ok status to the connection.
	if err := binary.Write(conn, binary.LittleEndian, statusOK); err != nil {
		return 0, fmt.Errorf("failed to write ok status: %v", err)
	}

	return length, nil
}

func NewNamedLimited(
		dirPath string, nameLenLimit, contentLenLimit int64) *NamedLimited {
	return &NamedLimited{
		dirPath: dirPath,
		contentLenLimit: contentLenLimit,
		nameLenLimit: nameLenLimit,
	}
}
