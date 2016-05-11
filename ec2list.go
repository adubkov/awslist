package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
	"strings"
)

var (
	// @readonly
	defaultRegion        = "us-west-1"
	awsErrError          = "[ERROR] %+v %+v %+v"
	awsErrRequestFailure = "[ERROR] %+v %+v %+v %+v"
)

type EC2List struct {
	EC2     *ec2.EC2
	Profile *Profile
	Region  string
}

// Returns a pointer to a new EC2List object
func NewEC2List(profile *Profile, region string) *EC2List {
	// If region is not specified, connect to default one - us-west-1
	if region == "" {
		region = defaultRegion
	}

	config := aws.Config{
		Region:      aws.String(region),
		Credentials: profile.Credentials,
	}

	return &EC2List{
		EC2:     ec2.New(session.New(), &config),
		Profile: profile,
		Region:  region,
	}
}

// Print list of instances in a format:
// {id},{name},{private_ip},{instance_size},{public_ip},{region},{account}
func (c *EC2List) fetchInstances(token string, channel chan ec2.Instance) {
	defer wg.Done()
	// Prepare request
	params := &ec2.DescribeInstancesInput{
		DryRun: aws.Bool(false),
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
		MaxResults: aws.Int64(1000),
		// Next page token
		NextToken: aws.String(token),
	}

	// Get list of ec2 instances
	res, err := c.EC2.DescribeInstances(params)
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
	for _, r := range res.Reservations {
		for _, i := range r.Instances {

			channel <- *i

			// If there is no tag "Name", return "None"
			name := "None"
			for _, keys := range i.Tags {
				if *keys.Key == "Name" {
					name = *keys.Value
				}
			}

			instance_string := []*string{
				i.InstanceId,
				&name,
				i.PrivateIpAddress,
				i.InstanceType,
				i.PublicIpAddress,
				&c.Region,
				&c.Profile.Name,
				i.KeyName,
				i.ImageId,
				i.Placement.AvailabilityZone,
				i.SubnetId,
				i.VpcId,
			}

			if i.IamInstanceProfile != nil {
				instance_string = append(instance_string, i.IamInstanceProfile.Arn)
			}

			//log.Printf("%+v", i)

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
		go c.fetchInstances(*res.NextToken, channel)
	}
}

// Goroutine to print instances from all regions within account
func fetchAccountInstances(profile *Profile, regions []string, channel chan ec2.Instance) {
	defer wg.Done()
	for _, region := range regions {
		wg.Add(1)
		go NewEC2List(profile, region).fetchInstances("", channel)
	}
}

// Returns all instances from all regions and accounts
func fetchInstances() []ec2.Instance {
	var profile *Profile

	// Clear output_buffer
	output_buffer = []string{}
	instances = []ec2.Instance{}
	ch_instances := make(chan ec2.Instance)

	// Run go routines to print instances
	for _, profile_name := range profiles {
		// If we didn't load regions already, then fill regions slice
		wg.Add(1)
		profile = NewProfile(profile_name)
		go fetchAccountInstances(profile, regions, ch_instances)
	}

	// Retreive results from all goroutines over channel
	go func() {
		for i := range ch_instances {
			instances = append(instances, i)
		}
	}()

	// Wait until receive info about all instances
	wg.Wait()

	// Close instances chaneel
	close(ch_instances)

	// Resize and fill screen buffer with output data
	screen_buffer = make([]string, len(output_buffer), (cap(output_buffer)+1)*2)
	copy(screen_buffer, output_buffer)
	return instances
}
