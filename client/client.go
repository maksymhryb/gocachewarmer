package client

import (
	"log"
	"net/http"
	"time"
)

type transport struct {
	http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "GoCacheWarmer")
	log.Printf("HTTP request [%s] %s\n", req.Method, req.URL)
	resp, err := t.RoundTripper.RoundTrip(req)
	log.Printf("HTTP response [%s] %s - %d\n", req.Method, req.URL, resp.StatusCode)

	return resp, err
}

var client *http.Client

func NewClient() *http.Client {
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &transport{
				RoundTripper: http.DefaultTransport,
			},
		}
	}

	return client
}
