package sidecar

import (
	"github.com/go-openapi/strfmt"
)

type InstanceStatus struct {
	InstanceID            string             `json:"instanceId"`
	ServiceComponents     []ServiceComponent `json:"serviceComponents"`
	Status                Status             `json:"status"`
	LastObservedTimestamp strfmt.DateTime    `json:"lastObservedTimestamp"`
}

type ServiceComponent struct {
	ID             string         `json:"id"`
	Alias          string         `json:"alias"`
	NodesEndpoints []NodeEndpoint `json:"nodesEndpoints"`
}

type NodeEndpoint struct {
	NodeName         string `json:"nodeName"`
	Endpoint         string `json:"endpoint"`
	AvailabilityZone string `json:"availabilityZone"`
}

type Status string

const (
	ACTIVE   Status = "ACTIVE"
	STARTING Status = "STARTING"
	PAUSED   Status = "PAUSED"
	FAILED   Status = "FAILED"
	UNKNOWN  Status = "UNKNOWN"
)
