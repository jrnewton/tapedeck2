package main

import (
	"flag"
	"fmt"
	"os"
	tapedeck "tapedeck/internal"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "configFile", "", "path to json config file (prod config only) ")

	flag.Parse()

	if configFile == "" {
		fmt.Println("configFile is required")
		flag.Usage()
		os.Exit(4)
	}

	rc, serverErr := tapedeck.RunServer(configFile)
	if serverErr != nil {
		fmt.Println(serverErr)
		os.Exit(rc)
	}
}
