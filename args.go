package main

import (
	"flag"
	"net/url"
	"strconv"
)

func parseFlags() (*url.URL, string, string, int) {
	target := flag.String("url", "https://jsonplaceholder.typicode.com", "Required. Set the base url you want to capture")
	proxyPort := flag.Int("port", 9000, "Set the proxy port")
	dashboard := flag.String("dashboard", "dashboard", "Set the dashboard name")
	maxCaptures := flag.Int("max-captures", 16, "Set the max number of captures to show in the dashboard")
	flag.Parse()
	targetURL, _ := url.Parse(*target)
	return targetURL, strconv.Itoa(*proxyPort), *dashboard, *maxCaptures
}
