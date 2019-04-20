package main

import (
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/hyperledger/sawtooth-sdk-go/logging"
	"github.com/hyperledger/sawtooth-sdk-go/processor"
)

//const defaultURL = "http://localhost:4004"
const defaultURL = "tcp://validator:4004"
const familyName = "cookiejar"
const version = "1.0"

func main() {
	// Setup a logger
	logger := logging.Get()

	// Get settings via environment
	verbose := "WARN"
	if v := os.Getenv("CJ_VERBOSITY"); v != "" {
		verbose = strings.ToUpper(v)
	}

	endpoint := defaultURL
	if v := os.Getenv("CJ_CONNECT"); v != "" {
		endpoint = v
	}

	var queue uint = 100
	if v := os.Getenv("CJ_QUEUE"); v != "" {
		q, err := strconv.Atoi(v)
		if err != nil {
			logger.Errorf("Failed to parse CJ_QUEUE: %v", err)
			os.Exit(1)
		}
		queue = uint(q)
	}

	var threads uint = 0
	if v := os.Getenv("CJ_THREADS"); v != "" {
		t, err := strconv.Atoi(v)
		if err != nil {
			logger.Errorf("Failed to parse CJ_THREADS: %v", err)
			os.Exit(1)
		}
		threads = uint(t)
	}

	switch verbose {
	case "DEBUG":
		logger.SetLevel(logging.DEBUG)
	case "INFO":
		logger.SetLevel(logging.INFO)
	case "WARN":
		logger.SetLevel(logging.WARN)
	default:
		logger.Errorf("Invalid value %q for CJ_VERBOSITY", verbose)
	}

	// Initialize and register a new transaction processor
	processor := processor.NewTransactionProcessor(endpoint)
	processor.SetMaxQueueSize(queue)
	if threads > 0 {
		processor.SetThreadCount(threads)
	}

	processor.AddHandler(NewCookiejarHandler()) // Add the handler
	processor.ShutdownOnSignal(syscall.SIGINT, syscall.SIGTERM)

	if err := processor.Start(); err != nil {
		logger.Error("Processor stopped: ", err)
	}
}
