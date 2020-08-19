package cloudprovider

//CloudProvider interface
type CloudProvider interface {
	Init()
	GetIPs() []string //Get a slice of IPs
}

//Provider properties
type Provider struct {
	ProviderName string
	IPKey        string
	IPs          []string
}

//Init - Load IP Data
func (p Provider) Init() Provider {
	switch p.ProviderName {
	case "AWS":
		p.IPs = getAWSIPs()
	case "DO":
		p.IPs = getDOIPs()
	case "GCP":
		p.IPs = getGCPIPs()
	case "TEST":
		p.IPs = []string{"127.0.0.1", "192.168.0.1", "192.168.100.1"}
	}
	return p
}

//GetIPs - Return IP Data
func (p *Provider) GetIPs() []string {
	return p.IPs
}
