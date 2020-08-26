package netscan

import (
	"log"
	"sync"

	"github.com/Ullaakut/nmap"
	"github.com/bradfitz/gomemcache/memcache"
	"ileansys.com/cloudiff/cloudprovider"
)

//ServiceChanges - object with service change properties
type ServiceChanges struct {
	NewServiceScanResults []byte
	ProviderResultstKey   string
}

//NetworkScan - properties of network scan
type NetworkScan struct {
	Type string
}

//Service - Conduct a Service Scan
func (ns NetworkScan) Service(providerResultsKey string, ipList []string, serviceChanges chan ServiceChanges, counter chan int, mc *memcache.Client) {

	var (
		resultBytes []byte
		errorBytes  []byte
	)

	log.Printf("Scanning outliers... %s", ipList)
	nseScripts := []string{
		"mysql-info",
		"mysql-brute",
		"mysql-databases",
		"mongodb-info",
		"mongodb-databases",
		"mongodb-brute",
		"redis-info",
		"redis-brute",
		"pgsql-brute",
		"http-jsonp-detection",
		"couchdb-databases",
		"couchdb-stats",
	}

	scanner, err := nmap.NewScanner(
		nmap.WithTargets(ipList...),
		nmap.WithPorts("80,443,27017,27018,5432,3306,6379,6380,22,2222,5984,8091,8092,8093,8094,8095,8096"), //Check for the Data ;-p
		nmap.WithScripts(nseScripts...),
		nmap.WithServiceInfo(),
		nmap.WithFilterHost(func(h nmap.Host) bool {
			for idx := range h.Ports {
				if h.Ports[idx].Status() == "open" {
					return true
				}
			}
			return false
		}),
	)

	if err != nil {
		log.Fatalf("unable to create nmap scanner: %v", err)
	}

	// Executes asynchronously, allowing results to be streamed in real time.
	if err := scanner.RunAsync(); err != nil {
		panic(err)
	}

	// Connect to stdout of scanner.
	stdout := scanner.GetStdout()

	// Connect to stderr of scanner.
	stderr := scanner.GetStderr()

	for stdout.Scan() {
		log.Println(stdout.Text())
		resultBytes = append(resultBytes, stdout.Bytes()...)
	}

	for stderr.Scan() {
		errorBytes = append(errorBytes, stderr.Bytes()...)
	}

	// Blocks until the scan has completed.
	if err := scanner.Wait(); err != nil {
		panic(err)
	} else {
		serviceChanges <- ServiceChanges{ProviderResultstKey: providerResultsKey, NewServiceScanResults: resultBytes}
	}

	counter <- 1
}

//RecieveAndScanOutliers - Receive and Scan outliers
func RecieveAndScanOutliers(wg *sync.WaitGroup, outliers chan cloudprovider.Outlier, serviceChanges chan ServiceChanges, counter chan int, mc *memcache.Client) {
	defer wg.Done()
	ns := NetworkScan{Type: "Service"}
	for out := range outliers {
		go ns.Service(out.ResultsKey, out.IPs, serviceChanges, counter, mc)
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
