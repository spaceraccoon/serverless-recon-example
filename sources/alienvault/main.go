// Package main extracts subdomains from the AlienVault API.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
)

// lambdaRequest contains the Lambda function arguments.
type lambdaRequest struct {
	Domain string `json:"domain"`
}

// lambdaResponse contains the Lambda function response.
type lambdaResponse struct {
	Subdomains []string `json:"subdomains"`
}

// alienVaultResponse contains the JSON representation of AlienVault's API response.
type alienVaultResponse struct {
	PassiveDNS []struct {
		Hostname string `json:"hostname"`
	} `json:"passive_dns"`
}

// isIpv4Address checks if host is an IPv4 address
func isIpv4Address(host string) bool {
	return net.ParseIP(host) != nil
}

// handler connects to the AlienVault API, queries for subdomains, and returns the parsed results.
func handler(request lambdaRequest) (lambdaResponse, error) {
	if request.Domain == "" {
		return lambdaResponse{}, errors.New("No domain in event payload")
	}

	resp, err := http.Get(fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", request.Domain))
	if err != nil {
		return lambdaResponse{}, err
	}

	if resp.StatusCode != 200 {
		return lambdaResponse{}, fmt.Errorf("Invalid status code: %d", resp.StatusCode)
	}

	var subdomains []string
	var response alienVaultResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return lambdaResponse{}, err
	}

	for _, record := range response.PassiveDNS {
		// Discard IPv4 address hostnames.
		if !isIpv4Address(record.Hostname) {
			subdomains = append(subdomains, record.Hostname)
		}
	}

	return lambdaResponse{subdomains}, nil
}

func main() {
	lambda.Start(handler)
}
