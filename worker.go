// Worker has two main tasks:
// - populate redis with cache html content from mobilizations published
// - validate certificates with dns challenge
package main

import "github.com/jasonlvhit/gocron"

// Worker act as manager to cache and certs
type Worker struct {
	cache *CacheManager
	certs *CertManager
	s     Specification
	g     *Graphql
	r     *Redis
}

// Start are entry point to recurring tasks
func (w *Worker) Start() {
	w.cache = &CacheManager{
		g: w.g,
		s: w.s,
		r: w.r,
	}
	w.certs = &CertManager{
		g: w.g,
		s: w.s,
		r: w.r,
	}

	gocron.Every(30).Seconds().Do(w.cache.Populate)
	gocron.Every(30).Seconds().Do(w.certs.Populate)
}
