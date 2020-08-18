package data

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
)

var (
	memcachedServer string = "127.0.0.1:11211"
)

//StoreIPCountByProvider - Store IP Count for a specific cloud provider
func StoreIPCountByProvider(cloudProviderIPCountkey string, numberOfIPs int) {
	mc := memcache.New(memcachedServer)
	err := mc.Set(&memcache.Item{Key: cloudProviderIPCountkey, Value: []byte(strconv.Itoa(numberOfIPs))})
	if err != nil {
		log.Fatal(err)
	}
}

//GetIPCountByProvider - Get IP Count for a specific cloud provider
func GetIPCountByProvider(cloudProviderIPCountKey string) string {
	mc := memcache.New(memcachedServer)
	ipCount, err := mc.Get(cloudProviderIPCountKey)
	if err != nil {
		log.Fatal(err)
	}
	return string(ipCount.Value)
}

//StoreIPSByProvider - Store IPs based on cloud provider
func StoreIPSByProvider(cloudProviderJSONKey string, ips []string) {
	mc := memcache.New(memcachedServer)

	jsonIPSByteArray, err := json.Marshal(ips) //Marshall slice of IPs into JSON
	if err != nil {
		log.Fatal(err)
	}

	merr := mc.Set(&memcache.Item{Key: cloudProviderJSONKey, Value: jsonIPSByteArray}) //Store Marshalled Slice of IPs
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
