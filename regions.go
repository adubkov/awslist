package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var (
	// @readonly
	errNoProfiles = awserr.New("NoProfiles", "There are no profiles found in aws credentials file.", nil)
)

// Returns list of aws regions
func fetchRegions() ([]string, error) {
	var regions []string

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
