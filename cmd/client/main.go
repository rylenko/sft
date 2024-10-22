package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rylenko/sft/internal/client"
	"github.com/rylenko/sft/internal/sender"
)

const (
	missingRequiredParamsExitCode int = 1
)

var (
	address *string = flag.String("address", "", "destination address")
	path *string = flag.String("path", "", "path of file to send")
)

func main() {
	flag.Parse()

	// Validate required CLI parameters.
	if *address == "" || *path == "" {
		fmt.Fprintln(os.Stderr, "Missing required paramters\n")
		flag.Usage()
		os.Exit(missingRequiredParamsExitCode)
	}

	// Launch client with parsed CLI arguments.
	if err := client.Launch(sender.NewNamed(), *address, *path); err != nil {
		log.Fatal("failed to launch client: ", err)
	}
}
