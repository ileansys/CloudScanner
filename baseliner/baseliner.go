package baseliner

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"reflect"
	"sort"
	"strings"

	"ileansys.com/cloudiff/cloudprovider"
	"ileansys.com/cloudiff/netscan"
	"ileansys.com/cloudiff/notifier"

	"ileansys.com/cloudiff/data"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/scylladb/go-set/strset"
)

var (
	localhost = cloudprovider.Outlier{ResultsKey: "LocalHostNmapResults", IPs: []string{"0.0.0.0"}}
)

//CheckServiceBaselineChanges - check for service baseline changes and send changes to email alerts channel
func CheckServiceBaselineChanges(serviceChangeAlerts chan notifier.XMLEmailAlert, serviceChanges chan netscan.ServiceChanges, serviceChangesCounter chan int, mc *memcache.Client) {
	for changes := range serviceChanges {
		go compareTwoServiceScans(changes.ProviderResultstKey, changes.NewServiceScanResults, serviceChangeAlerts, serviceChangesCounter, mc)
	}
}

//CheckIPBaselineChange - check for IP Baseline Change and Send changes to outlier channel
func CheckIPBaselineChange(provider *cloudprovider.Provider, outliers chan cloudprovider.Outlier, alerts chan notifier.EmailAlert, mc *memcache.Client) {
	sips, err := data.GetIPSByProvider(mc, provider.IPKey)                                                           //Retrieve IP data from memcache
	nips := cloudprovider.Outlier{Key: provider.OutlierKey, ResultsKey: provider.ResultsKey, IPs: provider.GetIPs()} //Retrieve IP data from cloud
	baselineUpdate := fmt.Sprintf("%d IP(s) detected. IP(s): %s", len(nips.IPs), nips)                               //Baseline Update Alert
	subject := fmt.Sprintf("New IP Baseline for %s", provider.ProviderName)
	if err != nil {
		miss := err.Error() == "memcache: cache miss" //Cache Miss?
		if miss {
			data.StoreIPSByProvider(mc, provider)                                                                      //Store the new IP baseline data
			outliers <- nips                                                                                           //Scan the new set of IPs in the baseline
			alerts <- notifier.EmailAlert{Body: baselineUpdate, Subject: subject, ProviderName: provider.ProviderName} //Send Baseline Update alert
		} else {
			log.Fatal(err)
			outliers <- localhost                                                                                       //Scan myself :-). Do Nothing
			alerts <- notifier.EmailAlert{Body: "Localhost", Subject: "IP Baseline Changes", ProviderName: "Localhost"} //Send Localhost scan alert
		}
	} else if len(sips) == 0 { //Empty IP for provider?
		log.Printf("IP Baseline for %s doesnt exist.", provider.ProviderName)
		data.StoreIPSByProvider(mc, provider)                                                                      //Store the new IP Data
		outliers <- nips                                                                                           //Scan the new set of IPs in the baseline
		alerts <- notifier.EmailAlert{Body: baselineUpdate, Subject: subject, ProviderName: provider.ProviderName} //Send Baseline Update alert
	} else {
		compareTwoIPSets(sips, provider, outliers, alerts) //Compare Memcached IP Data with newly fetched IP Data
	}
}

func compareTwoIPSets(currentIPBaseline []string, provider *cloudprovider.Provider, outliers chan cloudprovider.Outlier, alerts chan notifier.EmailAlert) {
	sort.Strings(currentIPBaseline)
	sort.Strings(provider.GetIPs())
	b := reflect.DeepEqual(currentIPBaseline, provider.GetIPs())
	if b == true {
		log.Printf("IP Baseline for %s has not changed. \n", provider.ProviderName)
		outliers <- localhost                                                                                          //Scan myself :-). Do Nothing
		alerts <- notifier.EmailAlert{Body: "Localhost", Subject: "No IP Baseline Changes", ProviderName: "Localhost"} //Send Localhost scan alert
	} else {
		getIPBaselineOutliers(currentIPBaseline, provider.GetIPs(), provider, outliers, alerts)
		log.Printf("IP Baseline for %s has changed. \n", provider.ProviderName)
	}
}

func getIPBaselineOutliers(currentIPBaseline []string, newIPs []string, provider *cloudprovider.Provider, outliers chan cloudprovider.Outlier, alerts chan notifier.EmailAlert) {
	ipSet1 := strset.New(currentIPBaseline...)
	ipSet2 := strset.New(newIPs...)
	ipSet3 := strset.SymmetricDifference(ipSet1, ipSet2)

	//Send email alerts
	ipSetStrings := fmt.Sprintf("%d IP(s) detected. IP(s): %s", len(ipSet3.List()), ipSet3.List())
	subject := fmt.Sprintf("IP Changes for %s", provider.ProviderName)
	alerts <- notifier.EmailAlert{Body: ipSetStrings, Subject: subject, ProviderName: provider.ProviderName}
	outliers <- cloudprovider.Outlier{ResultsKey: provider.ResultsKey, IPs: ipSet3.List()} //Compare IP Baselines
}

//CompareTwoServiceScans - To check for changes in service basleines
func compareTwoServiceScans(resultsKey string, newServiceChanges []byte, serviceChangeAlerts chan notifier.XMLEmailAlert, serviceChangesCounter chan int, mc *memcache.Client) {
	currentServiceBaselineResults, err := data.GetNmapScanResults(mc, resultsKey) //Get Service Baseline
	if err != nil {
		miss := err.Error() == "memcache: cache miss" //Cache Miss?
		if miss {
			data.StoreNmapScanResults(mc, resultsKey, newServiceChanges) //Store Nmap Result Data
			//newxml := tokenizeXML(newServiceChanges)
			subject := fmt.Sprintf("New Service Baseline for %s", resultsKey)
			serviceChangeAlerts <- notifier.XMLEmailAlert{Body: newServiceChanges, Subject: subject, ProviderName: resultsKey}
		} else {
			log.Fatal(err)
		}
	} else {
		baseline, berr := netscan.Parse(bytes.TrimSpace(currentServiceBaselineResults)) //Parse the Service Baseline
		if berr != nil {
			log.Fatal(berr)
		}

		changes, cerr := netscan.Parse(bytes.TrimSpace(newServiceChanges)) //Parse new scan results
		if cerr != nil {
			log.Fatal(cerr)
		}

		log.Println("Drawing comparisons...")
		if diff := reflect.DeepEqual(baseline.Hosts, changes.Hosts); diff != true { //Compare Results
			//xml := tokenizeXML(newServiceChanges)
			subject := fmt.Sprintf("Service Changes for %s", resultsKey)
			serviceChangeAlerts <- notifier.XMLEmailAlert{Body: newServiceChanges, Subject: subject, ProviderName: resultsKey}
		} else {
			log.Printf("There are no service changes for %s: ", resultsKey)
		}
	}

	serviceChangesCounter <- 1
}

func tokenizeXML(scanResultsBytes []byte) []byte {
	buf := new(bytes.Buffer)
	d := xml.NewDecoder(strings.NewReader(string(scanResultsBytes)))
	e := xml.NewEncoder(buf)
	e.Indent("", " ")
tokenize:
	for {
		tok, err := d.Token()
		switch {
		case err == io.EOF:
			e.Flush()
			break tokenize
		case err != nil:
			log.Fatal(err)
		}
		e.EncodeToken(tok)
	}
	newxml := buf.Bytes()
	return newxml
}

//TrackServiceChanges - numberOfServiceChanges is equal numberOfProviders
func TrackServiceChanges(numberOfAlerts int, serviceChanges chan netscan.ServiceChanges, counter chan int) {
	c := 0
	for {
		select {
		case i := <-counter:
			c = c + i
			log.Printf("Service Change #%d ...", c)
			if c == numberOfAlerts {
				close(serviceChanges)
				break
			}
		}
	}
}
