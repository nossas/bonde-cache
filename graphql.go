package main

import (
	"context"
	"fmt"

	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

// Graphql specification to configure endpoint
type Graphql struct {
	s                     Specification
	client                *graphql.Client
	queryAllMobilizations struct {
		AllMobilizations struct {
			Edges []struct {
				Node   Mobilization
				Cursor graphql.String
			}
		} `graphql:"allMobilizations"`
	}
	queryAllDNSHostedZones struct {
		AllDnsHostedZones struct {
			Edges []struct {
				Node   DnsHostedZone
				Cursor graphql.String
			}
		} `graphql:"allDnsHostedZones"`
	}
	queryAllCertificates struct {
		AllCertificates struct {
			Edges []struct {
				Node   Certificate
				Cursor graphql.String
			}
		} `graphql:"allCertificates"`
	}
}

// CreateClient to query api micro services
func (g *Graphql) CreateClient() *Graphql {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.s.APIServiceToken},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	g.client = graphql.NewClient(g.s.APIServiceURL, httpClient)
	return g
}

// GetAllMobilizations via api micro services
func (g *Graphql) GetAllMobilizations() *Graphql {
	err2 := g.client.Query(context.Background(), &g.queryAllMobilizations, nil)
	if err2 != nil {
		fmt.Println("[worker]Error querying api services: ", err2)
	}
	// printJSON(query)
	return g
}

// GetAllDNSHostedZones via api micro services
func (g *Graphql) GetAllDNSHostedZones() *Graphql {
	err := g.client.Query(context.Background(), &g.queryAllDNSHostedZones, nil)
	if err != nil {
		fmt.Println("[worker]Error querying api services: ", err)
	}
	// printJSON(query3)
	return g
}

// GetAllCertificates via api micro services
func (g *Graphql) GetAllCertificates() *Graphql {
	err := g.client.Query(context.Background(), &g.queryAllCertificates, nil)
	if err != nil {
		fmt.Println("[worker]Error querying api services: ", err)
	}
	// printJSON(query3)
	return g
}
