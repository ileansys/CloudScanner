package main

import (
	"sync"

	"ileansys.com/cloudiff/aws"
	"ileansys.com/cloudiff/baseliner"
	"ileansys.com/cloudiff/data"
	"ileansys.com/cloudiff/do"
	"ileansys.com/cloudiff/gcp"
	"ileansys.com/cloudiff/scanner"
)

//Baseline - baseline
func main() {

	var wg sync.WaitGroup
	outliers := make(chan []string)

	wg.Add(1)
	go scanOutliers(&wg, outliers)
	wg.Add(1)
	go checkIPChanges(data.DOIPsKey.String(), do.GetDOIPs(), &wg, outliers) //get IP changes on DO and scan outliers
	wg.Add(1)
	go checkIPChanges(data.GCPIPsKey.String(), gcp.GetGCPIPs(), &wg, outliers) //get IP changes on GCP and scan outliers
	wg.Add(1)
	go checkIPChanges(data.AWSIPsKey.String(), aws.GetAWSIPs(), &wg, outliers) //get IP changes on AWS and scan outliers
	wg.Wait()

}

func checkIPChanges(cloudProviderIPskey string, newIPs []string, wg *sync.WaitGroup, outliers chan []string) {
	defer wg.Done()
	baseliner.CheckIPBaselineChange(cloudProviderIPskey, newIPs, outliers)
}

func scanOutliers(wg *sync.WaitGroup, outliers chan []string) {
	defer wg.Done()
	for outlier := range outliers {
		wg.Add(1)
		go scanner.ServiceScan(outlier, wg)
	}
	close(outliers)
}
