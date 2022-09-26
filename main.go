package main

import (
	"log"
	"time"

	"cloudscanner.app/cmd"
	"cloudscanner.app/pkg/netscan"
	"github.com/go-co-op/gocron"
)

func main() {

	config := netscan.GetConfig()

	log.Println("Starting CloudScanner...")

	var scanFrequency string = config.Cron[0]
	var invalidateFrequency string = config.Cron[1]

	s := gocron.NewScheduler(time.UTC)

	// cron expressions supported
	s.Cron(scanFrequency).Do(cmd.Scan)             // every minute
	s.Cron(invalidateFrequency).Do(cmd.Invalidate) // every minute

	// // starts the scheduler and blocks current execution path
	s.StartBlocking()

}
