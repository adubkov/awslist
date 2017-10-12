package main

import (
	"log"
	"time"
)

var (
	// @readonly
	profilesPullerMsg = "[INFO] Profiles fetched from file: %v"
	regionsPullerMsg  = "[INFO] Regions fetched from aws: %v"
)

func runProfilesPuller(ticker *time.Ticker) {
	for range ticker.C {
		profiles, _ = fetchProfiles()
		log.Printf(profilesPullerMsg, profiles)
	}
}

func runRegionsPuller(ticker *time.Ticker) {
	for range ticker.C {
		regions, _ = fetchRegions()
		log.Printf(regionsPullerMsg, regions)
	}
}

func runInstancesPuller(ticker *time.Ticker) {
	for range ticker.C {
		instances = fetchInstances()
	}
}

func runElbPuller(ticker *time.Ticker) {
	for range ticker.C {
		elbs = fetchElb()
	}
}
