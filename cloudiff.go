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
	gocron.Every(5).Minute().Do(scan)
	<-gocron.Start()
	_, stime := gocron.NextRun()
	log.Printf("Running scan at %v", stime)

}

func scan() {

	var swg sync.WaitGroup
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

	//ip outliers and service change channels
	ipOutliers := make(chan cloudprovider.Outlier, len(providers))
	serviceChanges := make(chan netscan.ServiceChanges, len(providers))

	//change alert channels
	ipChangeAlerts := make(chan notifier.EmailAlert, len(providers))
	serviceChangeAlerts := make(chan notifier.EmailAlert, len(providers))

	//update counters and close channels
	scanCounter := make(chan int)
	serviceChangesCounter := make(chan int)
	ipAlertCounter := make(chan int)
	serviceChangeAlertCounter := make(chan int)

	//track scanners
	go netscan.TrackScanners(len(providers), ipOutliers, scanCounter)

	//track ip change alerts
	go notifier.TrackIPChangeAlerts(len(providers), ipChangeAlerts, ipAlertCounter)

	//recieve and scan outliers sent from baseliner
	swg.Add(1)
	go netscan.RecieveAndScanOutliers(&swg, ipOutliers, serviceChanges, scanCounter, mc)

	//send IP change alerts from baseliner
	swg.Add(1)
	go notifier.SendIPChangeAlerts(&swg, ipChangeAlerts, ipAlertCounter)

	//check for IP changes
	for _, p := range providers {
		swg.Add(1)
		go checkIPChanges(mc, p, &swg, ipOutliers, ipChangeAlerts)
	}

	//track change alerts from baseliner
	go notifier.TrackServiceChangeAlerts(len(providers), serviceChangeAlerts, serviceChangeAlertCounter)

	//track number of service changes
	go baseliner.TrackServiceChanges(len(providers), serviceChanges, serviceChangesCounter)

	//send service baseline change alerts
	swg.Add(1)
	go notifier.SendServiceChangeAlerts(&swg, serviceChangeAlerts, serviceChangeAlertCounter)

	//check for service baseline changes
	go baseliner.CheckServiceBaselineChanges(serviceChangeAlerts, serviceChanges, serviceChangesCounter, mc)

	swg.Wait()

}

func checkIPChanges(mc *memcache.Client, provider cloudprovider.Provider, wg *sync.WaitGroup, outliers chan cloudprovider.Outlier, alerts chan notifier.EmailAlert) {
	defer wg.Done()
	baseliner.CheckIPBaselineChange(&provider, outliers, alerts, mc)
}
