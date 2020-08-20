package main

import (
	"log"
	"sync"

	"ileansys.com/cloudiff/baseliner"
	"ileansys.com/cloudiff/cloudprovider"
	"ileansys.com/cloudiff/data"
	"ileansys.com/cloudiff/scanner"
)

var (
	numberOfProviders = 3
)

func main() {

	var wg sync.WaitGroup
	var providers = []cloudprovider.Provider{ //some provider properties will be pre-configured/marshalled on yaml later...
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

	outliers := make(chan cloudprovider.Outlier, len(providers)) //size channel based on number of providers
	counter := make(chan int)

	go trackScanners(outliers, counter)
	wg.Add(1)
	go scanOutliers(&wg, outliers, counter)
	for _, p := range providers {
		wg.Add(1)
		go checkIPChanges(p, &wg, outliers)
	}
	wg.Wait()

}

func checkIPChanges(provider cloudprovider.Provider, wg *sync.WaitGroup, outliers chan cloudprovider.Outlier) {
	defer wg.Done()
	baseliner.CheckIPBaselineChange(&provider, outliers)
}

func updateIPBaselineData(provider *cloudprovider.Provider, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("Updating %s baseline...", provider.ProviderName)
	data.StoreIPSByProvider(provider)
}

func scanOutliers(wg *sync.WaitGroup, outliers chan cloudprovider.Outlier, counter chan int) {
	defer wg.Done()

	for outlier := range outliers {
		go scanner.ServiceScan(outlier.ResultsKey, outlier.IPs, counter)
	}
}

func trackScanners(outliers chan cloudprovider.Outlier, counter chan int) {
	c := 0
	for {
		select {
		case i := <-counter:
			c = c + i
			log.Printf("Scanner #%d completed...", c)
			if c == numberOfProviders {
				close(outliers)
				break
			}
		}
	}
}
