package main

import (
	"errors"
	"math/rand"
	"time"

	"github.com/OWASP/Amass/v3/config"
	"github.com/OWASP/Amass/v3/datasrcs"
	"github.com/OWASP/Amass/v3/enum"
	"github.com/OWASP/Amass/v3/systems"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	// ErrNoDomain no domain in event payload
	ErrNoDomain = errors.New("No domain in event payload")
	// ErrFailedEnum failed to initialize new enumeration
	ErrFailedEnum = errors.New("Failed to initialize enumeration")
)

type amassEnumRequest struct {
	Domain string `json:"domain"`
}

type amassEnumResponse struct {
	Subdomains []string `json:"subdomains"`
}

func handler(request amassEnumRequest) (amassEnumResponse, error) {
	if request.Domain == "" {
		return amassEnumResponse{}, ErrNoDomain
	}

	// Seed the default pseudo-random number generator
	rand.Seed(time.Now().UTC().UnixNano())

	// Setup the most basic amass configuration
	cfg := config.NewConfig()
	cfg.AddDomain(request.Domain)
	cfg.Active = true
	cfg.Timeout = 14

	sys, err := systems.NewLocalSystem(cfg)
	if err != nil {
		return amassEnumResponse{}, ErrFailedEnum
	}
	sys.SetDataSources(datasrcs.GetAllSources(sys))

	e := enum.NewEnumeration(cfg, sys)
	if e == nil {
		return amassEnumResponse{}, err
	}
	defer e.Close()

	e.Start()

	var subdomains []string
	for _, o := range e.ExtractOutput(nil) {
		subdomains = append(subdomains, o.Name)
	}

	return amassEnumResponse{subdomains}, nil
}

func main() {
	lambda.Start(handler)
}
