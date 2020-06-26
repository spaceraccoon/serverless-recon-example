// Package main extracts subdomains from the sublist3r API.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

// handler connects to the sublist3r API, queries for subdomains, and returns the parsed results.
func handler(request lambdaRequest) (lambdaResponse, error) {
	if request.Domain == "" {
		return lambdaResponse{}, errors.New("No domain in event payload")
	}

	resp, err := http.Get(fmt.Sprintf("https://api.sublist3r.com/search.php?domain=%s", request.Domain))
	if err != nil {
		return lambdaResponse{}, err
	}

	var subdomains []string

	err = json.NewDecoder(resp.Body).Decode(&subdomains)
	if err != nil {
		return lambdaResponse{}, err
	}

	return lambdaResponse{subdomains}, nil
}

func main() {
	lambda.Start(handler)
}
