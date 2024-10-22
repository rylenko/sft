package receiver

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Status byte

const (
	statusOK Status = iota
	statusTooLong
)

const (
	chunkReadLen int64 = 2048
)

type NamedSized struct {
	dirPath string
	nameBytesLenLimit int64
	contentLenLimit int64
	measureMinDelay time.Duration
}

func (receiver *NamedSized) Receive(
		conn net.Conn, measureChan chan<- Measure) error {
	// Try to receive file name from the connection.
	name, err := receiver.receiveName(conn)
	if err != nil {
		return fmt.Errorf("receive file name: %v", err)
	}

	// Try to receive file content length from the connection.
	contentLen, err := receiver.receiveLen(conn, receiver.contentLenLimit)
	if err != nil {
		return fmt.Errorf("receive file content length: %v", err)
	}

	// Try to receive file content.
	err = receiver.receiveContent(conn, contentLen, name, measureChan)
	if err != nil {
		return fmt.Errorf("receive content with length %d: %v", contentLen, err)
	}

	return nil
}

func (receiver *NamedSized) createFile(
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

func (receiver *NamedSized) receiveContent(
		conn net.Conn,
		contentLen int64,
		name string,
		measureChan chan<- Measure) error {
	// Try to create a file using accepted content length and name.
	file, err := receiver.createFile(name, contentLen)
	if err != nil {
		return fmt.Errorf(
			"create a file with name %s and length %d: %v", name, contentLen, err)
	}

	instantReadedLen := int64(0)
	totalReadedLen := int64(0)

	startTime := time.Now()
	lastMeasureTime := startTime

	for totalReadedLen < contentLen {
		// Try to copy chunk of bytes from connection to the created file.
		chunkReadedLen, err := io.CopyN(file, conn, chunkReadLen)
		if err != nil {
			return fmt.Errorf("copy %d bytes to file %s: %v", chunkReadLen, name, err)
		}

		// Add readed chunk length to instant and total lengths.
		instantReadedLen += chunkReadedLen
		totalReadedLen += chunkReadedLen

		currentTime := time.Now()
		elapsedTime := currentTime.Sub(lastMeasureTime)

		if elapsedTime < receiver.measureMinDelay && totalReadedLen < contentLen {
			continue
		}

		measure := NewMeasure(0, 0)
		if elapsedTime.Seconds() > 0 {
			measure.InstantSpeed = float64(instantReadedLen) / elapsedTime.Seconds()
		}

		totalElapsedTime := currentTime.Sub(startTime)
		if totalElapsedTime.Seconds() > 0 {
			measure.AverageSpeed = float64(totalReadedLen) / totalElapsedTime.Seconds()
		}

		// Send measure to channel without block.
		select {
		case measureChan <- measure:
		default:
		}

		lastMeasureTime = currentTime

		// Zeroize instant length for the next measure accumulation.
		instantReadedLen = 0
	}

	close(measureChan)
	return nil
}

func (receiver *NamedSized) receiveName(conn net.Conn) (string, error) {
	// Try to receive name bytes length.
	nameBytesLen, err := receiver.receiveLen(conn, receiver.nameBytesLenLimit)
	if err != nil {
		return "", fmt.Errorf("failed to receive bytes length: %v", err)
	}

	// Try to receive a name bytes.
	nameBytes := make([]byte, nameBytesLen)
	if _, err := io.ReadFull(conn, nameBytes); err != nil {
		return "", fmt.Errorf("failed to receive %d bytes: %v", nameBytesLen, err)
	}

	return string(nameBytes), nil
}

func (receiver *NamedSized) receiveLen(
		conn net.Conn, limit int64) (int64, error) {
	// Try to receive a length.
	var length int64
	if err := binary.Read(conn, binary.LittleEndian, &length); err != nil {
		return 0, fmt.Errorf("failed to read: %v", err)
	}

	// Validate readed length using accepted limit.
	if length > limit {
		// Try to write too long status to the connection.
		err := binary.Write(conn, binary.LittleEndian, statusTooLong)
		if err != nil {
			return 0, fmt.Errorf("failed to write too long status: %v", err)
		}

		return 0, fmt.Errorf(
			"too long (%d), limit is (%d)", length, receiver.nameBytesLenLimit)
	}

	// Try to write ok status to the connection.
	if err := binary.Write(conn, binary.LittleEndian, statusOK); err != nil {
		return 0, fmt.Errorf("failed to write ok status: %v", err)
	}

	return length, nil
}

func NewNamedSized(
		dirPath string,
		nameBytesLenLimit,
		contentLenLimit int64,
		measureMinDelay time.Duration) *NamedSized {
	return &NamedSized{
		dirPath: dirPath,
		contentLenLimit: contentLenLimit,
		nameBytesLenLimit: nameBytesLenLimit,
		measureMinDelay: measureMinDelay,
	}
}
