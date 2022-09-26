package cmd

import (
	"log"
	"sync"

	"cloudscanner.app/pkg/baseliner"
	"cloudscanner.app/pkg/cloudprovider"
	"cloudscanner.app/pkg/data"
	"cloudscanner.app/pkg/netscan"
	"cloudscanner.app/pkg/notifier"
	"github.com/bradfitz/gomemcache/memcache"
)

var (
	memcachedServer string = "127.0.0.1:11211"
)

func Scan() {

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
	serviceChangeAlerts := make(chan notifier.XMLEmailAlert, len(providers))

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

func Invalidate() {
	mc := memcache.New(memcachedServer) //Triggers an update and a scan
	err := mc.DeleteAll()
	if err != nil {
		log.Fatal(err)
	}
}

func checkIPChanges(mc *memcache.Client, provider cloudprovider.Provider, wg *sync.WaitGroup, outliers chan cloudprovider.Outlier, alerts chan notifier.EmailAlert) {
	defer wg.Done()
	baseliner.CheckIPBaselineChange(&provider, outliers, alerts, mc)
}
