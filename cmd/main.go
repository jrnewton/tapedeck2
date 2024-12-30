package main

import (
	"fmt"
	"os"
	"tapedeck"
)

func usage() {
	fmt.Println("Usage: the first arg is either 'dev' or 'prod'. The subsequent args are based on the first arg.")
	fmt.Println("")
	fmt.Println("  dev server configuration, for local testing only.")
	fmt.Println(" ./tapedeck dev <directory> <listen addr>")
	fmt.Println("    directory: server working directory containing ./templates, ./static, etc.  Required.")
	fmt.Println("    listen addr: bind server to this address, format is <address>:<port>.  Use ':<port>' to bind to all local interfaces. Required.")
	fmt.Println("")
	fmt.Println("  prod server configuration, for production use.")
	fmt.Println(" ./tapedeck prod <json config file>")
	fmt.Println("    json config file: server configuration file.")
	os.Exit(4)
}

func main() {
	args := os.Args
	// first arg is always program name
	if len(args) < 2 {
		usage()
	}

	configType := args[1]

	switch configType {
	case "dev":
		if len(args) != 4 {
			usage()
		}

		serverDir := args[2]
		listenAddr := args[3]

		rc, serverErr := tapedeck.RunDevServer(serverDir, listenAddr)
		if serverErr != nil {
			fmt.Println(serverErr)
			os.Exit(rc)
		}
	case "prod":
		if len(args) != 3 {
			usage()
		}

		jsonConfigPath := args[2]
		rc, serverErr := tapedeck.RunProdServer(jsonConfigPath)
		if serverErr != nil {
			fmt.Println(serverErr)
			os.Exit(rc)
		}
	default:
		fmt.Println("unknown config type:", configType)
		os.Exit(62)
	}
}
