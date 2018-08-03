package main

import (
	"flag"
	"net/url"
)

type Args struct {
	targetURL   *url.URL
	proxyPort   string
	dashboard   string
	maxCaptures int
}

func ParseArgs() Args {
	targetURL := flag.String("url", "https://jsonplaceholder.typicode.com", "Required. Set the base url you want to capture")
	proxyPort := flag.String("port", "9000", "Set the proxy port")
	dashboard := flag.String("dashboard", "dashboard", "Set the dashboard name")
	maxCaptures := flag.Int("max-captures", 16, "Set the max number of captures to show in the dashboard")
	flag.Parse()
	url, _ := url.Parse(*targetURL)
	return Args{url, *proxyPort, *dashboard, *maxCaptures}
}
