package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	accept          = "Accept"
	authorization   = "Authorization"
	contentType     = "Content-Type"
	github          = "github.com"
	jsonContentType = "application/json; charset=utf-8"
	localhost       = "github.localhost"
	modulePath      = "github.com/cli/go-gh"
	timeZone        = "Time-Zone"
	userAgent       = "User-Agent"
)

func NewGHClient(host string) RestClient {
	return RestClient{
		client: *http.DefaultClient,
		host:   host,
	}
}

func restURL(hostname string, pathOrURL string) string {
	if strings.HasPrefix(pathOrURL, "https://") || strings.HasPrefix(pathOrURL, "http://") {
		return pathOrURL
	}
	return restPrefix(hostname) + pathOrURL
}

func restPrefix(hostname string) string {
	if isGarage(hostname) {
		return fmt.Sprintf("https://%s/api/v3/", hostname)
	}
	hostname = normalizeHostname(hostname)
	if isEnterprise(hostname) {
		return fmt.Sprintf("https://%s/api/v3/", hostname)
	}
	if strings.EqualFold(hostname, localhost) {
		return fmt.Sprintf("http://api.%s/", hostname)
	}
	return fmt.Sprintf("https://api.%s/", hostname)
}

type RestClient struct {
	client http.Client
	host   string
}

func (c RestClient) Get(path string, resp interface{}) error {
	return c.Do(http.MethodGet, path, nil, resp)
}

func (c RestClient) Do(method string, path string, body io.Reader, response interface{}) error {
	return c.DoWithContext(context.Background(), method, path, body, response)
}

func (c RestClient) DoWithContext(ctx context.Context, method string, path string, body io.Reader, response interface{}) error {
	url := restURL(c.host, path)
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !success {
		defer resp.Body.Close()
		return fmt.Errorf("error: %s", resp.Status)
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &response)
	if err != nil {
		return err
	}

	return nil
}

func isGarage(host string) bool {
	return strings.EqualFold(host, "garage.github.com")
}

func isEnterprise(host string) bool {
	return host != github && host != localhost
}
