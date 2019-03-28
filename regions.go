package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var (
	// @readonly
	defaultRegion = "us-west-1"
	errNoAccounts = awserr.New("NoAccounts", "AWS_ACCOUNTS must contains list of aws accounts id", nil)
)

// Returns list of aws regions
func fetchRegions() ([]string, error) {
	var regions []string

	if len(accounts) == 0 {
		return []string{}, errNoAccounts
	}

	// Use first account to fetch the regions
	creds = assumeRole(fmt.Sprintf(roleArnTemplate, accounts[0], roleName))

	config := aws.Config{
		Region:      aws.String(defaultRegion),
		Credentials: creds,
	}
	con := ec2.New(session.New(), &config)

	// Get aws regions
	res, err := con.DescribeRegions(&ec2.DescribeRegionsInput{})
	if err != nil {
		return []string{}, err
	}

	// Fill slice with regions
	for _, region := range res.Regions {
		regions = append(regions, *region.RegionName)
	}

	return regions, nil
}
