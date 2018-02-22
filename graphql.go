package main

import (
	"context"
	"log"

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

// DnsHostedZone represent valid root domain with dns verified
type DnsHostedZone struct {
	ID          int    `json:"id" redis:"id" graphql:"id"`
	CommunityID int    `json:"community_id" redis:"community_id" graphql:"communityId"`
	DomainName  string `json:"domain_name" redis:"domain_name" graphql:"domainName"`
	Comment     string `json:"comment" redis:"comment" graphql:"comment"`
	CreatedAt   string `json:"created_at" redis:"created_at" graphql:"createdAt"`
	UpdatedAt   string `json:"updated_at" redis:"updated_at" graphql:"updatedAt"`
	// Response        string `json:"response" redis:"response" graphql:"response"`
	// IsActive bool `json:"ns_ok" redis:"ns_ok" graphql:"nsOk"`
}

// Certificate from Graphql and cached at Redis
type Certificate struct {
	ID              int    `json:"id" redis:"id" graphql:"id"`
	CommunityID     int    `json:"community_id" redis:"community_id" graphql:"communityId"`
	MobilizationID  int    `json:"mobilization_id" redis:"mobilization_id" graphql:"mobilizationId"`
	DnsHostedZoneID int    `json:"dns_hosted_zone_id" redis:"dns_hosted_zone_id" graphql:"dnsHostedZoneId"`
	Domain          string `json:"domain" redis:"domain" graphql:"domain"`
	FileContent     []byte `json:"file_content" redis:"file_content" graphql:"fileContent"`
	ExpireOn        string `json:"expire_on" redis:"expire_on" graphql:"expireOn"`
	IsActive        bool   `json:"is_active" redis:"is_active" graphql:"isActive"`
	IsGenerated     bool   `json:"is_generated" redis:"is_generated" graphql:"isGenerated"`
	IsImported      bool   `json:"is_imported" redis:"is_imported" graphql:"isImported"`
	CreatedAt       string `json:"created_at" redis:"created_at" graphql:"createdAt"`
	UpdatedAt       string `json:"updated_at" redis:"updated_at" graphql:"updatedAt"`
}

// Mobilization from Graphql and cached at Redis
type Mobilization struct {
	ID              int    `json:"id" redis:"id" graphql:"id"`
	CommunityID     int    `json:"community_id" redis:"community_id" graphql:"communityId"`
	Name            string `json:"name" redis:"name" graphql:"name"`
	Content         []byte `json:"content" redis:"content" graphql:""`
	CachedAt        string `json:"cached_at" redis:"cached_at" graphql:""`
	Slug            string `json:"slug" redis:"slug" graphql:"slug"`
	CustomDomain    string `json:"custom_domain" redis:"custom_domain" graphql:"customDomain"`
	UpdatedAt       string `json:"updated_at" redis:"updated_at" graphql:"updatedAt"`
	Public          bool   `json:"public" redis:"public" graphql:""`
	CertificateRoot bool   `json:"certificate_root" redis:"certificate_root" graphql:""`
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
		log.Println("[worker]Error querying api services: ", err2)
	}
	// printJSON(query)
	return g
}

// GetAllDNSHostedZones via api micro services
func (g *Graphql) GetAllDNSHostedZones() *Graphql {
	err := g.client.Query(context.Background(), &g.queryAllDNSHostedZones, nil)
	if err != nil {
		log.Println("[worker]Error querying api services: ", err)
	}
	// printJSON(query3)
	return g
}

// GetAllCertificates via api micro services
func (g *Graphql) GetAllCertificates() *Graphql {
	err := g.client.Query(context.Background(), &g.queryAllCertificates, nil)
	if err != nil {
		log.Println("[worker]Error querying api services: ", err)
	}
	// printJSON(query3)
	return g
}
