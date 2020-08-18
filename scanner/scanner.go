package scanner

import (
	"log"
	"sync"

	"github.com/Ullaakut/nmap"
)

//ServiceScan - Conduct a Service Scan
func ServiceScan(ipList []string, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Scanning outliers...")
	scanner, err := nmap.NewScanner(
		nmap.WithTargets(ipList...),
		nmap.WithPorts("0-10000"),
		nmap.WithServiceInfo(),
		nmap.WithTimingTemplate(nmap.TimingAggressive),
		nmap.WithFilterHost(func(h nmap.Host) bool {
			// Filter out hosts with no open ports.
			for idx := range h.Ports {
				if h.Ports[idx].Status() == "open" {
					return true
				}
			}
			return false
		}),
	)
	if err != nil {
		log.Fatalf("unable to create nmap scanner: %v", err)
	}

	result, _, err := scanner.Run()
	if err != nil {
		log.Fatalf("nmap scan failed: %v", err)
	}

	for _, host := range result.Hosts {
		log.Printf("Host %s\n", host.Addresses[0])

		for _, port := range host.Ports {
			log.Printf("\tPort %d open with protocol %s", port.ID, port.Protocol)
		}
	}
}
