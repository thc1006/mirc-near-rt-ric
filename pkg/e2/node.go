// pkg/e2/node.go
package e2

import "fmt"

// E2Node represents an E2 node
type E2Node struct {
	ID      string
	Address string
}

// E2NodeManager manages the lifecycle of E2 nodes
type E2NodeManager struct {
	nodes map[string]E2Node
}

// NewE2NodeManager creates a new E2NodeManager
func NewE2NodeManager() *E2NodeManager {
	return &E2NodeManager{nodes: make(map[string]E2Node)}
}

// AddNode adds a new E2 node
func (m *E2NodeManager) AddNode(node E2Node) {
	m.nodes[node.ID] = node
	fmt.Printf("Added E2 node: %s\n", node.ID)
}

// RemoveNode removes an E2 node
func (m *E2NodeManager) RemoveNode(nodeID string) {
	delete(m.nodes, nodeID)
	fmt.Printf("Removed E2 node: %s\n", nodeID)
}

