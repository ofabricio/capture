package main

import (
	"flag"
	"net/url"
	"strconv"
)

type Args struct {
	url         *url.URL
	port        string
	dashboard   string
	maxCaptures int
}

var args Args

func parseArgs() {
	proxyURL := flag.String("url", "https://jsonplaceholder.typicode.com", "Required. Set the base url you want to capture")
	proxyPort := flag.Int("port", 9000, "Set the proxy port")
	dashboard := flag.String("dashboard", "dashboard", "Set the dashboard name")
	maxCaptures := flag.Int("max-captures", 16, "Set the max number of captures to show in the dashboard")
	flag.Parse()
	args = Args{}
	args.url, _ = url.Parse(*proxyURL)
	args.port = strconv.Itoa(*proxyPort)
	args.dashboard = *dashboard
	args.maxCaptures = *maxCaptures
}
