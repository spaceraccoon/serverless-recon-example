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

type functionRequest struct {
	Domain string `json:"domain"`
}

type functionResponse struct {
	Subdomains []string `json:"subdomains"`
}

type FunctionError struct {
	ErrorMessage string `json:"errorMessage"`
	ErrorType    string `json:"errorType"`
}

func invokeFunction(domain string, subdomains map[string]bool, client *lambda.Lambda, functionName string, payload []byte) map[string]bool {
	result, err := client.Invoke(&lambda.InvokeInput{FunctionName: aws.String(functionName), Payload: payload})
	if err != nil {
		fmt.Printf("Error calling %s\n", functionName)
		os.Exit(1)
	}

	var resp functionResponse

	err = json.Unmarshal(result.Payload, &resp)
	if err != nil {
		fmt.Printf("Error unmarshalling %s response\n", functionName)
		os.Exit(1)
	}

	var functionError FunctionError

	// If result.FunctionError exists, the call failed
	if result.FunctionError != nil {
		err = json.Unmarshal(result.Payload, &functionError)
		if err != nil {
			fmt.Printf("Error unmarshalling %s error\n", functionName)
			os.Exit(1)
		}

		fmt.Println(functionError.ErrorMessage)
		os.Exit(1)
	}

	for _, subdomain := range resp.Subdomains {
		subdomains[subdomain] = true
	}

	return subdomains
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

	request := functionRequest{domain}
	uniqueSubdomains := map[string]bool{}

	payload, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error marshalling function request")
		os.Exit(1)
	}

	// uniqueSubdomains = invokeFunction(domain, uniqueSubdomains, client, "CrtShFunction", payload)
	uniqueSubdomains = invokeFunction(domain, uniqueSubdomains, client, "AmassEnumFunction", payload)

	for subdomain := range uniqueSubdomains {
		fmt.Println(subdomain)
	}
}
