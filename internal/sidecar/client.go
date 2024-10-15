package sidecar

import (
	"context"
	"log"
	"net/http"
	"time"
)

/**
 * This file contains all APIs used to interact with omnistrate platform via proxy companion process - sidecar.
 */
type Client struct {
	context    context.Context
	httpClient *http.Client
}

func NewClient(ctx context.Context) *Client {
	return &Client{ctx, &http.Client{Timeout: 60 * time.Second, Transport: http.DefaultTransport}}
}

/**
 * This API is used to get backend instance status via mapped proxy port.
 * In Omnistrate platform, when creating serverless backend instance, proxy ports will be assigned to the backend instance based on the serverless configuration.
 */
func (c *Client) QueryBackendInstanceStatus(port string) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(c.context, http.MethodGet, "http://127.0.0.1:49750/instanceStatus/ports/"+port, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return nil, err
	}
	resp, err = c.httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to get backends endpoints: %v", err)
	}
	log.Printf("Response: %v", resp)

	return resp, err
}

/**
 * This API is used to start backend instance via instance id, proxy will need to obtain instance id first before calling this API
 */
func (c *Client) StartInstance(instanceId string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(c.context, http.MethodPost, "http://127.0.0.1:49750/instanceStatus/start/"+instanceId, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("Failed start instance:"+instanceId+" %v", err)
	}
	log.Printf("Response: %v", resp)

	return resp, err
}

/**
 * This API is used to stop backend instance via instance id, proxy will need to obtain instance id first before calling this API.
 * Note that you may not need this API if you enable auto pause in serverless configuration.
 */
func (c *Client) StopInstance(instanceId string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(c.context, http.MethodPost, "http://127.0.0.1:49750/instanceStatus/stop/"+instanceId, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("Failed stop instance:"+instanceId+" %v", err)
	}
	log.Printf("Response: %v", resp)

	return resp, err
}
