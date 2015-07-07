package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/vaughan0/go-ini"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	// @readonly
	LoadProfilesList     = awserr.New("LoadProfilesList", "failed to load profiles list from credentials file", nil)
	rootHandlerMsg       = "[INFO][%s]: %s %s request from %s. %d instances was returned.\n"
	defaultRegion        = "us-west-1"
	awsErrError          = "[ERROR] %+v %+v %+v"
	awsErrRequestFailure = "[ERROR] %+v %+v %+v %+v"
	runHttpMsg           = "[INFO] Runing http server on port: %d"
	portMsg              = "Listen port"
	serviceMsg           = "Run as service"
	intervalMsg          = "Interval to pool data in seconds"
)

// AWSList provide interface to print list of instances in aws account
type AWSList struct {
	EC2 *ec2.EC2
	// ec2 AWS region to connect
	Region string
	// AWS account name
	Account string
	// Path to credentials file
	Filename string
}

// NewAWSList returns a pointer to a new AWSList object
func NewAWSList(profile string, region ...string) *AWSList {
	// Load aws credentials from ~/.aws/config file
	filename := filepath.Join(os.Getenv("HOME"), ".aws", "config")
	creds := credentials.NewSharedCredentials(filename, profile)

	// If region is not specified, connect to default one - us-west-1
	if region == nil {
		region = []string{defaultRegion}
	}

	// If profile name specified, we extract account name from it.
	if len(profile) > 8 {
		profile = profile[8:]
	}

	return &AWSList{
		EC2: ec2.New(
			&aws.Config{
				Credentials: creds,
				Region:      region[0],
			}),
		Region:   region[0],
		Account:  profile,
		Filename: filename,
	}
}

// ListProfiles return list of aws profiles from credentials file
func (a *AWSList) ListProfiles() ([]string, error) {
	// Parse credentials file
	config, err := ini.LoadFile(a.Filename)
	if err != nil {
		return []string{}, LoadProfilesList
	}

	// Fill profiles slice with list of profiles
	profiles := []string{}
	for profile := range config {
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// ListRegions returns list of aws regions
func (a *AWSList) ListRegions() ([]string, error) {
	// Prepare request
	params := &ec2.DescribeRegionsInput{
		DryRun: aws.Boolean(false),
	}

	// Get aws regions
	res, err := a.EC2.DescribeRegions(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			log.Printf(awsErrError, awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				log.Printf(awsErrRequestFailure, reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			log.Printf(err.Error())
		}
		return []string{}, err
	}

	var regions []string
	// Extract regions name from result and fill regions slice with them
	for _, region := range res.Regions {
		regions = append(regions, *region.RegionName)
	}

	return regions, nil
}

// ListInstances print list of instances in a format:
// {id},{name},{private_ip},{instance_size},{public_ip},{region},{account}
func (a *AWSList) ListInstances(token string) {
	defer wg.Done()
	// Prepare request
	params := &ec2.DescribeInstancesInput{
		DryRun: aws.Boolean(false),
		Filters: []*ec2.Filter{
			{
				// Return only "running" and "pending" instances
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("running"),
					aws.String("pending"),
				},
			},
		},
		// Maximum count instances on one result page
		MaxResults: aws.Long(1000),
		// Next page token
		NextToken: aws.String(token),
	}

	// Get list of ec2 instances
	res, err := a.EC2.DescribeInstances(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			log.Printf(awsErrError, awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				log.Printf(awsErrRequestFailure, reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			log.Printf(err.Error())
		}
	}

	// Extract instances info from result and print it
	for _, reservation := range res.Reservations {
		for _, instance := range reservation.Instances {

			// If there is no tag "Name", return "None"
			name := "None"
			for _, keys := range instance.Tags {
				if *keys.Key == "Name" {
					name = *keys.Value
				}
			}

			instance_string := []*string{
				instance.InstanceID,
				&name,
				instance.PrivateIPAddress,
				instance.InstanceType,
				instance.PublicIPAddress,
				&a.Region,
				&a.Account,
			}

			output_string := []string{}
			for _, str := range instance_string {
				if str == nil {
					output_string = append(output_string, "None")
				} else {
					output_string = append(output_string, *str)
				}
			}

			instance := strings.Join(output_string, ",")
			// If running in service mode, write in output buffer, else just print
			if *service {
				output_buffer = append(output_buffer, instance)
			} else {
				fmt.Printf("%s\n", instance)
			}
		}
	}

	// If there are more instances repeat request with a token
	if res.NextToken != nil {
		wg.Add(1)
		go a.ListInstances(*res.NextToken)
	}
}

// PrintInstances runs go routines to print instances from all regions within account
func PrintInstances(profile string, regions []string) {
	defer wg.Done()
	for _, region := range regions {
		wg.Add(1)
		go NewAWSList(profile, region).ListInstances("")
	}
}

// getInstances run go routines to print all instances from all regions and all accounts
func getInstances() {
	var regions []string

	// Clear output_buffer
	output_buffer = []string{}

	// Get list of profiles from ~/.aws/config file
	profiles, _ := NewAWSList("").ListProfiles()

	// Run go routines to print instances
	for _, profile := range profiles {
		// If we didn't load regions already, then fill regions slice
		if len(regions) == 0 {
			r, _ := NewAWSList(profile).ListRegions()
			for _, region := range r {
				regions = append(regions, region)
			}
		}

		wg.Add(1)
		go PrintInstances(profile, regions)
	}

	// Wait until receive info about all instances
	wg.Wait()

	// Resize and fill screen buffer with output data
	screen_buffer = make([]string, len(output_buffer), (cap(output_buffer)+1)*2)
	copy(screen_buffer, output_buffer)
}

// Root handler return screen buffer as a respond
func rootHandler(res http.ResponseWriter, req *http.Request) {
	log.Printf(rootHandlerMsg, req.Host, req.Method, req.URL, req.RemoteAddr,
		len(screen_buffer))
	statusCode := 200
	res.WriteHeader(statusCode)
	fmt.Fprintf(res, strings.Join(screen_buffer, "\n"))
}

// runHttpServer runs http listener on specific port
func runHttpServer(port int) {
	http.HandleFunc("/", rootHandler)
	log.Printf(runHttpMsg, port)
	sockaddr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(sockaddr, nil))
}

var output_buffer []string
var screen_buffer []string
var service *bool
var port *int
var interval *int
var counter int
var wg sync.WaitGroup

func main() {
	// Parse arguments
	port = flag.Int("port", 8080, portMsg)
	service = flag.Bool("service", false, serviceMsg)
	interval = flag.Int("interval", 30, intervalMsg)
	flag.Parse()

	// Get list of instances
	getInstances()

	// If specified service mode, run program as a service, and listen port
	if *service {
		ticker := time.NewTicker(time.Second * time.Duration(*interval))

		// Each 30 seconds (by default)
		go func() {
			for range ticker.C {
				// Get list of all instances
				getInstances()
			}
		}()

		// Run http server on specifig port
		runHttpServer(*port)
	}
}
