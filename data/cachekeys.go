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
)

func (ip IPBaseline) String() string {
	return [...]string{"GCPIPCount", "AWSIPCount", "DOIPCount", "GCPIPs", "AWSIPs", "DOIPs"}[ip]
}
