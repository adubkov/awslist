package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/vaughan0/go-ini"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	// @readonly
	LoadProfilesList     = awserr.New("LoadProfilesList", "failed to load profiles list from credentials file", nil)
	defaultRegion        = "us-west-1"
	awsErrError          = "[ERROR] %+v %+v %+v"
	awsErrRequestFailure = "[ERROR] %+v %+v %+v %+v"
)

type Profile struct {
	Name string
	//Credentials *credentials.Credentials
	Region string
}

type EC2List struct {
	Connection *ec2.EC2
	Region     string
	Profile    Profile
}

// ListProfiles return list of aws profiles from credentials file
func ListProfiles() ([]string, error) {
	filename := filepath.Join(os.Getenv("HOME"), ".aws", "config")

	// Parse credentials file
	config, err := ini.LoadFile(filename)
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
func NewAWSList(profile, region string) *AWSList {
	// Load aws credentials from ~/.aws/config file
	filename := filepath.Join(os.Getenv("HOME"), ".aws", "config")
	creds := credentials.NewSharedCredentials(filename, profile)

	// If region is not specified, connect to default one - us-west-1
	if region == "" {
		region = defaultRegion
	}

	// If profile name specified, we extract account name from it.
	if len(profile) > 8 {
		profile = profile[8:]
	}

	config := aws.Config{
		Region:      aws.String(region),
		Credentials: creds,
	}

	return &AWSList{
		EC2:      ec2.New(session.New(), &config),
		Region:   region,
		Account:  profile,
		Filename: filename,
	}
}

// ListRegions returns list of aws regions
func (a *AWSList) ListRegions() ([]Profile, error) {
	// Prepare request
	params := &ec2.DescribeRegionsInput{
		DryRun: aws.Bool(false),
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
		return []Profile{}, err
	}

	var profiles []Profile
	// Extract regions name from result and fill regions slice with them
	for _, region := range res.Regions {
		profiles = append(profiles, Profile{
			Name:   a.Account,
			Region: *region.RegionName,
		})
	}

	return profiles, nil
}

// ListInstances print list of instances in a format:
// {id},{name},{private_ip},{instance_size},{public_ip},{region},{account}
func (a *AWSList) ListInstances(token string) {
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
				instance.InstanceId,
				&name,
				instance.PrivateIpAddress,
				instance.InstanceType,
				instance.PublicIpAddress,
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
