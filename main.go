package main

import (
	"flag"
	_ "github.com/aws/aws-sdk-go/service/ec2"
	"sync"
	"time"
)

var (
	// @readonly
	portMsg     = "Listen port"
	serviceMsg  = "Run as service"
	intervalMsg = "Interval to pool data in seconds"
	compatMsg   = "By default return v1 formatted output"
)

var service *bool
var compat *bool
var port *int
var interval *int
var counter int
var ec2_wg, elb_wg sync.WaitGroup

var profiles []string
var regions []string

var instances []Instance
var elbs []Elb

func main() {
	// Parse arguments
	port = flag.Int("port", 8080, portMsg)
	service = flag.Bool("service", false, serviceMsg)
	compat = flag.Bool("compat", false, compatMsg)
	interval = flag.Int("interval", 30, intervalMsg)
	flag.Parse()

	profiles, _ = fetchProfiles()
	regions, _ = fetchRegions()
	instances = fetchInstances()
	elbs = fetchElb()

	// If specified service mode, run program as a service, and listen port
	if *service {
		// Each 30 seconds (by default)
		ticker := time.NewTicker(time.Second * time.Duration(*interval))
		go runInstancesPoller(ticker)

		elb_ticker := time.NewTicker(time.Minute * time.Duration(1))
		go runElbPoller(elb_ticker)

		general_ticker := time.NewTicker(time.Minute * time.Duration(5))
		go runRegionsPoller(general_ticker)
		go runProfilesPoller(general_ticker)

		// Run http server on specifig port
		new(HttpServer).Run(*port)
	}
}
