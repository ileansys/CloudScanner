package cloudprovider

import (
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//GetIPs - Get AWS IPs
func getAWSIPs() []string {
	log.Println("Fetching AWS IPs...")
	// Load session from shared config
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	//Create an new EC2 client
	svc := ec2.New(sess)
	input := &ec2.DescribeRegionsInput{}
	results, err := svc.DescribeRegions(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	}

	listOfIPAddresses := make([]string, 0)
	ipChannel := make(chan string) //create an IP channel
	var wg sync.WaitGroup

	go func() {
		for ip := range ipChannel {
			listOfIPAddresses = append(listOfIPAddresses, ip) //wait for ip addresses from different regions
		}
	}()

	for _, region := range results.Regions {
		wg.Add(1)
		go getAWSIPsByRegion(*region.RegionName, &wg, ipChannel) //get IP from different regions and send them to ip channel
	}

	wg.Wait()
	close(ipChannel) //close ip channel

	return listOfIPAddresses //return the ip addresses
}

//GetAWSIPsByRegion - Fetch AWS IPs
func getAWSIPsByRegion(region string, wg *sync.WaitGroup, ipc chan string) {

	defer wg.Done()
	// Load session from shared config
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	// Create new EC2 client
	ec2Svc := ec2.New(sess)

	// Call to get detailed information on each instance
	result, err := ec2Svc.DescribeNetworkInterfaces(nil)
	if err != nil {
		fmt.Println("Error", err)
	} else {
		for _, networkInterface := range result.NetworkInterfaces {
			if networkInterface.Association != nil {
				ipc <- *networkInterface.Association.PublicIp //send IPs to IP channel
			}
		}
	}

}
