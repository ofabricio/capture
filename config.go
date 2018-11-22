package main

import (
	"flag"
	"fmt"
)

type Config struct {
	TargetURL             string `json:"targetURL"`
	ProxyPort             string `json:"proxyPort"`
	MaxCaptures           int    `json:"maxCaptures"`
	Dashboard             string `json:"dashboard"`
	DashboardPath         string `json:"dashboardPath"`
	DashboardConnPath     string `json:"dashboardConnPath"`
	DashboardClearPath    string `json:"dashboardClearPath"`
	DashboardItemInfoPath string `json:"dashboardItemInfoPath"`
}

func ReadConfig() Config {
	targetURL := flag.String("url", "https://jsonplaceholder.typicode.com", "Required. Set the base url you want to capture")
	proxyPort := flag.String("port", "9000", "Set the proxy port")
	dashboard := flag.String("dashboard", "dashboard", "Set the dashboard name")
	maxCaptures := flag.Int("max-captures", 16, "Set the max number of captures to show in the dashboard")
	flag.Parse()

	dashboardConnPath := "/socket.io/"
	dashboardPath := fmt.Sprintf("/%s/", *dashboard)
	dashboardClearPath := fmt.Sprintf("/%s/clear/", *dashboard)
	dashboardItemInfoPath := fmt.Sprintf("/%s/items/", *dashboard)

	return Config{
		TargetURL:             *targetURL,
		ProxyPort:             *proxyPort,
		MaxCaptures:           *maxCaptures,
		Dashboard:             *dashboard,
		DashboardPath:         dashboardPath,
		DashboardConnPath:     dashboardConnPath,
		DashboardClearPath:    dashboardClearPath,
		DashboardItemInfoPath: dashboardItemInfoPath,
	}
}
