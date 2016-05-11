package main

import (
	"log"
	"time"
)

var (
	// @readonly
	profilesPoolerMsg = "[INFO] Profiles fetched from file: %v"
	regionsPoolerMsg  = "[INFO] Regions fetched from aws: %v"
)

func runInstancesPoller(ticker *time.Ticker) {
	for range ticker.C {
		fetchInstances()
	}
}

func runRegionsPoller(ticker *time.Ticker) {
	for range ticker.C {
		regions, _ = fetchRegions()
		log.Printf(regionsPoolerMsg, regions)
	}
}

func runProfilesPoller(ticker *time.Ticker) {
	for range ticker.C {
		profiles, _ = fetchProfiles()
		log.Printf(profilesPoolerMsg, profiles)
	}
}
