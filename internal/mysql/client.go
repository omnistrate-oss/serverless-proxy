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
	context context.Context
	*http.Client
}

func NewClient(ctx context.Context) *Client {
	return &Client{ctx, &http.Client{Timeout: 60 * time.Second, Transport: http.DefaultTransport}}
}

/**
 * This API is used to get backend instance status via mapped proxy port.
 * In Omnistrate platform, when creating serverless backend instance, proxy ports will be assigned to the backend instance based on the serverless configuration.
 */
func (c *Client) QueryBackendInstanceStatus(port string) (*http.Response, error) {

	resp, err := http.Get("http://127.0.0.1:49750/instanceStatus/ports/" + port)
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
	resp, err := http.Post("http://127.0.0.1:49750/instanceStatus/start/"+instanceId, "application/json", nil)
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
	resp, err := http.Post("http://127.0.0.1:49750/instanceStatus/stop/"+instanceId, "application/json", nil)
	if err != nil {
		log.Printf("Failed stop instance:"+instanceId+" %v", err)
	}
	log.Printf("Response: %v", resp)

	return resp, err
}