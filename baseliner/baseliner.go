package baseliner

import (
	"fmt"
	"log"
	"reflect"
	"sort"

	"ileansys.com/cloudiff/notifier"

	"ileansys.com/cloudiff/cloudprovider"

	"ileansys.com/cloudiff/data"

	"github.com/scylladb/go-set/strset"
)

var (
	localhost = cloudprovider.Outlier{ResultsKey: "LocalHostNmapResults", IPs: []string{"0.0.0.0"}}
)

//CheckIPBaselineChange - check for IP Baseline Change and Send changes to outlier channel
func CheckIPBaselineChange(provider *cloudprovider.Provider, outliers chan cloudprovider.Outlier) {
	sips, err := data.GetIPSByProvider(provider.IPKey)                                     //stored IPs Data from memcache
	nips := cloudprovider.Outlier{ResultsKey: provider.ResultsKey, IPs: provider.GetIPs()} //new IPs
	if err != nil {
		miss := err.Error() == "memcache: cache miss" //If there is a cache miss
		if miss {
			data.StoreIPSByProvider(provider) //Store the new IP baseline data
			outliers <- nips                  //Scan the new set of IPs in the baseline
		} else {
			log.Fatal(err)
			outliers <- localhost //Scan myself :-). Do Nothing
		}
	} else if len(sips) == 0 { //if there is empty IP Data for a provider
		log.Printf("IP Baseline for %s doesnt exist.", provider.ProviderName)
		data.StoreIPSByProvider(provider) //Store the new IP Data
		outliers <- nips                  //Scan the new set of IPs in the baseline
	} else {
		compareTwoIPSets(sips, provider, outliers) //Compare Memcached IP Data with newly fetched IP Data
	}
}

func compareTwoIPSets(currentIPBaseline []string, provider *cloudprovider.Provider, outliers chan cloudprovider.Outlier) {
	sort.Strings(currentIPBaseline)
	sort.Strings(provider.GetIPs())
	b := reflect.DeepEqual(currentIPBaseline, provider.GetIPs())
	if b == true {
		log.Printf("IP Baseline for %s has not changed. \n", provider.ProviderName)
		outliers <- localhost //Scan myself :-). Do Nothing
	} else {
		getIPBaselineOutliers(currentIPBaseline, provider.GetIPs(), provider.ResultsKey, outliers)
		log.Printf("IP Baseline for %s has changed. \n", provider.ProviderName)
	}
}

func getIPBaselineOutliers(currentIPBaseline []string, newIPs []string, providerResultKey string, outliers chan cloudprovider.Outlier) {
	ipSet1 := strset.New(currentIPBaseline...)
	ipSet2 := strset.New(newIPs...)
	ipSet3 := strset.SymmetricDifference(ipSet1, ipSet2)

	//Send email alerts
	ipSetStrings := fmt.Sprintf("%d IP(s) detected. IP(s): %s", len(ipSet3.List()), ipSet3.List())
	notifier.Send(ipSetStrings, providerResultKey)
	outliers <- cloudprovider.Outlier{ResultsKey: providerResultKey, IPs: ipSet3.List()} //Compare IP Baselines
}
