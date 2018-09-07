package main

import "flag"

type Args struct {
	TargetURL   string `json:"targetURL"`
	ProxyPort   string `json:"proxyPort"`
	Dashboard   string `json:"dashboard"`
	MaxCaptures int    `json:"maxCaptures"`
}

func ParseArgs() Args {
	targetURL := flag.String("url", "https://jsonplaceholder.typicode.com", "Required. Set the base url you want to capture")
	proxyPort := flag.String("port", "9000", "Set the proxy port")
	dashboard := flag.String("dashboard", "dashboard", "Set the dashboard name")
	maxCaptures := flag.Int("max-captures", 16, "Set the max number of captures to show in the dashboard")
	flag.Parse()
	return Args{*targetURL, *proxyPort, *dashboard, *maxCaptures}
}
