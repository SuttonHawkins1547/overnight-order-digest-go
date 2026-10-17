package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPostReportsServerStatusBeforeEnvelopeRejection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"ok":false,"error":{"message":"unavailable"}}`))
	}))
	defer server.Close()

	c := client{baseURL: server.URL, http: server.Client()}
	_, err := c.post("/v1/pdf/generate", map[string]any{"html": "<p>test</p>"})
	if err == nil || !strings.Contains(err.Error(), "transport status 500") {
		t.Fatalf("post() error = %v, want transport status 500", err)
	}
}

func TestShouldShipDigest(t *testing.T) {
	tests := []struct {
		name   string
		orders []Order
		want   bool
	}{
		{"completed backlog", []Order{{Status: "fulfilled"}}, true},
		{"empty backlog", nil, false},
		{"unclassified records", []Order{{ID: "x"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldShipDigest(tt.orders); got != tt.want {
				t.Fatalf("shouldShipDigest() = %v, want %v", got, tt.want)
			}
		})
	}
}
