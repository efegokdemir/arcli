package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// FeatureDisabledError describes an explicitly disabled feature. A missing route
// alone does not establish the server configuration.
type FeatureDisabledError struct {
	Feature   string // human name, e.g. "retention policies"
	ConfigKey string // e.g. "retention.enabled"
}

func (e *FeatureDisabledError) Error() string {
	return fmt.Sprintf("%s are disabled on this server (%s=false)", e.Feature, e.ConfigKey)
}

// redactedEndpoint hides any userinfo an endpoint URL might carry
// before it is echoed in an error.
func redactedEndpoint(endpoint string) string {
	if u, err := url.Parse(endpoint); err == nil {
		return u.Redacted()
	}
	return endpoint
}

// classifyFeatureRouteMissing preserves the original HTTP error. A health
// check can help diagnose a wrong endpoint, but cannot establish why a route
// is missing (configuration, licensing, server version, or otherwise).
func (c *Client) classifyFeatureRouteMissing(ctx context.Context, err error, feature, configKey string) error {
	if !isRouteMissing(err) {
		return err
	}
	if _, herr := c.Health(ctx); herr != nil {
		return fmt.Errorf("%w (and %s does not answer like an Arc server: %v; check the endpoint path)", err, redactedEndpoint(c.cfg.Endpoint), herr)
	}
	return err
}

// featureJSON is the shared request helper for feature-gated JSON
// routes: optional JSON body, bounded read, non-2xx → inspectable *HTTPError,
// 2xx → body returned and,
// when out is non-nil, decoded into it. A 2xx body of "null" is
// returned as-is so callers can normalise it.
func (c *Client) featureJSON(ctx context.Context, feature, configKey, method, path string, reqBody any, limit int64, out any) ([]byte, error) {
	var body io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.cfg.Endpoint+path, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c.setCrossDBHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, limit))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, c.classifyFeatureRouteMissing(ctx, decodeWriteError(resp.StatusCode, respBody), feature, configKey)
	}
	if out != nil {
		if err := json.Unmarshal(respBody, out); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
	}
	return respBody, nil
}
