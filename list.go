package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
)

var (
	// @readonly
	errNoProfiles = awserr.New("NoProfiles", "There are no profiles found in aws credentials file.", nil)
)

// Returns list of aws regions
func fetchRegions() ([]string, error) {
	// Prepare request
	params := &ec2.DescribeRegionsInput{
		DryRun: aws.Bool(false),
	}

	if len(profiles) == 0 {
		return []string{}, errNoProfiles
	}

	profile := NewProfile(profiles[0])
	config := aws.Config{
		Region:      aws.String(defaultRegion),
		Credentials: profile.Credentials,
	}
	con := ec2.New(session.New(), &config)

	// Get aws regions
	res, err := con.DescribeRegions(params)
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
