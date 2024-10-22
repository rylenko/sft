package main

import (
	"flag"
	"log"
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
	nameLenLimit *int64 = flag.Int64(
		"nameLenLimit", fourKB, "file name bytes length limit")
	contentLenLimit *int64 = flag.Int64(
		"contentLenLimit", oneTB, "file content length limit")
	printMeasureDelay *time.Duration = flag.Duration(
		"printMeasureDelay", 3 * time.Second, "delay of measure printing")
)

func main() {
	flag.Parse()

	// Launch client with accepted arguments.
	r := receiver.NewNamedLimited(*dir, *nameLenLimit, *contentLenLimit)
	if err := server.Launch(r, *port, *printMeasureDelay); err != nil {
		log.Fatal("failed to launch client: ", err)
	}
}
