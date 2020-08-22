package baseliner

import (
	"fmt"
	"log"
	"reflect"
	"sort"

	"ileansys.com/cloudiff/cloudprovider"
	"ileansys.com/cloudiff/notifier"

	"ileansys.com/cloudiff/data"

	"github.com/scylladb/go-set/strset"
)

var (
	localhost = cloudprovider.Outlier{ResultsKey: "LocalHostNmapResults", IPs: []string{"0.0.0.0"}}
)

//CheckIPBaselineChange - check for IP Baseline Change and Send changes to outlier channel
func CheckIPBaselineChange(provider *cloudprovider.Provider, outliers chan cloudprovider.Outlier, alerts chan notifier.EmailAlert) {
	sips, err := data.GetIPSByProvider(provider.IPKey)                                                  //Retrieve IP data from memcache
	nips := cloudprovider.Outlier{ResultsKey: provider.ResultsKey, IPs: provider.GetIPs()}              //Retrieve IP data from cloud
	baselineUpdate := fmt.Sprintf("Baseline update. %d IP(s) detected. IP(s): %s", len(nips.IPs), nips) //Baseline Update Alert
	if err != nil {
		miss := err.Error() == "memcache: cache miss" //Cache Miss?
		if miss {
			data.StoreIPSByProvider(provider)                                                        //Store the new IP baseline data
			outliers <- nips                                                                         //Scan the new set of IPs in the baseline
			alerts <- notifier.EmailAlert{Body: baselineUpdate, ProviderName: provider.ProviderName} //Send Baseline Update alert
		} else {
			log.Fatal(err)
			outliers <- localhost                                                       //Scan myself :-). Do Nothing
			alerts <- notifier.EmailAlert{Body: "Localhost", ProviderName: "Localhost"} //Send Localhost scan alert
		}
	} else if len(sips) == 0 { //Empty IP for provider?
		log.Printf("IP Baseline for %s doesnt exist.", provider.ProviderName)
		data.StoreIPSByProvider(provider)                                                        //Store the new IP Data
		outliers <- nips                                                                         //Scan the new set of IPs in the baseline
		alerts <- notifier.EmailAlert{Body: baselineUpdate, ProviderName: provider.ProviderName} //Send Baseline Update alert
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
		outliers <- localhost                                                       //Scan myself :-). Do Nothing
		alerts <- notifier.EmailAlert{Body: "Localhost", ProviderName: "Localhost"} //Send Localhost scan alert
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
	alerts <- notifier.EmailAlert{Body: ipSetStrings, ProviderName: provider.ProviderName}
	outliers <- cloudprovider.Outlier{ResultsKey: provider.ResultsKey, IPs: ipSet3.List()} //Compare IP Baselines
}
