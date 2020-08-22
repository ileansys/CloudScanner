package main

import (
	"log"
	"sync"

	"github.com/jasonlvhit/gocron"
	"ileansys.com/cloudiff/baseliner"
	"ileansys.com/cloudiff/cloudprovider"
	"ileansys.com/cloudiff/data"
	"ileansys.com/cloudiff/notifier"
	"ileansys.com/cloudiff/scanner"
)

func main() {

	//Scan Outliers Scheduler
	scanScheduler := gocron.NewScheduler()
	scanScheduler.Every(5).Minute().Do(scan)
	//scanScheduler.Every(12).Minute().Do(update)
	<-scanScheduler.Start()
	_, stime := scanScheduler.NextRun()
	log.Printf("Running scan at %v", stime)
}

func scan() {

	var swg sync.WaitGroup //scan and baseliner waitgroup
	//some provider properties to be pre-configured/marshalled on yaml later...
	var providers = []cloudprovider.Provider{
		cloudprovider.Provider{
			ProviderName: "DO",
			IPKey:        data.DOIPsKey.String(),
			ResultsKey:   data.DONmapResultsKey.String(),
		}.Init(),
		cloudprovider.Provider{
			ProviderName: "AWS",
			IPKey:        data.AWSIPsKey.String(),
			ResultsKey:   data.AWSNmapResultsKey.String(),
		}.Init(),
		cloudprovider.Provider{
			ProviderName: "GCP",
			IPKey:        data.GCPIPsKey.String(),
			ResultsKey:   data.GCPNmapResultsKey.String(),
		}.Init(),
	}

	//channel size based on number of providers
	outliers := make(chan cloudprovider.Outlier, len(providers))
	alerts := make(chan notifier.EmailAlert, len(providers))
	scanCounter := make(chan int)
	alertCounter := make(chan int)

	//track scanners
	go trackScanners(len(providers), outliers, scanCounter)

	//track alerts
	go trackEmailAlerts(len(providers), alerts, alertCounter)

	//scan outliers sent from baseliner
	swg.Add(1)
	go scanOutliers(&swg, outliers, scanCounter)
	swg.Add(1)
	go sendAlerts(&swg, alerts, alertCounter)

	//check ip changes
	for _, p := range providers {
		swg.Add(1)
		go checkIPChanges(p, &swg, outliers, alerts)
	}
	swg.Wait()

}

func update() {
	var uwg sync.WaitGroup //update baseline waitgroup
	var providers = []cloudprovider.Provider{
		cloudprovider.Provider{
			ProviderName: "DO",
			IPKey:        data.DOIPsKey.String(),
			ResultsKey:   data.DONmapResultsKey.String(),
		}.Init(),
		cloudprovider.Provider{
			ProviderName: "AWS",
			IPKey:        data.AWSIPsKey.String(),
			ResultsKey:   data.AWSNmapResultsKey.String(),
		}.Init(),
		cloudprovider.Provider{
			ProviderName: "GCP",
			IPKey:        data.GCPIPsKey.String(),
			ResultsKey:   data.GCPNmapResultsKey.String(),
		}.Init(),
	}
	//update IP baseline
	for _, p := range providers {
		uwg.Add(1)
		go updateIPBaselineData(p, &uwg)
	}
	uwg.Wait()
}

func checkIPChanges(provider cloudprovider.Provider, wg *sync.WaitGroup, outliers chan cloudprovider.Outlier, alerts chan notifier.EmailAlert) {
	defer wg.Done()
	baseliner.CheckIPBaselineChange(&provider, outliers, alerts)
}

func updateIPBaselineData(provider cloudprovider.Provider, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("Updating %s baseline...", provider.ProviderName)
	data.StoreIPSByProvider(&provider)
}

func scanOutliers(wg *sync.WaitGroup, outliers chan cloudprovider.Outlier, counter chan int) {
	defer wg.Done()
	for outlier := range outliers {
		go scanner.ServiceScan(outlier.ResultsKey, outlier.IPs, counter)
	}
}

func sendAlerts(wg *sync.WaitGroup, alerts chan notifier.EmailAlert, aCounter chan int) {
	defer wg.Done()
	for alert := range alerts {
		go alert.Send(aCounter)
	}
}

//numberOfScanners is equal numberOfProviders
func trackScanners(numberOfScanners int, outliers chan cloudprovider.Outlier, counter chan int) {
	c := 0
	for {
		select {
		case i := <-counter:
			c = c + i
			log.Printf("Scanner #%d completed...", c)
			if c == numberOfScanners {
				close(outliers)
				break
			}
		}
	}
}

//numberOfAlerts is equal numberOfProviders
func trackEmailAlerts(numberOfAlerts int, alerts chan notifier.EmailAlert, counter chan int) {
	c := 0
	for {
		select {
		case i := <-counter:
			c = c + i
			log.Printf("Email #%d sent...", c)
			if c == numberOfAlerts {
				close(alerts)
				break
			}
		}
	}
}
