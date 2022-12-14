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
	var wg sync.WaitGroup

	go func() {
		for ip := range ipChannel {
			listOfIPAddresses = append(listOfIPAddresses, ip) //wait for ip addresses
		}
		close(ipChannel)
	}()
	wg.Add(1)
	go getComputeInstanceIPs(&wg, ipChannel)
	wg.Add(1)
	go getForwardingRuleIPs(&wg, ipChannel)
	wg.Wait()
	return listOfIPAddresses //return the ip addresses

}

//GetIPs - List of compute addresses
func getComputeInstanceIPs(wg *sync.WaitGroup, ipChannel chan string) {
	defer wg.Done()

	dirname, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	err = godotenv.Load(dirname + "/" + ".env") //Load Environmental Variables
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
	if err := req.Pages(ctx, func(page *compute.InstanceAggregatedList) error {
		for _, instancesScopedList := range page.Items {
			for _, instance := range instancesScopedList.Instances {
				for _, accessConfig := range instance.NetworkInterfaces[0].AccessConfigs {
					if accessConfig.NatIP != "" {
						ipChannel <- accessConfig.NatIP
					}
				}
			}
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
}

func getForwardingRuleIPs(wg *sync.WaitGroup, ipChannel chan string) {
	defer wg.Done()

	dirname, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	err = godotenv.Load(dirname + "/" + ".env") //Load Environmental Variables
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
}
