package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"

	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type crtShRequest struct {
	Domain string `json:"domain"`
}

type crtShResponse struct {
	Subdomains []string `json:"subdomains"`
}

type FunctionError struct {
	ErrorMessage string `json:"errorMessage"`
	ErrorType    string `json:"errorType"`
}

func main() {
	awsRegion := flag.String("region", "", "optional: set aws region")
	flag.Parse()

	// Create Lambda service client
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	var client *lambda.Lambda

	if *awsRegion != "" {
		client = lambda.New(sess, &aws.Config{Region: aws.String(*awsRegion)})
	} else {
		client = lambda.New(sess)
	}

	var domain string

	if domain = flag.Arg(0); domain == "" {
		fmt.Println("Missing domain argument")
		os.Exit(1)
	}

	request := crtShRequest{domain}

	payload, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error marshalling CrtShFunction request")
		os.Exit(1)
	}

	result, err := client.Invoke(&lambda.InvokeInput{FunctionName: aws.String("CrtShFunction"), Payload: payload})
	if err != nil {
		fmt.Println("Error calling CrtShFunction")
		os.Exit(1)
	}

	var resp crtShResponse

	err = json.Unmarshal(result.Payload, &resp)
	if err != nil {
		fmt.Println("Error unmarshalling CrtShFunction response")
		os.Exit(1)
	}

	var functionError FunctionError

	// If the status code is NOT 200, the call failed
	if result.FunctionError != nil {
		err = json.Unmarshal(result.Payload, &functionError)
		if err != nil {
			fmt.Println("Error unmarshalling FunctionError")
			os.Exit(1)
		}

		fmt.Println(functionError.ErrorMessage)
		os.Exit(1)
	}

	// Print out subdomains
	if len(resp.Subdomains) > 0 {
		for i := range resp.Subdomains {
			fmt.Println(resp.Subdomains[i])
		}
	} else {
		fmt.Println("There were no subdomains")
	}
}
