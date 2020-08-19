package main

import (
	"log"
	"sync"

	"ileansys.com/cloudiff/baseliner"
	"ileansys.com/cloudiff/cloudprovider"
	"ileansys.com/cloudiff/data"
	"ileansys.com/cloudiff/scanner"
)

func main() {

	var wg sync.WaitGroup

	outliers := make(chan []string)
	results := make(chan []byte)

	aws := cloudprovider.Provider{ProviderName: "AWS", IPKey: data.AWSIPsKey.String()}.Init() //Create AWS Provider
	do := cloudprovider.Provider{ProviderName: "DO", IPKey: data.DOIPsKey.String()}.Init()    //Create DO Provider
	//gcp := cloudprovider.Provider{ProviderName: "GCP", IPKey: data.DOIPsKey.String()}.Init()  //Create DO Provider

	wg.Add(1)
	go scanOutliers(&wg, outliers, results)
	wg.Add(1)
	go processScanResults(&wg, results)
	wg.Add(1)
	go checkIPChanges(&do, &wg, outliers) //get IP changes on DO and scan outliers
	wg.Add(1)
	go checkIPChanges(&aws, &wg, outliers) //get IP changes on AWS and scan outliers
	wg.Wait()
	close(outliers)
	close(results)

	wg.Add(1)
	go updateIPBaselineData(&do, &wg) //update DO baseline
	wg.Add(1)
	go updateIPBaselineData(&aws, &wg) //update AWS IP baseline
}

func checkIPChanges(provider *cloudprovider.Provider, wg *sync.WaitGroup, outliers chan []string) {
	defer wg.Done()
	baseliner.CheckIPBaselineChange(provider, outliers)
}

func updateIPBaselineData(provider *cloudprovider.Provider, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("Updating %s baseline...", provider.ProviderName)
	data.StoreIPSByProvider(provider)
}

func scanOutliers(wg *sync.WaitGroup, outliers chan []string, results chan []byte) {
	defer wg.Done()
	for outlier := range outliers {
		wg.Add(1)
		go scanner.ServiceScan(outlier, wg, results)
	}
}

func processScanResults(wg *sync.WaitGroup, results chan []byte) {
	defer wg.Done()
	for result := range results {
		log.Println(string(result))
	}
}
