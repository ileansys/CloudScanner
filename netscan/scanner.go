package netscan

import (
	"bytes"
	"fmt"
	"log"
	"sync"

	"github.com/Ullaakut/nmap"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/google/go-cmp/cmp"
	"ileansys.com/cloudiff/cloudprovider"
	"ileansys.com/cloudiff/data"
	"ileansys.com/cloudiff/notifier"
)

//NetworkScan - properties of network scan
type NetworkScan struct {
	Type string
}

//Service - Conduct a Service Scan
func (ns NetworkScan) Service(providerResultsKey string, ipList []string, counter chan int, mc *memcache.Client) {

	var (
		resultBytes []byte
		errorBytes  []byte
	)

	log.Printf("Scanning outliers... %s", ipList)
	scanner, err := nmap.NewScanner(
		nmap.WithTargets(ipList...),
		nmap.WithPorts("80,443,27017,27018,5432,3306,6379,6380,22,2222"), ///Check for the Data ;-p
		nmap.WithServiceInfo(),
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
		if providerResultsKey != "LocalHostNmapResults" {
			CompareTwoServiceScans(providerResultsKey, resultBytes, mc)
		}
		data.StoreNmapScanResults(mc, providerResultsKey, resultBytes)
	}

	counter <- 1
}

//CompareTwoServiceScans - To check for changes in service basleines
func CompareTwoServiceScans(resultsKey string, newServiceBaseline []byte, mc *memcache.Client) {
	results, err := data.GetNmapScanResults(mc, resultsKey) //Get Service Baseline
	if err != nil {
		miss := err.Error() == "memcache: cache miss"
		if miss {
			return
		}
	}

	baseline, berr := Parse(bytes.TrimSpace(results)) //Parse the Service Baseline
	if berr != nil {
		log.Fatal(berr)
	}

	changes, cerr := Parse(newServiceBaseline) //Parse new scan results
	if cerr != nil {
		log.Fatal(cerr)
	}

	if diff := cmp.Diff(baseline.Hosts, changes.Hosts); diff != "" { //Compare results
		changes := fmt.Sprintf("Service Baseline Changes: (+baseline -changes):\n %s", diff)
		notifier.EmailAlert{Body: changes, ProviderName: resultsKey}.Send() //Send alert to show differences
	}
}

//Outliers - Scan outlierss
func Outliers(wg *sync.WaitGroup, outliers chan cloudprovider.Outlier, counter chan int, mc *memcache.Client) {
	defer wg.Done()
	ns := NetworkScan{Type: "Service"}
	for out := range outliers {
		go ns.Service(out.ResultsKey, out.IPs, counter, mc)
	}
}
