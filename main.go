package main

import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/lambda"
    
	"encoding/json"
	"flag"
    "fmt"
    "os"
    "strconv"
)

type crtShRequest struct {
    Domain string `json:"domain"`
}

type crtShResponseBody struct {
	Error		 string	  `json:"error"`
    Subdomains   []string `json:"subdomains"`
}

type crtShResponse struct {
    StatusCode int                  `json:"statusCode"`
    Body       crtShResponseBody    `json:"body"`
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
        os.Exit(0)
    }

    request := crtShRequest{domain}

    payload, err := json.Marshal(request)
    if err != nil {
        fmt.Println("Error marshalling CrtShFunction request")
        os.Exit(0)
    }

    result, err := client.Invoke(&lambda.InvokeInput{FunctionName: aws.String("CrtShFunction"), Payload: payload})
    if err != nil {
        fmt.Println("Error calling CrtShFunction")
        os.Exit(0)
    }

    var resp crtShResponse

    err = json.Unmarshal(result.Payload, &resp)
    if err != nil {
        fmt.Println("Error unmarshalling CrtShFunction response")
        os.Exit(0)
    }

    // If the status code is NOT 200, the call failed
    if resp.StatusCode != 200 {
        fmt.Println(strconv.Itoa(resp.StatusCode) + ": " + resp.Body.Error)
        os.Exit(0)
    }

    // Print out subdomains
    if len(resp.Body.Subdomains) > 0 {
        for i := range resp.Body.Subdomains {
            fmt.Println(resp.Body.Subdomains[i])
        }
    } else {
        fmt.Println("There were no subdomains")
    }
}