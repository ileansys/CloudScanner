package main

import (
	"log"
	"sync"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/jasonlvhit/gocron"
	"ileansys.com/cloudiff/baseliner"
	"ileansys.com/cloudiff/cloudprovider"
	"ileansys.com/cloudiff/data"
	"ileansys.com/cloudiff/netscan"
	"ileansys.com/cloudiff/notifier"
)

var (
	memcachedServer string = "127.0.0.1:11211"
)

func main() {

	//Scan Outliers Scheduler
	gocron.NewScheduler()
	gocron.Every(15).Minute().Do(scan)
	gocron.Every(28).Minute().Do(update)
	<-gocron.Start()
	_, stime := gocron.NextRun()
	log.Printf("Running scan at %v", stime)

}

func scan() {

	var swg sync.WaitGroup //scan and baseliner waitgroup
	mc := memcache.New(memcachedServer)
	var providers = []cloudprovider.Provider{
		cloudprovider.Provider{
			ProviderName: "DO",
			IPKey:        data.DOIPsKey.String(),
			OutlierKey:   data.DOOutliersKey.String(),
			ResultsKey:   data.DONmapResultsKey.String(),
		}.Init(),
		cloudprovider.Provider{
			ProviderName: "AWS",
			IPKey:        data.AWSIPsKey.String(),
			OutlierKey:   data.AWSOutliersKey.String(),
			ResultsKey:   data.AWSNmapResultsKey.String(),
		}.Init(),
		cloudprovider.Provider{
			ProviderName: "GCP",
			IPKey:        data.GCPIPsKey.String(),
			OutlierKey:   data.GCPOutliersKey.String(),
			ResultsKey:   data.GCPNmapResultsKey.String(),
		}.Init(),
	}

	//channel size based on number of providers
	outliers := make(chan cloudprovider.Outlier, len(providers))
	alerts := make(chan notifier.EmailAlert, len(providers))
	scanCounter := make(chan int)
	alertCounter := make(chan int)

	//track scanners
	go TrackScanners(len(providers), outliers, scanCounter)

	//track alerts
	go notifier.TrackEmailAlerts(len(providers), alerts, alertCounter)

	//scan outliers sent from baseliner
	swg.Add(1)
	go netscan.Outliers(&swg, outliers, scanCounter, mc)
	swg.Add(1)
	go sendAlerts(&swg, alerts, alertCounter)

	//check ip changes
	for _, p := range providers {
		swg.Add(1)
		go checkIPChanges(mc, p, &swg, outliers, alerts)
	}
	swg.Wait()

}

func update() {
	var uwg sync.WaitGroup //update baseline waitgroup
	mc := memcache.New(memcachedServer)
	var providers = []cloudprovider.Provider{
		cloudprovider.Provider{
			ProviderName: "DO",
			IPKey:        data.DOIPsKey.String(),
			OutlierKey:   data.DOOutliersKey.String(),
			ResultsKey:   data.DONmapResultsKey.String(),
		}.Init(),
		cloudprovider.Provider{
			ProviderName: "AWS",
			IPKey:        data.AWSIPsKey.String(),
			OutlierKey:   data.AWSOutliersKey.String(),
			ResultsKey:   data.AWSNmapResultsKey.String(),
		}.Init(),
		cloudprovider.Provider{
			ProviderName: "GCP",
			IPKey:        data.GCPIPsKey.String(),
			OutlierKey:   data.GCPOutliersKey.String(),
			ResultsKey:   data.GCPNmapResultsKey.String(),
		}.Init(),
	}
	//update IP baseline
	for _, p := range providers {
		uwg.Add(1)
		go updateIPBaselineData(mc, p, &uwg)
	}
	uwg.Wait()
}

func checkIPChanges(mc *memcache.Client, provider cloudprovider.Provider, wg *sync.WaitGroup, outliers chan cloudprovider.Outlier, alerts chan notifier.EmailAlert) {
	defer wg.Done()
	baseliner.CheckIPBaselineChange(&provider, outliers, alerts, mc)
}

func updateIPBaselineData(mc *memcache.Client, provider cloudprovider.Provider, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("Updating %s baseline...", provider.ProviderName)
	data.StoreIPSByProvider(mc, &provider)
}

func sendAlerts(wg *sync.WaitGroup, alerts chan notifier.EmailAlert, aCounter chan int) {
	defer wg.Done()
	for alert := range alerts {
		go alert.SendViaChannel(aCounter)
	}
}

//TrackScanners - numberOfScanners is equal numberOfProviders
func TrackScanners(numberOfScanners int, outliers chan cloudprovider.Outlier, counter chan int) {
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
