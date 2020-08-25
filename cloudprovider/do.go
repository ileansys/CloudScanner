package cloudprovider

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/digitalocean/godo"
	"gopkg.in/yaml.v2"
)

//GetIPs - returns a list of DigitalOceans Public IPs (Both Floating IPs and Droplet IPs)
func getDOIPs() []string {

	ips := make([]string, 0) //cumulative list of droplet and floating ips

	doAccessToken := getAccessToken()
	client := godo.NewFromToken(doAccessToken)
	ctx := context.TODO()

	floatingIPlist, floatingIPErr := getFloatingIPList(ctx, client) //get a list of floatingips
	if floatingIPErr != nil {
		log.Fatal(floatingIPErr)
	}

	droptletList, dropletErr := getDropletList(ctx, client) //get a list of droplets
	if dropletErr != nil {
		log.Fatal(dropletErr)
	}

	for _, floatingIP := range floatingIPlist { //loop through the FloatingIPList and extract the IPs
		ips = append(ips, floatingIP.IP)
	}

	for _, droplet := range droptletList {
		ips = append(ips, droplet.Networks.V4[0].IPAddress) //loop through the dropletList and extract the IPs
	}

	return ips
}

func getAccessToken() string {

	log.Printf("Fetching DO Config File %s ...", os.Getenv("DIGITALOCEAN_DOCTL_CONFIG"))
	m := make(map[string]string)
	yamlFile, err := ioutil.ReadFile(os.Getenv("DIGITALOCEAN_DOCTL_CONFIG"))
	unmarshalError := yaml.Unmarshal(yamlFile, &m)

	if err != nil {
		fmt.Printf("Error parsing YAML file: %s\n", unmarshalError)
	}
	return m["access-token"]
}

//getDropletList - Fetch Droplet list from DigitalOcean
func getDropletList(ctx context.Context, client *godo.Client) ([]godo.Droplet, error) {

	log.Println("Fetching DO Droplet IPs...")
	// create a list to hold our droplets
	list := []godo.Droplet{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}
	for {
		droplets, resp, err := client.Droplets.List(ctx, opt)
		if err != nil {
			return nil, err
		}

		// append the current page's droplets to our list
		list = append(list, droplets...)

		// if we are at the last page, break out the for loop
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		// set the page we want for the next request
		opt.Page = page + 1
	}

	return list, nil
}

//getFloatingIPList - Fetch FloatingIP list from DigitalOcean
func getFloatingIPList(ctx context.Context, client *godo.Client) ([]godo.FloatingIP, error) {

	log.Println("Fetching DO Floating IPs...")
	// create a list to hold our droplets
	list := []godo.FloatingIP{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}
	for {
		floatingips, resp, err := client.FloatingIPs.List(ctx, opt)
		if err != nil {
			return nil, err
		}

		// append the current page's droplets to our list
		list = append(list, floatingips...)

		// if we are at the last page, break out the for loop
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		// set the page we want for the next request
		opt.Page = page + 1
	}

	return list, nil
}
