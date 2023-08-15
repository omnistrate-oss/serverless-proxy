package sidecar

import (
	"context"
	"log"
	"net/http"
	"time"
)

type Client struct {
	context context.Context
	*http.Client
}

func NewClient(ctx context.Context) *Client {
	return &Client{ctx, &http.Client{Timeout: 60 * time.Second, Transport: http.DefaultTransport}}
}

func (c *Client) SendAPIRequest() (*http.Response, error) {

	resp, err := http.Get("http://127.0.0.1:49750/getBackendsStatus")
	if err != nil {
		log.Printf("Failed to get backends endpoints")
	}
	log.Printf("Response: %v", resp)

	return resp, err
}
