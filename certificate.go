package main

import (
	"fmt"
	"log"

	"github.com/joeguo/tldextract"
)

// func generateNewCertificates() []byte {
// 	content, err := ioutil.ReadFile("testdata/hello")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Printf("File contents: %s", content)

// 	return content

// 	// tentar criar com desafio dns do lets encrypt, salva na api-microservice

// }

// ler da api-microservice, salva no redis e cria como arquivo
// checa certificados encontrados no redis existem como arquivo
//
// queries:
// - list communities with one or more mobilizations(custom_domain not empty)
// - list mobilizations with custom_domain not empty
// - list domains with activate field equal true
// - create certificates (community_id, mobilization_id, domain_id, key, content)

// Load Certificate or generate a one if new domain created
func populateCertificates(s Specification) {
	log.Println("[populateCertificates] job started")
	_, mobs := GetUrls(s)

	cache := "/tmp/tld.cache"
	extract, _ := tldextract.New(cache, false)

	for _, mob := range mobs {
		rootDomain := extract.Extract(mob.CustomDomain)
		fmt.Printf("%s\n", rootDomain.Root)

		var cachedCert = RedisReadMobilization("cached_certificates:" + rootDomain.Root)
		if string(cachedCert.Name) == "" {
			log.Println("[populateCertificate] NEW CERT FOUND")
			// generateNewCertificates()
		} else {

		}
	}
	// time.Sleep(30 * time.Second)
	// pid := os.Getpid()
	// proc, _ := os.FindProcess(pid)
	// proc.Signal(os.Interrupt)
}
