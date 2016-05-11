package main

import (
	"flag"
	"log"
	"sync"
	"time"
)

var (
	// @readonly
	portMsg           = "Listen port"
	serviceMsg        = "Run as service"
	intervalMsg       = "Interval to pool data in seconds"
	profilesPoolerMsg = "[INFO] Profiles fetched from file: %v"
)

// PrintInstances runs go routines to print instances from all regions within account
func PrintInstances(profile *Profile, regions []string) {
	defer wg.Done()
	for _, region := range regions {
		wg.Add(1)
		go NewEC2List(profile, region).ListInstances("")
	}
}

// getInstances run go routines to print all instances from all regions and all accounts
func getInstances() {
	var regions []string
	var profile *Profile

	// Clear output_buffer
	output_buffer = []string{}

	// Run go routines to print instances
	for _, profile_name := range profiles {
		// If we didn't load regions already, then fill regions slice
		profile = NewProfile(profile_name)

		if len(regions) == 0 {
			r, _ := NewEC2List(profile, "").ListRegions()
			for _, region := range r {
				regions = append(regions, region)
			}
		}

		wg.Add(1)
		go PrintInstances(profile, regions)
	}

	// Wait until receive info about all instances
	wg.Wait()

	// Resize and fill screen buffer with output data
	screen_buffer = make([]string, len(output_buffer), (cap(output_buffer)+1)*2)
	copy(screen_buffer, output_buffer)
}

// Continuously pool list of instances from aws.
func runInstancesPoller(ticker *time.Ticker) {
	for range ticker.C {
		// Get list of all instances
		getInstances()
	}
}

/*func runRegionsPoller(ticker *time.Ticker) {
    for range ticker.C {
        fetchRegions(&regions)
    }
}*/

func runProfilesPoller(ticker *time.Ticker) {
	for range ticker.C {
		profiles, _ = fetchProfiles()
		log.Printf(profilesPoolerMsg, profiles)
	}
}

var output_buffer []string
var screen_buffer []string
var service *bool
var port *int
var interval *int
var counter int
var wg sync.WaitGroup

var profiles []string
var regions []string

func main() {
	// Parse arguments
	port = flag.Int("port", 8080, portMsg)
	service = flag.Bool("service", false, serviceMsg)
	interval = flag.Int("interval", 30, intervalMsg)
	flag.Parse()

	// Get list of profiles from ~/.aws/config file
	profiles, _ = fetchProfiles()
	//regions, _ = fetchRegions()

	// Get list of instances
	getInstances()

	// If specified service mode, run program as a service, and listen port
	if *service {
		// Each 30 seconds (by default)
		ticker := time.NewTicker(time.Second * time.Duration(*interval))
		go runInstancesPoller(ticker)

		//go runRegionsPoller(time.NewTicker(time.Minute * time.Duration(5)))
		go runProfilesPoller(time.NewTicker(time.Second * time.Duration(5)))

		// Run http server on specifig port
		new(HttpServer).Run(*port)
	}
}
