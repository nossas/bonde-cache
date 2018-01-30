package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/joeguo/tldextract"
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

// Certs handle async files to use at web server
type CertManager struct {
	s Specification
	r *Redis
	g *Graphql
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
func (certManager *CertManager) Populate(s Specification) {
	log.Println("[populate_cert_manager] job started")

	cache := "/tmp/tld.cache"
	extract, _ := tldextract.New(cache, false)
	var domainsAvailableCertificate []string

	var q = certManager.g.GetAllDNSHostedZones()
	for _, mob := range q.queryAllDNSHostedZones.AllDnsHostedZones.Edges {
		rootDomain := extract.Extract(mob.Node.DomainName)
		var d = rootDomain.Root + "." + rootDomain.Tld
		domainsAvailableCertificate = append(domainsAvailableCertificate, d)
	}

	q = certManager.g.GetAllCertificates()
	for _, mob := range q.queryAllCertificates.AllCertificates.Edges {
		rootDomain := extract.Extract(mob.Node.Domain)
		var d = rootDomain.Root + "." + rootDomain.Tld
		domainsAvailableCertificate = append(domainsAvailableCertificate, d)
		// write certificate file content do disk
	}

	q = certManager.g.GetAllMobilizations()
	for _, mob := range q.queryAllMobilizations.AllMobilizations.Edges {
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
