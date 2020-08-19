package baseliner

import (
	"log"
	"reflect"
	"sort"

	"ileansys.com/cloudiff/cloudprovider"

	"ileansys.com/cloudiff/data"

	"github.com/scylladb/go-set/strset"
)

//CheckIPBaselineChange - check for IP Baseline Change and Send changes to outlier channel
func CheckIPBaselineChange(provider *cloudprovider.Provider, outliers chan []string) {
	sips, err := data.GetIPSByProvider(provider.IPKey) //Get stored IP Data from memcache
	if err != nil {
		log.Fatal(err) //Need to determine cache miss
	} else if len(sips) == 0 { //if there is no IP Data Store new IP Data
		log.Printf("IP Baseline for %s doesnt exist.", provider.IPKey)
		data.StoreIPSByProvider(provider) //Store the new IP Data
	} else {
		compareTwoIPSets(sips, provider, outliers) //Compare Memcached IP Data with newly fetched IP Data
	}
}

func compareTwoIPSets(currentIPBaseline []string, provider *cloudprovider.Provider, outliers chan []string) {
	sort.Strings(currentIPBaseline)
	sort.Strings(provider.GetIPs())
	b := reflect.DeepEqual(currentIPBaseline, provider.GetIPs())
	if b == true {
		log.Printf("IP Baseline for %s has not changed. \n", provider.ProviderName)
		//time.Sleep()
	} else {
		getNewOutlyingIPs(currentIPBaseline, provider.GetIPs(), outliers)
		log.Printf("IP Baseline for %s has changed. \n", provider.ProviderName)
	}
}

func getNewOutlyingIPs(currentIPBaseline []string, newIPs []string, outliers chan []string) {
	ipSet1 := strset.New(currentIPBaseline...)
	ipSet2 := strset.New(newIPs...)
	ipSet3 := strset.SymmetricDifference(ipSet1, ipSet2)
	outliers <- ipSet3.List()
}
