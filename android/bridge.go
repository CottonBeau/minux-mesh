package android

import "fmt"

// NodeStatus holds basic node information
type NodeStatus struct {
	DeviceID string
	Address  string
	Running  bool
}

// Start starts the minux node
func Start() *NodeStatus {
	return &NodeStatus{
		DeviceID: "pending",
		Address:  "pending",
		Running:  true,
	}
}

// Ping tests if the bridge is working
func Ping() string {
	return fmt.Sprintf("minux bridge alive")
}
