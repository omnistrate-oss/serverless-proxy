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

func (c *Client) SendAPIRequest(port string) (*http.Response, error) {

	resp, err := http.Get("http://127.0.0.1:49750/instanceStatus/ports/" + port)
	if err != nil {
		log.Printf("Failed to get backends endpoints: %v", err)
	}
	log.Printf("Response: %v", resp)

	return resp, err
}

func (c *Client) StartInstance(instanceId string) (*http.Response, error) {
	resp, err := http.Post("http://127.0.0.1:49750/instanceStatus/start/"+instanceId, "application/json", nil)
	if err != nil {
		log.Printf("Failed start instance:"+instanceId+" %v", err)
	}
	log.Printf("Response: %v", resp)

	return resp, err
}
