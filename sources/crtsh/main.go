// Package main extracts subdomains from the crt.sh Postgres database.
package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/lib/pq"
)

// crt.sh database details.
const (
	host   = "crt.sh"
	port   = 5432
	user   = "guest"
	dbname = "certwatch"
)

// lambdaRequest contains the Lambda function arguments.
type lambdaRequest struct {
	Domain string `json:"domain"`
}

// lambdaResponse contains the Lambda function response.
type lambdaResponse struct {
	Subdomains []string `json:"subdomains"`
}

// handler connects to the crt.sh database, queries for subdomains, and returns the parsed results.
func handler(request lambdaRequest) (lambdaResponse, error) {
	if request.Domain == "" {
		return lambdaResponse{}, errors.New("No domain in event payload")
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, port, user, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return lambdaResponse{}, err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return lambdaResponse{}, err
	}

	// Make Postgres SQL query.
	rows, err := db.Query("SELECT DISTINCT NAME_VALUE FROM certificate_identity ci WHERE ci.NAME_TYPE = 'dNSName' AND reverse(lower(ci.NAME_VALUE)) LIKE reverse(lower('%.' || $1))", request.Domain)
	if err != nil {
		return lambdaResponse{}, err
	}
	defer rows.Close()

	var subdomains []string

	for rows.Next() {
		var subdomain string
		err = rows.Scan(&subdomain)
		if err != nil {
			return lambdaResponse{}, err
		}

		subdomains = append(subdomains, subdomain)
	}

	// Get any error encountered during iteration.
	err = rows.Err()
	if err != nil {
		return lambdaResponse{}, err
	}

	return lambdaResponse{subdomains}, nil
}

func main() {
	lambda.Start(handler)
}
