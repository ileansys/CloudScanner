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
	numberOfProviders = 2
)

func main() {

	var wg sync.WaitGroup
	outliers := make(chan cloudprovider.Outlier, numberOfProviders)
	counter := make(chan int)

	do := cloudprovider.Provider{
		ProviderName: "DO",
		IPKey:        data.DOIPsKey.String(),
		ResultsKey:   data.DONmapResultsKey.String(),
	}.Init()

	aws := cloudprovider.Provider{
		ProviderName: "AWS",
		IPKey:        data.AWSIPsKey.String(),
		ResultsKey:   data.AWSNmapResultsKey.String(),
	}.Init()

	go func() {
		c := 0
		for {
			select {
			case i := <-counter:
				c = c + i
				log.Printf("Scanner %d completed...", c)
				if c == numberOfProviders {
					close(outliers)
					break
				}
			}
		}
	}()

	wg.Add(1)
	go scanOutliers(&wg, outliers, counter)
	wg.Add(1)
	go checkIPChanges(&aws, &wg, outliers)
	wg.Add(1)
	go checkIPChanges(&do, &wg, outliers)
	wg.Wait()
}

func checkIPChanges(provider *cloudprovider.Provider, wg *sync.WaitGroup, outliers chan cloudprovider.Outlier) {
	defer wg.Done()
	baseliner.CheckIPBaselineChange(provider, outliers)
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
