package netscan

import (
	"io/ioutil"
	"log"
	"sync"

	"cloudscanner.app/pkg/cloudprovider"
	"github.com/Ullaakut/nmap"
	"github.com/bradfitz/gomemcache/memcache"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Cron    [2]string `yaml:"cron"`
	Ports   [1]string `yaml:"ports"`
	Scripts []string  `yaml:"scripts"`
}

//ServiceChanges - object with service change properties
type ServiceChanges struct {
	NewServiceScanResults []byte
	ProviderResultstKey   string
}

//NetworkScan - properties of network scan
type NetworkScan struct {
	Type string
}

func GetConfig() Config {
	var config Config
	//commands := make([][]string, 0)
	originFile, _ := ioutil.ReadFile("config.yaml")
	yaml.Unmarshal(originFile, &config)

	return config
}

//Service - Conduct a Service Scan
func (ns NetworkScan) Service(providerResultsKey string, ipList []string, serviceChanges chan ServiceChanges, counter chan int, mc *memcache.Client) {

	var (
		resultBytes []byte
		errorBytes  []byte
		config      Config
	)

	config = GetConfig()
	nseScripts := make([]string, 0)
	nmapPorts := make([]string, 0)

	for _, script := range config.Scripts {
		nseScripts = append(nseScripts, script)
	}

	for _, ports := range config.Ports {
		nmapPorts = append(nmapPorts, ports)
	}

	scanner, err := nmap.NewScanner(
		nmap.WithTargets(ipList...),
		nmap.WithPorts(nmapPorts...),
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
		log.Println(stderr.Text())
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
