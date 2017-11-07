package main

import (
	"flag"
	"strconv"
)

type Args struct {
	url         string
	port        string
	dashboard   string
	maxCaptures int
}

func parseArgs() Args {
	proxyURL := flag.String("url", "https://jsonplaceholder.typicode.com", "Required. Set the base url you want to capture")
	proxyPort := flag.Int("port", 9000, "Set the port you want to capture")
	maxCaptures := flag.Int("max-captures", 16, "Set the max number of captures")
	dashboard := flag.String("dashboard", "dashboard", "Set the dashboard name")
	flag.Parse()
	return Args{
		url:         *proxyURL,
		port:        strconv.Itoa(*proxyPort),
		dashboard:   *dashboard,
		maxCaptures: *maxCaptures,
	}
}
