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
		"banner",
		"mysql-info",
		"mysql-empty-password",
		"mysql-brute",
		"mysql-databases",
		"mongodb-info",
		"mongodb-brute",
		"mongodb-databases",
		"redis-info",
		"redis-brute",
		"couchdb-databases",
		"couchdb-stats",
		"elasticsearch",
		"membase-brute",
		"membase-http-info",
		"memcached-info",
		"pgsql-brute",
		"http-jsonp-detection",
	}

	scanner, err := nmap.NewScanner(
		nmap.WithTargets(ipList...),
		nmap.WithPorts("80,443,8080,27017-27020,5432,3306,6379,6380,22,2222,5984,9200,9201,5601,9300,9301,4369,8091-8096,9100-9106,9110-9118,9120-9122,9130,9999,11209-11211,21100"),
		nmap.WithScripts(nseScripts...),
		nmap.WithServiceInfo(),
		nmap.WithVersionAll(),
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
