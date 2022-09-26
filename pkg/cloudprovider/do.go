package cloudprovider

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/projectdiscovery/cloudlist/pkg/inventory"
	"github.com/projectdiscovery/cloudlist/pkg/schema"
	"gopkg.in/yaml.v2"
)

//GetIPs - returns a list of DigitalOceans Public IPs (Both Floating IPs and Droplet IPs)
func getDOIPs() []string {

	ips := make([]string, 0) //cumulative list of droplet and floating ips
	doAccessToken := getAccessToken()

	inventory, err := inventory.New(schema.Options{
		schema.OptionBlock{"provider": "digitalocean", "digitalocean_token": doAccessToken},
	})

	if err != nil {
		log.Fatalf("%s\n", err)
	}

	for _, provider := range inventory.Providers {
		resources, err := provider.Resources(context.Background())
		fmt.Println(resources)
		if err != nil {
			log.Fatalf("%s\n", err)
		}
		for _, resource := range resources.Items {
			fmt.Println(resource.PublicIPv4)
			ips = append(ips, resource.PublicIPv4)
		}
	}

	return ips
}

func getAccessToken() string {

	dirname, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	err = godotenv.Load(dirname + "/" + ".env") //Load Environmental Variables
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Printf("Fetching DO Config File %s ...", os.Getenv("DIGITALOCEAN_DOCTL_CONFIG"))
	m := make(map[string]string)
	yamlFile, err := ioutil.ReadFile(os.Getenv("DIGITALOCEAN_DOCTL_CONFIG"))
	unmarshalError := yaml.Unmarshal(yamlFile, &m)

	if err != nil {
		fmt.Printf("Error parsing YAML file: %s\n", unmarshalError)
	}
	return m["access-token"]
}
