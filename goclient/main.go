package main

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/hyperledger/sawtooth-sdk-go/logging"
)

const (
	keyName = "mycookiejar"
	//defaultURL = "http://localhost:8008"
	defaultURL = "http://rest-api:8008"
)

var logger *logging.Logger = logging.Get()

// printHelp will print how to use the CLI tool
func printHelp(msg string) {
	if msg != "" {
		fmt.Println(msg)
	}
	fmt.Printf("Usage: %s <command>\n\nCommands:\n", os.Args[0])
	fmt.Printf("bake <amount>\neat <amount>\ncount\nclear")
}

// UserHomeDir returns the user's home directory
// for Go 1.11 and above use os.UserHomeDir from the standard library
func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

// getPrivateKeyFile reads and returns the locally stored private key
func getPrivateKeyFile(keyName string) (string, error) {
	home := UserHomeDir()

	return path.Join(home, ".sawtooth", "keys", keyName), nil
}

func main() {
	args := len(os.Args)
	if args < 2 {
		printHelp("")
		os.Exit(1)
	}

	// Get the locally stored private key
	keyFile, err := getPrivateKeyFile(fmt.Sprintf("%s.priv", keyName))
	if err != nil {
		fmt.Printf("Failed to generate filename: %v\n", err)
		os.Exit(1)
	}

	// Instantiate a new cookiejar client
	client, err := NewCookiejarClient(defaultURL, keyFile)
	if err != nil {
		fmt.Printf("Failed to initialize cookiejar client: %v\n", err)
		os.Exit(1)
	}

	// Check the exectured argument
	switch strings.ToLower(os.Args[1]) {
	case "bake":
		if args != 3 {
			printHelp("bake requires 1 argument")
			os.Exit(1)
		}

		// Convert the amount to int
		amount, err := strconv.Atoi(os.Args[2])
		if err != nil {
			printHelp(err.Error())
			os.Exit(2)
		}

		// Execute the action
		resp, err := client.bake(amount)
		if err != nil {
			fmt.Printf("Failed to register baked cookies: %v\n", err)
			os.Exit(2)
		}

		fmt.Println(resp)
	case "eat":
		if args != 3 {
			printHelp("eat requires 1 argument")
			os.Exit(1)
		}

		// Convert the amount to int
		amount, err := strconv.Atoi(os.Args[2])
		if err != nil {
			printHelp(err.Error())
			os.Exit(2)
		}

		// Execute the action
		resp, err := client.eat(amount)
		if err != nil {
			fmt.Printf("Failed to register eating cookies: %v\n", err)
			os.Exit(2)
		}

		fmt.Println(resp)
	case "count":
		// Execture the action
		resp, err := client.count()
		if err != nil {
			fmt.Printf("Failed to register baked cookies: %v\n", err)
			os.Exit(2)
		}

		fmt.Println(resp)
	case "clear":
		// Excecute the action
		err := client.clear()
		if err != nil {
			fmt.Printf("Failed to register baked cookies: %v\n", err)
			os.Exit(2)
		}
	default:
		printHelp("Invalid command")
		os.Exit(1)
	}
}
