package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/go-getter"
)

var version = "DEV"

func main() {
	modeRaw := flag.String("mode", "any", "get mode (any, file, dir)")
	verPtr := flag.Bool("version", false, "print version")
	flag.Parse()

	if *verPtr {
		fmt.Printf("version: %s\n", version)
		os.Exit(0)
	}
	args := flag.Args()
	if len(args) < 2 {
		log.Fatalf("Expected two args: URL and dst")
	}

	// Get the mode
	var mode getter.ClientMode
	switch *modeRaw {
	case "any":
		mode = getter.ClientModeAny
	case "file":
		mode = getter.ClientModeFile
	case "dir":
		mode = getter.ClientModeDir
	default:
		log.Fatalf("Invalid client mode, must be 'any', 'file', or 'dir': %s", *modeRaw)
	}

	// Get the pwd
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting wd: %s", err)
	}

	// Build the client
	client := &getter.Client{
		Src:  args[0],
		Dst:  args[1],
		Pwd:  pwd,
		Mode: mode,
	}

	if err := client.Get(); err != nil {
		log.Fatalf("Error downloading: %s", err)
	}

	log.Println("Success!")
}
