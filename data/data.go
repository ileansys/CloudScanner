package data

import (
	"encoding/json"
	"log"

	"ileansys.com/cloudiff/cloudprovider"

	"github.com/bradfitz/gomemcache/memcache"
)

//StoreIPSByProvider - Store IPs based on cloud provider
func StoreIPSByProvider(mc *memcache.Client, provider *cloudprovider.Provider) error {
	jsonIPSByteArray, err := json.Marshal(provider.GetIPs()) //Marshall slice of IPs into JSON
	if err != nil {
		log.Fatal(err)
	}

	merr := mc.Set(&memcache.Item{Key: provider.IPKey, Value: jsonIPSByteArray}) //Store Marshalled Slice of IPs
	if merr != nil {
		log.Fatal(merr)
		return merr
	}
	return nil
}

//GetIPSByProvider - Get IPs for a specific cloud provider
func GetIPSByProvider(mc *memcache.Client, cloudProviderIPsKey string) ([]string, error) {
	ips, err := mc.Get(cloudProviderIPsKey)
	if err != nil {
		return nil, err
	}

	sliceOfIPs := make([]string, 0)
	err = json.Unmarshal(ips.Value, &sliceOfIPs)
	if err != nil {
		return nil, err
	}

	return sliceOfIPs, nil
}

//StoreNmapScanResults - Store Nmap Scan results
func StoreNmapScanResults(mc *memcache.Client, nmapResultsKey string, scanResults []byte) error {
	merr := mc.Set(&memcache.Item{Key: nmapResultsKey, Value: scanResults}) //Store Marshalled Slice of IPs
	if merr != nil {
		log.Fatal(merr)
		return merr
	}
	return nil
}

//GetNmapScanResults - Retrieve Nmap Scan results
func GetNmapScanResults(mc *memcache.Client, nmapResultsKey string) ([]byte, error) {

	results, err := mc.Get(nmapResultsKey)
	if err != nil {
		return nil, err
	}

	return results.Value, nil
}
