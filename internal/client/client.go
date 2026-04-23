package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
)

type Client struct {
	APIKey   string
	Endpoint string
	HTTP     *retryablehttp.Client
	Verbose  bool
	Stderr   io.Writer
}

type Options struct {
	APIKey   string
	Endpoint string
	Timeout  time.Duration
	Retries  int
	Verbose  bool
	Stderr   io.Writer
}

func New(o Options) *Client {
	rc := retryablehttp.NewClient()
	if o.Retries >= 0 {
		rc.RetryMax = o.Retries
	} else {
		rc.RetryMax = 2
	}
	rc.Logger = nil
	if o.Timeout > 0 {
		rc.HTTPClient.Timeout = o.Timeout
	} else {
		rc.HTTPClient.Timeout = 120 * time.Second
	}
	rc.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if err != nil {
			return true, nil
		}
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			return true, nil
		}
		return false, nil
	}
	stderr := o.Stderr
	if stderr == nil {
		stderr = io.Discard
	}
	return &Client{
		APIKey:   o.APIKey,
		Endpoint: strings.TrimRight(o.Endpoint, "/"),
		HTTP:     rc,
		Verbose:  o.Verbose,
		Stderr:   stderr,
	}
}

type Response struct {
	Status      int
	ContentType string
	Body        []byte
}

// Do calls the given fully-qualified URL with the given HTTP method.
// For GET, query params are encoded from `params`. For POST, the body is JSON of `body`.
func (c *Client) Do(ctx context.Context, method, endpoint string, params url.Values, body any) (*Response, error) {
	if c.APIKey == "" {
		return nil, errors.New("no API key configured (set HASDATA_API_KEY, pass --api-key, or run `hasdata configure`)")
	}

	method = strings.ToUpper(method)
	var reqBody io.Reader
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("marshal body: %w", err)
			}
			reqBody = bytes.NewReader(b)
			if c.Verbose {
				fmt.Fprintf(c.Stderr, "> %s %s\n> body: %s\n", method, endpoint, string(b))
			}
		}
	}
	u := endpoint
	if method == http.MethodGet && len(params) > 0 {
		if strings.Contains(u, "?") {
			u += "&" + params.Encode()
		} else {
			u += "?" + params.Encode()
		}
	}
	if c.Verbose && method == http.MethodGet {
		fmt.Fprintf(c.Stderr, "> %s %s\n", method, u)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, method, u, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("x-request-source", "hasdata-cli")
	req.Header.Set("User-Agent", "hasdata-cli")
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		req.Header.Set("Idempotency-Key", uuid.NewString())
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if c.Verbose {
		for _, k := range []string{"X-RateLimit-Remaining", "X-RateLimit-Limit", "X-RateLimit-Reset", "Retry-After"} {
			if v := resp.Header.Get(k); v != "" {
				fmt.Fprintf(c.Stderr, "< %s: %s\n", k, v)
			}
		}
	}
	return &Response{
		Status:      resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		Body:        b,
	}, nil
}
