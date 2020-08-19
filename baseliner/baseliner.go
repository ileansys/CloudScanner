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
		miss := err.Error() == "memcache: cache miss" //Find cache miss
		if miss {
			data.StoreIPSByProvider(provider) //Store the new IP data
		} else {
			log.Fatal(err)
		}
	}

	if len(sips) == 0 { //if there is no IP Data Store new IP Data
		log.Printf("IP Baseline for %s doesnt exist.", provider.IPKey)
		data.StoreIPSByProvider(provider) //Store the new IP Data
	} else {
		compareTwoIPSets(sips, provider, outliers) //Compare Memcached IP Data with newly fetched IP Data
	}
	close(outliers)
}

func compareTwoIPSets(currentIPBaseline []string, provider *cloudprovider.Provider, outliers chan []string) {
	sort.Strings(currentIPBaseline)
	sort.Strings(provider.GetIPs())
	b := reflect.DeepEqual(currentIPBaseline, provider.GetIPs())
	if b == true {
		log.Printf("IP Baseline for %s has not changed. \n", provider.ProviderName)
		//time.Sleep()
	} else {
		getIPBaselineOutliers(currentIPBaseline, provider.GetIPs(), outliers)
		log.Printf("IP Baseline for %s has changed. \n", provider.ProviderName)
	}
}

func getIPBaselineOutliers(currentIPBaseline []string, newIPs []string, outliers chan []string) {
	ipSet1 := strset.New(currentIPBaseline...)
	ipSet2 := strset.New(newIPs...)
	ipSet3 := strset.SymmetricDifference(ipSet1, ipSet2)
	outliers <- ipSet3.List()
}
