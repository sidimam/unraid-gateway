package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Query runs a GraphQL operation server-side with the given API key and returns the "data" part.
// GraphQL-level errors come back as an error with the messages joined.
func (g *GraphQL) Query(ctx context.Context, apiKey, query string, variables map[string]any) (json.RawMessage, error) {
	payload, _ := json.Marshal(map[string]any{"query": query, "variables": variables})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.target, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", apiKey)
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unraid unreachable: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	var out struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("unraid answered HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body[:min(len(body), 200)])))
	}
	if len(out.Errors) > 0 {
		msgs := make([]string, 0, len(out.Errors))
		for _, e := range out.Errors {
			msgs = append(msgs, e.Message)
		}
		return out.Data, errors.New(strings.Join(msgs, "; "))
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("unraid answered HTTP %d", resp.StatusCode)
	}
	return out.Data, nil
}
