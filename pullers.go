package main

import (
	"log"
	"time"
)

var (
	// @readonly
	regionsPullerMsg = "[INFO] Regions fetched from aws: %v"
)

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
