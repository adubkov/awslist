package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
	"strings"
	"time"
)

var (
	// @readonly
	awsErrError          = "[ERROR] %+v %+v %+v"
	awsErrRequestFailure = "[ERROR] %+v %+v %+v %+v"
)

type EC2List struct {
	Credentials *credentials.Credentials
	Account     string
}

type Instance struct {
	Instance ec2.Instance
	Account  string
}

// Returns a pointer to a new EC2List object
func NewEC2List(account string) *EC2List {
	creds := assumeRole(fmt.Sprintf(roleArnTemplate, account, roleName))
	return &EC2List{Account: account, Credentials: creds}
}

// Print instances from all regions within account
func (c *EC2List) fetchInstances(channel chan Instance) {
	defer ec2_wg.Done()
	var next_token string
	for _, region := range regions {
		ec2_wg.Add(1)
		go c.fetchRegionInstances(region, next_token, channel)
	}
}

// Print and send to channel list of instances.
func (c *EC2List) fetchRegionInstances(region, next_token string, channel chan Instance) {
	defer ec2_wg.Done()

	// Connect to region
	config := aws.Config{
		Region:      aws.String(region),
		Credentials: c.Credentials,
		MaxRetries:  aws.Int(20),
	}
	con := ec2.New(session.New(), &config)

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
		NextToken: aws.String(next_token),
	}

	// Get list of ec2 instances
	res, err := con.DescribeInstances(params)
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

	// Send instances to channel
	for _, r := range res.Reservations {
		for _, i := range r.Instances {
			if *service {
				channel <- Instance{Instance: *i, Account: c.Account}
			} else {
				s := formatInstanceOutput(c.Account, *i)
				fmt.Printf("%s", s)
			}
		}
	}

	// If there are more instances repeat request with a token
	if res.NextToken != nil {
		ec2_wg.Add(1)
		go c.fetchRegionInstances(region, *res.NextToken, channel)
	}
}

// Returns all instances from all regions and accounts
func fetchInstances() []Instance {
	var instances []Instance

	ch_instances := make(chan Instance)
	defer close(ch_instances)

	// Run go routines to print instances
	for _, account := range accounts {
		// If we didn't load regions already, then fill regions slice
		ec2_wg.Add(1)
		go NewEC2List(account).fetchInstances(ch_instances)
	}

	// Retreive results from all goroutines over channel
	go func() {
		for i := range ch_instances {
			instances = append(instances, i)
		}
	}()

	// Wait until receive info about all instances
	ec2_wg.Wait()

	return instances
}

// Returns formatted string with instance information.
// This function is for backward compatibility with v1.
func formatInstanceOutputV1(account string, i ec2.Instance) string {
	// If there is no tag "Name", return ""
	var name string
	for _, keys := range i.Tags {
		switch strings.ToLower(*keys.Key) {
		case "name":
			name = *keys.Value
		}
	}

	instance := []*string{
		i.InstanceId,
		&name,
		i.PrivateIpAddress,
		i.InstanceType,
		i.PublicIpAddress,
		i.Placement.AvailabilityZone,
		&account,
	}

	return makeFormattedOutput(instance)
}

// Returns formatted string with instance information.
func formatInstanceOutput(account string, i ec2.Instance) string {
	// If there is no tag "Name", return "None"
	var name, team, autoscaling_group_name, instance_profile string
	for _, keys := range i.Tags {
		switch strings.ToLower(*keys.Key) {
		case "name":
			name = *keys.Value
		case "team":
			team = *keys.Value
		case "aws:autoscaling:groupname":
			autoscaling_group_name = *keys.Value
		}
	}

	if i.IamInstanceProfile != nil {
		iam_parts := strings.Split(*i.IamInstanceProfile.Arn, "/")
		instance_profile = iam_parts[len(iam_parts)-1]
	}

	launch_time := i.LaunchTime.Format(time.RFC3339)

	instance := []*string{
		i.InstanceId,
		&name,
		&team,
		i.PrivateIpAddress,
		i.PublicIpAddress,
		&autoscaling_group_name,
		i.Placement.AvailabilityZone,
		i.InstanceType,
		&account,
		i.KeyName,
		i.ImageId,
		i.SubnetId,
		i.VpcId,
		&instance_profile,
		&launch_time,
	}

	return makeFormattedOutput(instance)
}
