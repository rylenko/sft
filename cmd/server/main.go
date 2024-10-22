package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rylenko/sft/internal/receiver"
	"github.com/rylenko/sft/internal/server"
)

const (
	oneTB int64 = 1 << 40
	fourKB int64 = 1 << 12
)

var (
	dir *string = flag.String("path", "./uploads", "dir to save files")
	port *int = flag.Int("port", 8000, "port to listen connection")
	contentLenLimit *int64 = flag.Int64(
		"contentLenLimit", oneTB, "file content length limit")
	nameBytesLenLimit *int64 = flag.Int64(
		"nameBytesLenLimit", fourKB, "file name bytes length limit")
	measureMinDelay *time.Duration = flag.Duration(
		"measureMinDelay", 3 * time.Second, "minimum delay of measure printing")
)

func main() {
	flag.Parse()

	// Launch client with accepted arguments.
	err := server.Launch(receiver.NewNamedSized(), *address, *path)
	if err != nil {
		log.Fatal("failed to launch client:", err)
	}
}
