package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//GetAWSIPs - Get AWS IPs
func GetAWSIPs() []string {
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
	var listOfIPAddresessPerRegion []string
	for _, region := range results.Regions {
		listOfIPAddresessPerRegion = getAWSIPsByRegion(*region.RegionName)
		listOfIPAddresses = append(listOfIPAddresses, listOfIPAddresessPerRegion...)
	}

	return listOfIPAddresses
}

//GetAWSIPsByRegion - Fetch AWS IPs
func getAWSIPsByRegion(region string) []string {

	// Load session from shared config
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	// Create new EC2 client
	ec2Svc := ec2.New(sess)

	listOfIPAddresses := make([]string, 0)
	// Call to get detailed information on each instance
	result, err := ec2Svc.DescribeNetworkInterfaces(nil)
	if err != nil {
		fmt.Println("Error", err)
	} else {
		for _, networkInterface := range result.NetworkInterfaces {
			if networkInterface.Association != nil {
				listOfIPAddresses = append(listOfIPAddresses, *networkInterface.Association.PublicIp)
			}
		}
	}

	return listOfIPAddresses
}
