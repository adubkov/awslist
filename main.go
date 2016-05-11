package main

import (
	"flag"
	"github.com/aws/aws-sdk-go/service/ec2"
	"sync"
	"time"
)

var (
	// @readonly
	portMsg     = "Listen port"
	serviceMsg  = "Run as service"
	intervalMsg = "Interval to pool data in seconds"
)

var output_buffer []string
var screen_buffer []string
var service *bool
var port *int
var interval *int
var counter int
var wg sync.WaitGroup

var profiles []string
var regions []string

var instances []ec2.Instance

func main() {
	// Parse arguments
	port = flag.Int("port", 8080, portMsg)
	service = flag.Bool("service", false, serviceMsg)
	interval = flag.Int("interval", 30, intervalMsg)
	flag.Parse()

	profiles, _ = fetchProfiles()
	regions, _ = fetchRegions()
	instances = fetchInstances()

	// If specified service mode, run program as a service, and listen port
	if *service {
		// Each 30 seconds (by default)
		ticker := time.NewTicker(time.Second * time.Duration(*interval))
		go runInstancesPoller(ticker)

		general_ticker := time.NewTicker(time.Minute * time.Duration(5))
		go runRegionsPoller(general_ticker)
		go runProfilesPoller(general_ticker)

		// Run http server on specifig port
		new(HttpServer).Run(*port)
	}
}
