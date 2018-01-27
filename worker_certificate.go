package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/joeguo/tldextract"
	"github.com/shurcooL/graphql"
)

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

// Certificate is cached at Redis
type Certificate struct {
	ID              int    `json:"id" redis:"id" graphql:"id"`
	CommunityID     int    `json:"community_id" redis:"community_id" graphql:"communityId"`
	MobilizationID  int    `json:"mobilization_id" redis:"mobilization_id" graphql:"mobilizationId"`
	DnsHostedZoneID int    `json:"dns_hosted_zone_id" redis:"dns_hosted_zone_id" graphql:"dnsHostedZoneId"`
	Domain          string `json:"domain" redis:"domain" graphql:"domain"`
	FileContent     []byte `json:"file_content" redis:"file_content" graphql:"fileContent"`
	ExpireOn        string `json:"expire_on" redis:"expire_on" graphql:"expireOn"`
	IsActive        bool   `json:"is_active" redis:"is_active" graphql:"is_active"`
	CreatedAt       string `json:"created_at" redis:"created_at" graphql:"createdAt"`
	UpdatedAt       string `json:"updated_at" redis:"updated_at" graphql:"updatedAt"`
}

// func generateCertificates() []byte {
// 	log.Println("[generateCertificates] job started")
// 	return nil
// }

// // Restore Certificates from s3
// func restoreCertificates() {}

// // Import certificates from dir
// func importCertificates() {}

func removeDuplicates(elements []string) []string {
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

// Load Certificate or generate a one if new domain created
func populateCertificates(s Specification) {
	log.Println("[populateCertificates] job started")

	var query struct {
		AllMobilizations struct {
			Edges []struct {
				Node   Mobilization
				Cursor graphql.String
			}
		} `graphql:"allMobilizations"`
	}
	err2 := client.Query(context.Background(), &query, nil)
	if err2 != nil {
		fmt.Println("Error querying api services: ", err2)
	}

	var query2 struct {
		AllDnsHostedZones struct {
			Edges []struct {
				Node   DnsHostedZone
				Cursor graphql.String
			}
		} `graphql:"allDnsHostedZones"`
	}
	err3 := client.Query(context.Background(), &query2, nil)
	if err3 != nil {
		fmt.Println("Error querying api services: ", err3)
	}

	var query3 struct {
		AllCertificates struct {
			Edges []struct {
				Node   Certificate
				Cursor graphql.String
			}
		} `graphql:"allCertificates"`
	}
	err4 := client.Query(context.Background(), &query3, nil)
	if err2 != nil {
		fmt.Println("Error querying api services: ", err4)
	}
	// printJSON(query3)

	cache := "/tmp/tld.cache"
	extract, _ := tldextract.New(cache, false)

	var domainsAvailableCertificate []string
	for _, mob := range query2.AllDnsHostedZones.Edges {
		rootDomain := extract.Extract(mob.Node.DomainName)
		var d = rootDomain.Root + "." + rootDomain.Tld
		domainsAvailableCertificate = append(domainsAvailableCertificate, d)
	}

	for _, mob := range query3.AllCertificates.Edges {
		rootDomain := extract.Extract(mob.Node.Domain)
		var d = rootDomain.Root + "." + rootDomain.Tld
		domainsAvailableCertificate = append(domainsAvailableCertificate, d)
		// write certificate file content do disk
	}

	for _, mob := range query.AllMobilizations.Edges {
		rootDomain := extract.Extract(mob.Node.CustomDomain)
		var d = rootDomain.Root + "." + rootDomain.Tld
		for _, v := range domainsAvailableCertificate {
			if v == d {
				_, err := ioutil.ReadFile("./data/certificates/certificates/" + rootDomain.Sub + "." + d)
				if err != nil {
					fmt.Printf("Arquivo de certificado não encontrado, gerando um: %s\n", err)
					// - gera novo certificado com lib lego
					// - salvar na api-microservices o certificado encontrado
				}
			}
		}
	}
	// time.Sleep(30 * time.Second)
	// pid := os.Getpid()
	// proc, _ := os.FindProcess(pid)
	// proc.Signal(os.Interrupt)
}
