package main

import (
	"io/ioutil"
	"log"

	"github.com/joeguo/tldextract"
)

// CertManager handle async files to use at web server
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

// Populate or generate a one if new domain created
func (certManager *CertManager) Populate() {
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
		// TODO: write certificate file content from db to disk
	}

	q = certManager.g.GetAllMobilizations()
	for _, mob := range q.queryAllMobilizations.AllMobilizations.Edges {
		rootDomain := extract.Extract(mob.Node.CustomDomain)
		var d = rootDomain.Root + "." + rootDomain.Tld
		for _, v := range domainsAvailableCertificate {
			if v == d {
				_, err := ioutil.ReadFile("./data/certificates/certificates/" + rootDomain.Sub + "." + d)
				if err != nil {
					log.Printf("Arquivo de certificado não encontrado, gerando um: %s\n", err)
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

// Import certificates from dir
func (certManager *CertManager) importCertificates() {

}
