package main

import (
	"flag"
	"fmt"
)

type Args struct {
	TargetURL             string `json:"targetURL"`
	ProxyPort             string `json:"proxyPort"`
	MaxCaptures           int    `json:"maxCaptures"`
	Dashboard             string `json:"dashboard"`
	DashboardPath         string `json:"dashboardPath"`
	DashboardClearPath    string `json:"dashboardClearPath"`
	DashboardItemInfoPath string `json:"dashboardItemInfoPath"`
}

func ParseArgs() Args {
	targetURL := flag.String("url", "https://jsonplaceholder.typicode.com", "Required. Set the base url you want to capture")
	proxyPort := flag.String("port", "9000", "Set the proxy port")
	dashboard := flag.String("dashboard", "dashboard", "Set the dashboard name")
	maxCaptures := flag.Int("max-captures", 16, "Set the max number of captures to show in the dashboard")
	flag.Parse()

	dashboardPath := fmt.Sprintf("/%s/", *dashboard)
	dashboardClearPath := fmt.Sprintf("/%s/clear/", *dashboard)
	dashboardItemInfoPath := fmt.Sprintf("/%s/items/", *dashboard)

	return Args{
		TargetURL:             *targetURL,
		ProxyPort:             *proxyPort,
		MaxCaptures:           *maxCaptures,
		Dashboard:             *dashboard,
		DashboardPath:         dashboardPath,
		DashboardClearPath:    dashboardClearPath,
		DashboardItemInfoPath: dashboardItemInfoPath,
	}
}
