package sidecar

import "time"

type BackendsStatus struct {
	Backends              []BackendStatus `json:"backends"`
	LastObservedTimestamp time.Time       `json:"lastObservedTimestamp"`
}

type BackendStatus struct {
	InstanceName   string         `json:"instanceName"`
	Status         Status         `json:"status"`
	NodesEndpoints []NodeEndpoint `json:"nodesEndpoints"`
}

type NodeEndpoint struct {
	NodeName         string `json:"nodeName"`
	Endpoint         string `json:"endpoint"`
	AvailabilityZone string `json:"availabilityZone"`
}

type Status string

const (
	ACTIVE  Status = "ACTIVE"
	WAKEUP  Status = "WAKEUP"
	PAUSED  Status = "PAUSED"
	FAILED  Status = "FAILED"
	UNKNOWN Status = "UNKNOWN"
)
