// Package main extracts subdomains from the Censys API.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// lambdaRequest contains the Lambda function arguments.
type lambdaRequest struct {
	Domain string `json:"domain"`
}

// lambdaResponse contains the Lambda function response.
type lambdaResponse struct {
	Subdomains []string `json:"subdomains"`
}

// lambdaResponse contains the Lambda function response.
type censysAPIKey struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

type censysResponseResults struct {
	Data  []string `json:"parsed.extensions.subject_alt_name.dns_names"`
	Data1 []string `json:"parsed.names"`
}

type censysResponse struct {
	Results  []censysResponseResults `json:"results"`
	Metadata struct {
		Pages int `json:"pages"`
	} `json:"metadata"`
}

func getAPIKey() (censysAPIKey, error) {
	var apiKey censysAPIKey

	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	ssmsvc := ssm.New(sess, aws.NewConfig())
	keyName := os.Getenv("API_KEY_PATH")
	withDecryption := false
	getParameterInput := ssm.GetParameterInput{
		Name:           &keyName,
		WithDecryption: &withDecryption,
	}
	getParameterOutput, err := ssmsvc.GetParameter(&getParameterInput)
	if err != nil {
		return apiKey, err
	}

	apiKeyString := *getParameterOutput.Parameter.Value

	err = json.Unmarshal([]byte(apiKeyString), &apiKey)
	if err != nil {
		return apiKey, err
	}

	return apiKey, nil
}

// handler connects to the AlienVault API, queries for subdomains, and returns the parsed results.
func handler(request lambdaRequest) (lambdaResponse, error) {
	if request.Domain == "" {
		return lambdaResponse{}, errors.New("No domain in event payload")
	}

	apiKey, err := getAPIKey()
	if err != nil {
		return lambdaResponse{}, err
	}

	if apiKey.ID == "" || apiKey.Secret == "" {
		return lambdaResponse{}, errors.New("No API key provided")
	}

	var subdomains []string
	var response censysResponse

	currentPage := 1
	for {
		client := &http.Client{}
		var request = []byte(`{"query":"` + request.Domain + `", "page":` + strconv.Itoa(currentPage) + `, "fields":["parsed.names","parsed.extensions.subject_alt_name.dns_names"], "flatten":true}`)

		req, err := http.NewRequest("POST", "https://www.censys.io/api/v1/search/certificates", bytes.NewReader(request))
		if err != nil {
			return lambdaResponse{subdomains}, err
		}
		req.SetBasicAuth(apiKey.ID, apiKey.Secret)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return lambdaResponse{subdomains}, err
		}

		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			resp.Body.Close()
			return lambdaResponse{subdomains}, err
		}
		resp.Body.Close()

		// Exit the censys enumeration if max pages is reached
		if currentPage >= response.Metadata.Pages {
			break
		}

		for _, res := range response.Results {
			for _, part := range res.Data {
				subdomains = append(subdomains, part)
			}
			for _, part := range res.Data1 {
				subdomains = append(subdomains, part)
			}
		}
		currentPage++
	}

	return lambdaResponse{subdomains}, nil
}

func main() {
	lambda.Start(handler)
}
