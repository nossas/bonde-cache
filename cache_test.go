package main

import "testing"

func TestCacheManager_Populate(t *testing.T) {
	type fields struct {
		s Specification
		g *Graphql
		r *Redis
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
			tt.fields.s.APIServiceURL = "http://local:3002"
			tt.fields.s.APIServiceToken = "aaa"

			// c := &CacheManager{
			// 	s: tt.fields.s,
			// 	g: tt.fields.g,
			// 	r: tt.fields.r,
			// }
			// c.Populate()
		})
	}
}
