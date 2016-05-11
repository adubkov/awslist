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

// ListRegions returns list of aws regions
func (c *EC2List) ListRegions() ([]string, error) {
	// Prepare request
	params := &ec2.DescribeRegionsInput{
		DryRun: aws.Bool(false),
	}

	// Get aws regions
	res, err := c.EC2.DescribeRegions(params)
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

	var profiles []string
	// Extract regions name from result and fill regions slice with them
	for _, region := range res.Regions {
		profiles = append(profiles, *region.RegionName)
	}

	return profiles, nil
}

// ListInstances print list of instances in a format:
// {id},{name},{private_ip},{instance_size},{public_ip},{region},{account}
func (c *EC2List) ListInstances(token string) {
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
				instance.InstanceId,
				&name,
				instance.PrivateIpAddress,
				instance.InstanceType,
				instance.PublicIpAddress,
				&c.Region,
				&c.Profile.Name,
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
		go c.ListInstances(*res.NextToken)
	}
}
