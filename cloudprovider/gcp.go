package cloudprovider

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

//GetIPs - List of compute addresses
func getGCPIPs() []string {

	err := godotenv.Load("/home/cloudiff/.env") //Load Environmental Variables
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var (
		projectID     = os.Getenv("GOOGLE_PROJECTID")
		gcpConfigFile = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") //Get GCP credentials from config files
	)
	log.Println("Fetching GCP IPs...")
	ctx := context.Background()
	computeService, err := compute.NewService(ctx, option.WithCredentialsFile(gcpConfigFile))
	if err != nil {
		log.Fatal(err)
	}

	listOfIPAddresses := make([]string, 0)
	req := computeService.Instances.AggregatedList(projectID)
	var publicIP string
	if err := req.Pages(ctx, func(page *compute.InstanceAggregatedList) error {
		for _, instancesScopedList := range page.Items {
			for _, instance := range instancesScopedList.Instances {
				publicIP = instance.NetworkInterfaces[0].AccessConfigs[0].NatIP
				if publicIP != "" {
					listOfIPAddresses = append(listOfIPAddresses, publicIP)
				}
			}
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	return listOfIPAddresses
}
