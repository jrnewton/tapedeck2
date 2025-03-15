package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func processM3u(client *http.Client, res *http.Response) error {
	fmt.Printf("processing response body\n")

	// read the playlist
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error during io.ReadAll: %w", err)
	}

	payload := string(body)
	fmt.Printf("---- playlist contents start ----\n")
	fmt.Printf("%v", payload)
	fmt.Printf("---- playlist contents end   ----\n")

	lines := strings.Split(payload, "\n")
	for _, line := range lines {
		if strings.HasSuffix(line, ".mp3") {
			fmt.Printf("found mp3 %v\n", line)
			fetch(client, line)
		}
	}

	return nil
}

func processMp3(_ *http.Client, res *http.Response) error {
	fmt.Printf("processing response body\n")

	// get the file name
	path := res.Request.URL.Path
	segments := strings.Split(path, "/")
	name := segments[len(segments)-1]

	fmt.Printf("opening file %v\n", name)
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error during os.OpenFile: %w", err)
	}
	defer file.Close()
	fmt.Printf("file opened %v\n", name)

	// copy from response to file writer.
	w := bufio.NewWriter(file)
	written, err := io.Copy(w, res.Body)
	if err != nil {
		return fmt.Errorf("error during io.Copy: %w", err)
	}

	fmt.Printf("bytes written: %v\n", written)
	w.Flush()
	fmt.Printf("file writer flushed\n")

	return nil
}

func fetch(client *http.Client, url string) error {
	fmt.Printf("GET %v\n", url)

	res, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("error during GET %v: %w", url, err)
	}

	var process func(*http.Client, *http.Response) error

	// it's possible to have multiple values, this returns first.
	// do we need to deal with multiple values though?
	contentType := res.Header.Get("Content-Type")
	fmt.Printf("content-type: %q\n", contentType)

	switch contentType {
	case "audio/x-mpegurl":
		process = processM3u
	case "audio/mpeg":
		process = processMp3
	default:
		process = func(_ *http.Client, _ *http.Response) error {
			return fmt.Errorf("unsupported content type: %q", contentType)
		}
	}
	defer res.Body.Close()

	err = process(client, res)
	if err != nil {
		return fmt.Errorf("error processing response: %w", err)
	}

	return nil
}

func main() {
	var url string
	flag.StringVar(&url, "url", "", "URL to an audio file to process")

	flag.Parse()

	if url == "" {
		fmt.Println("url required")
		flag.Usage()
		os.Exit(4)
	}

	client := http.DefaultClient
	err := fetch(client, url)
	if err != nil {
		fmt.Printf("error during fetch: %v\n", err)
		os.Exit(8)
	}
}
