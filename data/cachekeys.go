package data

//IPBaseline - type for accessing IPBaseline Keys
type IPBaseline int

const (
	//GCPIPCountKey - memcached key for getting GCP IP Count
	GCPIPCountKey IPBaseline = iota
	//AWSIPCountKey - memcached key for getting AWS IP Count
	AWSIPCountKey
	//DOIPCountKey - memcached key for getting DO IP Count
	DOIPCountKey
	//GCPIPsKey - memcached key for getting GCP IPs
	GCPIPsKey
	//AWSIPsKey - memcached key for getting AWS IPs
	AWSIPsKey
	//DOIPsKey - memcached key for getting DO IPs
	DOIPsKey
	//TESTIPsKey - memcached key for testing IPs
	TESTIPsKey
	//DONmapResultsKey - memcached key for the latest nmap results for DO
	DONmapResultsKey
	//AWSNmapResultsKey - memcached key for the latest nmap results for AWS
	AWSNmapResultsKey
	//GCPNmapResultsKey - memcached key for the latest nmap results for GCP
	GCPNmapResultsKey
	//LocalHostNmapResults - memcached key for localhost nmap results
)

func (ip IPBaseline) String() string {
	return [...]string{"GCPIPCount", "AWSIPCount", "DOIPCount", "GCPIPs", "AWSIPs", "DOIPs", "TESTIPs", "DONmapResults", "AWSNmapResults", "GCPNmapResults", "LocalHostNmapResults"}[ip]
}
