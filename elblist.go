package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
	"log"
	"strings"
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

// Returns a pointer to a new ElbList object
func NewElbList(profile *Profile) *ElbList {
	return &ElbList{Profile: profile}
}

// Print instances from all regions within account
func (c *ElbList) fetchElb(channel chan Elb) {
	defer elb_wg.Done()
	var next_token string
	for _, region := range regions {
		elb_wg.Add(1)
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

	// Get list of elb
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

func formatListenersOutput(i elb.LoadBalancerDescription) string {
	var listener_str string
	var lb_port, i_port int64
	var lb_protocol, i_protocol string

	listeners := []string{}

	for _, l := range i.ListenerDescriptions {
		lb_port = *l.Listener.LoadBalancerPort
		i_port = *l.Listener.InstancePort
		lb_protocol = *l.Listener.Protocol
		i_protocol = *l.Listener.InstanceProtocol

		listener_str = fmt.Sprintf("%d:%s<=>:%d:%s",
			lb_port,
			lb_protocol,
			i_port,
			i_protocol)

		listeners = append(listeners, listener_str)
	}

	res := strings.Join(listeners, " ")
	return res
}

// Returns formatted string with elb information.
func formatElbOutput(profile string, i elb.LoadBalancerDescription) string {
	azs := formatSliceOutput(i.AvailabilityZones)
	subnets := formatSliceOutput(i.Subnets)
	listeners := formatListenersOutput(i)

	e := []*string{
		i.LoadBalancerName,
		i.Scheme,
		i.DNSName,
		&listeners,
		i.HealthCheck.Target,
		&azs,
		&subnets,
		i.VPCId,
		&profile,
	}

	return makeFormattedOutput(e)
}

// TODO: formatElbOutput show instances
// TODO: formatProfileElbOutput limit to profile
// TODO: formatAllElbOutput print all
