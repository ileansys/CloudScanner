package data

import (
	"encoding/json"
	"log"

	"ileansys.com/cloudiff/cloudprovider"

	"github.com/bradfitz/gomemcache/memcache"
)

var (
	memcachedServer string = "127.0.0.1:11211"
)

//StoreIPSByProvider - Store IPs based on cloud provider
func StoreIPSByProvider(provider *cloudprovider.Provider) {
	mc := memcache.New(memcachedServer)

	jsonIPSByteArray, err := json.Marshal(provider.GetIPs()) //Marshall slice of IPs into JSON
	if err != nil {
		log.Fatal(err)
	}

	merr := mc.Set(&memcache.Item{Key: provider.IPKey, Value: jsonIPSByteArray}) //Store Marshalled Slice of IPs
	if merr != nil {
		log.Fatal(merr)
	}
}

//GetIPSByProvider - Get IPs for a specific cloud provider
func GetIPSByProvider(cloudProviderIPsKey string) ([]string, error) {
	mc := memcache.New(memcachedServer)

	ips, err := mc.Get(cloudProviderIPsKey)
	if err != nil {
		return nil, err
	}

	sliceOfIPs := make([]string, 0)
	err = json.Unmarshal(ips.Value, &sliceOfIPs)
	if err != nil {
		log.Fatal(err)
	}

	return sliceOfIPs, nil
}
