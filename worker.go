// Worker has two main tasks:
// - populate redis with cache html content from mobilizations published
// - validate certificates with dns challenge
package main

import (
	"github.com/jasonlvhit/gocron"
)

// Worker are entry point to recurring tasks
func Worker(s Specification) {
	populateCache(s)
	populateCertificates(s)
	gocron.Every(30).Seconds().Do(populateCache, s)
	gocron.Every(30).Seconds().Do(populateCertificates, s)
}
