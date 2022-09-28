package main

import (
	"cse224/proj4/pkg/surfstore"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

// Arguments
const ARG_COUNT int = 3

// Usage strings
const USAGE_STRING = "./run-client.sh -d host:port baseDir blockSize"

const DEBUG_NAME = "d"
const DEBUG_USAGE = "Output log statements"

const ADDR_NAME = "host:port"
const ADDR_USAGE = "IP address and port of the MetaStore the client is syncing to"

const BASEDIR_NAME = "baseDir"
const BASEDIR_USAGE = "Base directory of the client"

const BLOCK_NAME = "blockSize"
const BLOCK_USAGE = "Size of the blocks used to fragment files"

// Exit codes
const EX_USAGE int = 64

func main() {
	// Custom flag Usage message
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage of %s:\n", USAGE_STRING)
		fmt.Fprintf(w, "  -%s: %v\n", DEBUG_NAME, DEBUG_USAGE)
		fmt.Fprintf(w, "  %s: %v\n", ADDR_NAME, ADDR_USAGE)
		fmt.Fprintf(w, "  %s: %v\n", BASEDIR_NAME, BASEDIR_USAGE)
		fmt.Fprintf(w, "  %s: %v\n", BLOCK_NAME, BLOCK_USAGE)
	}

	// Parse command-line arguments and flags
	debug := flag.Bool("d", false, DEBUG_USAGE)
	flag.Parse()

	// Use tail arguments to hold non-flag arguments
	args := flag.Args()

	if len(args) != ARG_COUNT {
		flag.Usage()
		os.Exit(EX_USAGE)
	}

	hostPort := args[0]
	baseDir := args[1]
	blockSize, err := strconv.Atoi(args[2])
	if err != nil {
		flag.Usage()
		os.Exit(EX_USAGE)
	}

	// Disable log outputs if debug flag is missing
	if !(*debug) {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	rpcClient := surfstore.NewSurfstoreRPCClient(hostPort, baseDir, blockSize)
	surfstore.ClientSync(rpcClient)
}
