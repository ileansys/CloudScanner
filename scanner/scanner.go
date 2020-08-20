package scanner

import (
	"log"

	"github.com/Ullaakut/nmap"
	"ileansys.com/cloudiff/data"
)

//ServiceScan - Conduct a Service Scan
func ServiceScan(providerResultsKey string, ipList []string, counter chan int) {
	var (
		resultBytes []byte
		errorBytes  []byte
	)
	log.Printf("Scanning outliers... %s", ipList)

	scanner, err := nmap.NewScanner(
		nmap.WithTargets(ipList...),
		nmap.WithPorts("443"),
	)

	if err != nil {
		log.Fatalf("unable to create nmap scanner: %v", err)
	}

	// Executes asynchronously, allowing results to be streamed in real time.
	if err := scanner.RunAsync(); err != nil {
		panic(err)
	}

	// Connect to stdout of scanner.
	stdout := scanner.GetStdout()

	// Connect to stderr of scanner.
	stderr := scanner.GetStderr()

	// Goroutine to watch for stdout and print to screen.
	for stdout.Scan() {
		log.Println(stdout.Text())
		resultBytes = append(resultBytes, stdout.Bytes()...)
	}

	// Goroutine to watch for stderr and print to screen.
	for stderr.Scan() {
		errorBytes = append(errorBytes, stderr.Bytes()...)
	}

	// Blocks until the scan has completed.
	if err := scanner.Wait(); err != nil {
		panic(err)
	}

	log.Printf("Storing scan results using %s ...", providerResultsKey)
	err = data.StoreNmapScanResults(providerResultsKey, resultBytes)
	if err != nil {
		log.Fatal(err)
	}

	counter <- 1

}
