package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// Goroutine function fetches and parses the passed url in order to find insecure resources and next urls to fetch from.
func fetchUrl(url string, queue chan string, registry *Registry) {

	// Lock url so that no one other goroutine can process it.
	registry.MarkAsProcessed(url)

	fetcher := InsecureResourceFetcher{}

	insecureResourceUrls, pageUrls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Errorf("Error occured: %v\n", err)
		return
	}

	for _, insecureResourceUrl := range insecureResourceUrls {
		fmt.Printf("%s: %s\n", url, insecureResourceUrl)
	}

	for _, url := range pageUrls {
		queue <- url
	}

}

// Crawl pages starting with url and find insecure resources.
func crawl(url string, fetcher Fetcher) {

	url = strings.TrimSuffix(url, "/")

	registry := &Registry{processed: make(map[string]int)}

	queue := make(chan string)

	go fetchUrl(url, queue, registry)

	tick := time.Tick(2000 * time.Millisecond)

	flag := false
	for {
		select {
		case url := <-queue:
			flag = false

			// Ignore processed urls.
			if !registry.IsNew(url) {
				continue
			}
			go fetchUrl(url, queue, registry)
		case <-tick:
			if flag {
				fmt.Println("-----")
				fmt.Printf("Analized pages:\n")
				fmt.Println("-----")
				fmt.Println(registry)
				return
			} else {
				flag = true
			}
		}
	}

}

// Get start url from the command line arguments.
func startUrl() (string, error) {
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		return "", errors.New("Please specify a starting point, e.g. https://example.com")
	}

	return args[0], nil
}

func main() {

	startUrl, err := startUrl()
	if err != nil {
		fmt.Errorf("Error occured: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("-----")
	fmt.Printf("Insecure resources (page: resource):\n")
	fmt.Println("-----")

	crawl(startUrl, InsecureResourceFetcher{})
}
