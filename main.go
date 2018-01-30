package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/jasonlvhit/gocron"
	"github.com/kelseyhightower/envconfig"
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

func main() {
	var s Specification

	err := envconfig.Process("cache", &s)
	if err != nil {
		log.Fatal(err.Error())
	}

	g := &Graphql{s: s}
	r := &Redis{s: s}
	web := &Web{s: s, g: g, r: r}
	worker := &Worker{s: s, g: g, r: r}

	g.CreateClient()
	r.CreatePool()

	go web.StartNonSSL()
	go web.StartSSL()

	worker.Start()
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
