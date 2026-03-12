package main

import "testing"

func TestParseTestConfig(t *testing.T) {
	tests := []struct {
		in     string
		expect ConnectionTest
	}{
		{
			in: "api=http://localhost:8080/health",
			expect: ConnectionTest{
				Service: "api",
				URL:     "http://localhost:8080/health",
			},
		},
		{
			in: "db=postgres://localhost:5432",
			expect: ConnectionTest{
				Service: "db",
				URL:     "postgres://localhost:5432",
			},
		},
		{
			// Missing '=' should result in empty Service and URL
			in: "",
			expect: ConnectionTest{
				Service: "",
				URL:     "",
			},
		},
		{
			// No value after '='
			in: "empty=",
			expect: ConnectionTest{
				Service: "empty",
				URL:     "",
			},
		},
	}

	for _, tt := range tests {
		got := parseTestConfig(tt.in)
		if got.Service != tt.expect.Service || got.URL != tt.expect.URL {
			t.Errorf("parseTestConfig(%q) = %+v, want %+v", tt.in, got, tt.expect)
		}
	}
}