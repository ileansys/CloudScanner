package scanner

import (
	"log"
	"sync"

	"github.com/Ullaakut/nmap"
)

//ServiceScan - Conduct a Service Scan
func ServiceScan(ipList []string, wg *sync.WaitGroup, results chan []byte) {
	defer wg.Done()
	log.Printf("Scanning outliers... %s", ipList)
	var (
		errorBytes []byte
	)

	s, err := nmap.NewScanner(
		nmap.WithTargets(ipList...),
		nmap.WithMostCommonPorts(5000),
		nmap.WithAggressiveScan(),
		nmap.WithOpenOnly(),
	)
	if err != nil {
		log.Fatalf("unable to create nmap scanner: %v", err)
	}

	// Executes asynchronously, allowing results to be streamed in real time.
	if err := s.RunAsync(); err != nil {
		panic(err)
	}

	// Connect to stdout of scanner.
	stdout := s.GetStdout()

	// Connect to stderr of scanner.
	stderr := s.GetStderr()

	// Goroutine to watch for stdout and print to screen. Additionally it stores
	// the bytes intoa variable for processiing later.
	go func() {
		for stdout.Scan() {
			//log.Println(stdout.Text())
			///resultBytes = append(resultBytes, stdout.Bytes()...)
			results <- stdout.Bytes()
		}
	}()

	// Goroutine to watch for stderr and print to screen. Additionally it stores
	// the bytes intoa variable for processiing later.
	go func() {
		for stderr.Scan() {
			errorBytes = append(errorBytes, stderr.Bytes()...)
		}
	}()

	// Blocks main until the scan has completed.
	if err := s.Wait(); err != nil {
		panic(err)
	}

	// Parsing the results into corresponding structs.
	// result, err := nmap.Parse(resultBytes)

	// Parsing the results into the NmapError slice of our nmap Struct.
	// result.NmapErrors = strings.Split(string(errorBytes), "\n")
	// if err != nil {
	// 	panic(err)
	// }

	// // Use the results to print an example output
	// for _, host := range result.Hosts {
	// 	if len(host.Ports) == 0 || len(host.Addresses) == 0 {
	// 		continue
	// 	}

	// 	log.Printf("Host %q:\n", host.Addresses[0])

	// 	for _, port := range host.Ports {
	// 		log.Printf("\tPort %d/%s %s %s\n", port.ID, port.Protocol, port.State, port.Service.Name)
	// 	}
	// }
}
