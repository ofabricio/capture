package main

import (
	"flag"
	"fmt"
)

// Config has all the configuration parsed from the command line
type Config struct {
	TargetURL   string `json:"targetURL"`
	ProxyPort   string `json:"proxyPort"`
	MaxCaptures int    `json:"maxCaptures"`

	DashboardPath      string `json:"dashboardPath"`
	DashboardConnPath  string `json:"dashboardConnPath"`
	DashboardInfoPath  string `json:"dashboardInfoPath"`
	DashboardClearPath string `json:"dashboardClearPath"`
	DashboardRetryPath string `json:"dashboardRetryPath"`
}

// ReadConfig reads the arguments from the command line
func ReadConfig() Config {
	targetURL := flag.String("url", "https://jsonplaceholder.typicode.com", "Required. Set the base url you want to capture")
	proxyPort := flag.String("port", "9000", "Set the proxy port")
	dashboard := flag.String("dashboard", "dashboard", "Set the dashboard name")
	maxCaptures := flag.Int("max-captures", 16, "Set the max number of captures to show in the dashboard")
	flag.Parse()

	return Config{
		TargetURL:   *targetURL,
		ProxyPort:   *proxyPort,
		MaxCaptures: *maxCaptures,

		DashboardPath:      fmt.Sprintf("/%s/", *dashboard),
		DashboardConnPath:  fmt.Sprintf("/%s/conn/", *dashboard),
		DashboardInfoPath:  fmt.Sprintf("/%s/info/", *dashboard),
		DashboardClearPath: fmt.Sprintf("/%s/clear/", *dashboard),
		DashboardRetryPath: fmt.Sprintf("/%s/retry/", *dashboard),
	}
}
