package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
	"log"
)

var (
// @readonly
)

type ElbList struct {
	Profile *Profile
}

type Elb struct {
	Elb     elb.LoadBalancerDescription
	Profile Profile
}

// Returns a pointer to a new EC2List object
func NewElbList(profile *Profile) *ElbList {
	return &ElbList{Profile: profile}
}

// Print instances from all regions within account
func (c *ElbList) fetchElb(channel chan Elb) {
	defer elb_wg.Done()
	for _, region := range regions {
		elb_wg.Add(1)
		next_token := ""
		go c.fetchRegionElb(region, next_token, channel)
	}
}

// Print and send to channel list of instances.
func (c *ElbList) fetchRegionElb(region, next_token string, channel chan Elb) {
	defer elb_wg.Done()

	// Connect to region
	config := aws.Config{
		Region:      aws.String(region),
		Credentials: c.Profile.Credentials,
		MaxRetries:  aws.Int(20),
	}
	con := elb.New(session.New(), &config)

	// Prepare request
	params := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{},
		// Maximum count instances on one result page
		PageSize: aws.Int64(400),
	}

	if next_token != "" {
		params.Marker = aws.String(next_token)
	}

	// Get list of ec2 instances
	res, err := con.DescribeLoadBalancers(params)
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
	for _, r := range res.LoadBalancerDescriptions {
		channel <- Elb{Elb: *r, Profile: *c.Profile}
	}

	// If there are more instances repeat request with a token
	if res.NextMarker != nil {
		elb_wg.Add(1)
		go c.fetchRegionElb(region, *res.NextMarker, channel)
	}
}

// Returns all instances from all regions and accounts
func fetchElb() []Elb {
	var profile *Profile
	var elb []Elb

	ch_elb := make(chan Elb)
	defer close(ch_elb)

	// Run go routines to print instances
	for _, profile_name := range profiles {
		// If we didn't load regions already, then fill regions slice
		elb_wg.Add(1)
		profile = NewProfile(profile_name)
		go NewElbList(profile).fetchElb(ch_elb)
	}

	// Retreive results from all goroutines over channel
	go func() {
		for i := range ch_elb {
			elb = append(elb, i)
		}
	}()

	// Wait until receive info about all instances
	elb_wg.Wait()

	return elb
}

// Returns formatted string with elb information.
func formatElbOutput(profile string, i elb.LoadBalancerDescription) string {

	//strings.Join(i.AvailabilityZones, ","),
	//strings.Join(&i.Subnets[:], ","),
	//i.ListenerDescriptions,
	//i.Subnets
	e := []*string{
		i.LoadBalancerName,
		i.Scheme,
		i.DNSName,
		i.HealthCheck.Target,
		&profile,
		i.VPCId,
	}

	return makeFormattedOutput(e)
}
