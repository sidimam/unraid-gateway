// Package proxy forwards GraphQL requests to the Unraid API using the
// caller's API key, so the Unraid WebGUI never has to be exposed.
package proxy

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"time"
)

// GraphQL is a minimal single-endpoint reverse proxy.
type GraphQL struct {
	target string
	client *http.Client
}

// New creates a proxy towards <unraidURL>/graphql.
func New(unraidURL string, insecure bool) *GraphQL {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	if insecure {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // opt-in for self-signed Unraid certs
	}
	return &GraphQL{
		target: strings.TrimRight(unraidURL, "/") + "/graphql",
		client: &http.Client{Transport: tr, Timeout: 60 * time.Second},
	}
}

// Forward sends the request body to Unraid with the given API key and copies
// the response back verbatim.
func (g *GraphQL) Forward(w http.ResponseWriter, r *http.Request, apiKey string, maxBody int64) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
	if err != nil {
		http.Error(w, `{"error":"cannot read body"}`, http.StatusBadRequest)
		return
	}
	if int64(len(body)) > maxBody {
		http.Error(w, `{"error":"body too large"}`, http.StatusRequestEntityTooLarge)
		return
	}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, g.target, bytes.NewReader(body))
	if err != nil {
		http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", apiKey)
	resp, err := g.client.Do(req)
	if err != nil {
		http.Error(w, `{"error":"unraid unreachable"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	for _, h := range []string{"Content-Type", "Content-Length"} {
		if v := resp.Header.Get(h); v != "" {
			w.Header().Set(h, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
