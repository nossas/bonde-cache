package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/jasonlvhit/gocron"
	"github.com/kelseyhightower/envconfig"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

// Specification are enviroment variables
type Specification struct {
	Env             string
	Sync            bool
	Interval        float64
	Port            string
	PortSsl         string
	Domain          string
	RedisURL        string
	APIServiceURL   string
	APIServiceToken string
}

var s Specification
var (
	client *graphql.Client
)

func main() {

	err := envconfig.Process("cache", &s)
	if err != nil {
		log.Fatal(err.Error())
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s.APIServiceToken},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client = graphql.NewClient(s.APIServiceURL, httpClient)

	pool = RedisPool(s)

	go WebRedirect(s)
	go WebCache(s)

	Worker(s)
	<-gocron.Start()
}

func printJSON(v interface{}) {
	w := json.NewEncoder(os.Stdout)
	w.SetIndent("", "\t")
	err := w.Encode(v)
	if err != nil {
		panic(err)
	}
}
