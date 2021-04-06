package main

import (
	"flag"
)

// Config has all the configuration parsed from the command line.
type Config struct {
	TargetURL     string
	ProxyPort     string
	DashboardPort string
	MaxCaptures   int
}

// ReadConfig reads the arguments from the command line.
func ReadConfig() Config {
	targetURL := flag.String("url", "https://jsonplaceholder.typicode.com", "Required. Set the url you want to proxy")
	proxyPort := flag.String("port", "9000", "Set the proxy port")
	dashboardPort := flag.String("dashboard", "9001", "Set the dashboard port")
	maxCaptures := flag.Int("captures", 16, "Set how many captures to show in the dashboard")
	flag.Parse()
	return Config{
		TargetURL:     *targetURL,
		ProxyPort:     *proxyPort,
		MaxCaptures:   *maxCaptures,
		DashboardPort: *dashboardPort,
	}
}
