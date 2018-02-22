package main

import "testing"

func TestWorker_Setup(t *testing.T) {
	type fields struct {
		cache *CacheManager
		certs *CertManager
		s     Specification
		g     *Graphql
		r     *Redis
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Worker{
				cache: tt.fields.cache,
				certs: tt.fields.certs,
				s:     tt.fields.s,
				g:     tt.fields.g,
				r:     tt.fields.r,
			}
			w.Setup()
		})
	}
}
