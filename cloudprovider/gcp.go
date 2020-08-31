package cloudprovider

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func getGCPIPs() []string {

	listOfIPAddresses := make([]string, 0)
	ipChannel := make(chan string) //create an IP channel
	counter := make(chan int)
	var wg sync.WaitGroup

	go func() {
		c := 0
		for {
			select {
			case i := <-counter:
				c = c + i
				if c == 2 {
					close(ipChannel)
					break
				}
			}
		}
	}()

	go func() {
		for ip := range ipChannel {
			listOfIPAddresses = append(listOfIPAddresses, ip) //wait for ip addresses
		}
	}()

	wg.Add(1)
	go getComputeInstanceIPs(&wg, ipChannel, counter)
	wg.Add(1)
	go getForwardingRuleIPs(&wg, ipChannel, counter)
	wg.Wait()

	return listOfIPAddresses //return the ip addresses

}

//GetIPs - List of compute addresses
func getComputeInstanceIPs(wg *sync.WaitGroup, ipChannel chan string, counter chan int) {

	err := godotenv.Load("/home/cloudiff/.env") //Load Environmental Variables
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var (
		projectID     = os.Getenv("GOOGLE_PROJECTID")
		gcpConfigFile = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") //Get GCP credentials from config files
	)
	log.Println("Fetching Compute Instance IPs...")
	ctx := context.Background()
	computeService, err := compute.NewService(ctx, option.WithCredentialsFile(gcpConfigFile))
	if err != nil {
		log.Fatal(err)
	}

	req := computeService.Instances.AggregatedList(projectID)
	var publicIP string
	if err := req.Pages(ctx, func(page *compute.InstanceAggregatedList) error {
		for _, instancesScopedList := range page.Items {
			for _, instance := range instancesScopedList.Instances {
				publicIP = instance.NetworkInterfaces[0].AccessConfigs[0].NatIP
				if publicIP != "" {
					ipChannel <- publicIP
				}
			}
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	counter <- 1
}

func getForwardingRuleIPs(wg *sync.WaitGroup, ipChannel chan string, counter chan int) {

	err := godotenv.Load("/home/cloudiff/.env") //Load Environmental Variables
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var (
		projectID     = os.Getenv("GOOGLE_PROJECTID")
		gcpConfigFile = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") //Get GCP credentials from config files
	)
	log.Println("Fetching Forwarding Rule IPs...")
	ctx := context.Background()
	computeService, err := compute.NewService(ctx, option.WithCredentialsFile(gcpConfigFile))
	if err != nil {
		log.Fatal(err)
	}

	req := computeService.ForwardingRules.AggregatedList(projectID)
	var publicIP string
	if err := req.Pages(ctx, func(page *compute.ForwardingRuleAggregatedList) error {
		for _, forwardingRuleScopedList := range page.Items {
			for _, forwardingRule := range forwardingRuleScopedList.ForwardingRules {
				publicIP = forwardingRule.IPAddress
				if publicIP != "" {
					ipChannel <- publicIP
				}
			}
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	counter <- 1
}
