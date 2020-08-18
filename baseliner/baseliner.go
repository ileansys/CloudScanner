package baseliner

import (
	"log"
	"reflect"
	"sort"

	"ileansys.com/cloudiff/data"

	"github.com/scylladb/go-set/strset"
)

//CheckIPBaselineChange - check for ip baseline change and return outliers
func CheckIPBaselineChange(cloudProviderIPskey string, newIPs []string, outliers chan []string) {
	ips, err := data.GetIPSByProvider(cloudProviderIPskey)
	if err != nil {
		log.Fatal(err)
	} else {
		compareTwoIPSets(cloudProviderIPskey, ips, newIPs, outliers)
	}
}

func compareTwoIPSets(cloudProvider string, currentIPBaseline []string, newIPs []string, outliers chan []string) {

	sort.Strings(currentIPBaseline)
	sort.Strings(newIPs)
	b := reflect.DeepEqual(currentIPBaseline, newIPs)
	if b == true {
		log.Printf("IP Baseline for %s has not changed \n", cloudProvider)
		//time.Sleep()
	} else {
		getNewOutlyingIPs(currentIPBaseline, newIPs, outliers)
		log.Printf("IP Baseline for %s has changed: \n", cloudProvider)
	}
}

func getNewOutlyingIPs(currentIPBaseline []string, newIPs []string, outliers chan []string) {
	ipSet1 := strset.New(currentIPBaseline...)
	ipSet2 := strset.New(newIPs...)
	ipSet3 := strset.SymmetricDifference(ipSet1, ipSet2)
	outliers <- ipSet3.List()
}
