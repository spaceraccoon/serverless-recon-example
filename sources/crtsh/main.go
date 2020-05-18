package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/lib/pq"
)

const (
	host   = "crt.sh"
	port   = 5432
	user   = "guest"
	dbname = "certwatch"
)

var (
	// ErrNoDomain no domain in event payload
	ErrNoDomain = errors.New("No domain in event payload")
)

type crtShRequest struct {
	Domain string `json:"domain"`
}

type crtShResponse struct {
	Subdomains []string `json:"subdomains"`
}

func handler(request crtShRequest) (crtShResponse, error) {
	if request.Domain == "" {
		return crtShResponse{}, ErrNoDomain
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, port, user, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return crtShResponse{}, err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return crtShResponse{}, err
	}

	rows, err := db.Query("SELECT DISTINCT NAME_VALUE FROM certificate_identity ci WHERE ci.NAME_TYPE = 'dNSName' AND reverse(lower(ci.NAME_VALUE)) LIKE reverse(lower('%.' || $1))", request.Domain)
	if err != nil {
		return crtShResponse{}, err
	}
	defer rows.Close()

	var subdomains []string

	for rows.Next() {
		var subdomain string
		err = rows.Scan(&subdomain)
		if err != nil {
			return crtShResponse{}, err
		}

		subdomains = append(subdomains, subdomain)
	}

	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		return crtShResponse{}, err
	}

	return crtShResponse{subdomains}, nil
}

func main() {
	lambda.Start(handler)
}
