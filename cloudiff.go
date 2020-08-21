package main

import (
	"log"
	"sync"

	"github.com/jasonlvhit/gocron"
	"ileansys.com/cloudiff/baseliner"
	"ileansys.com/cloudiff/cloudprovider"
	"ileansys.com/cloudiff/data"
	"ileansys.com/cloudiff/scanner"
)

func main() {

	//Scan Outliers Scheduler
	scanOutliersScheduler := gocron.NewScheduler()
	scanOutliersScheduler.Every(5).Minute().Do(scan)
	scanOutliersScheduler.Every(12).Minute().Do(update)
	<-scanOutliersScheduler.Start()
	_, stime := scanOutliersScheduler.NextRun()
	log.Printf("Running scan at %v", stime)

	//Update Baseline Scheduler
	// updateBaselineScheduler := gocron.NewScheduler()
	// updateBaselineScheduler.Every(10).Minute().Do(update)
	// <-updateBaselineScheduler.Start()
	// _, utime := updateBaselineScheduler.NextRun()
	// fmt.Printf("Running baseline update at %v", utime)

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
	counter := make(chan int)

	//track scanners
	go trackScanners(len(providers), outliers, counter)
	swg.Add(1)

	//scan outliers sent from baseliner
	go scanOutliers(&swg, outliers, counter)
	for _, p := range providers {
		swg.Add(1)
		go checkIPChanges(p, &swg, outliers)
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

func checkIPChanges(provider cloudprovider.Provider, wg *sync.WaitGroup, outliers chan cloudprovider.Outlier) {
	defer wg.Done()
	baseliner.CheckIPBaselineChange(&provider, outliers)
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
