package main

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/vaughan0/go-ini"
	"os"
	"path/filepath"
)

var (
	// @readonly
	LoadProfilesList    = awserr.New("LoadProfilesList", "failed to load profiles list from credentials file", nil)
	CredentialsFileName = "credentials"
)

type Profile struct {
	Name        string
	Credentials *credentials.Credentials
}

func NewProfile(profile string) *Profile {
	// Load aws credentials from ~/.aws/config file
	filename := filepath.Join(os.Getenv("HOME"), ".aws", CredentialsFileName)
	creds := credentials.NewSharedCredentials(filename, profile)

	return &Profile{
		Name:        profile,
		Credentials: creds,
	}
}

// Returns list of profiles names from credentials file
func fetchProfiles() ([]string, error) {
	filename := filepath.Join(os.Getenv("HOME"), ".aws", CredentialsFileName)

	// Parse credentials file
	f, err := ini.LoadFile(filename)
	if err != nil {
		return []string{}, LoadProfilesList
	}

	// Fill profiles slice with list of profiles
	profiles := []string{}
	for profile := range f {
		profiles = append(profiles, profile)
	}

	return profiles, nil
}
