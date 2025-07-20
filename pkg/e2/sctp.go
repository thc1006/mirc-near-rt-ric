// pkg/e2/sctp.go
package e2

import (
	"fmt"
	"net"

	"github.com/ishidawataru/sctp"
)

// SCTPManager handles SCTP connections
type SCTPManager struct {
	conn *sctp.SCTPConn
}

// NewSCTPManager creates a new SCTPManager
func NewSCTPManager() *SCTPManager {
	return &SCTPManager{}
}

// Connect establishes an SCTP connection to the given address
func (m *SCTPManager) Connect(address string) error {
	addr, err := net.ResolveIPAddr("ip", address)
	if err != nil {
		return fmt.Errorf("failed to resolve IP address: %w", err)
	}

	sctpAddr := &sctp.SCTPAddr{
		IPAddrs: []net.IPAddr{*addr},
		Port:    36421,
	}

	conn, err := sctp.DialSCTP("sctp", nil, sctpAddr)
	if err != nil {
		return fmt.Errorf("failed to dial SCTP: %w", err)
	}
	m.conn = conn
	return nil
}

// Send sends data over the SCTP connection
func (m *SCTPManager) Send(data []byte) error {
	if m.conn == nil {
		return fmt.Errorf("SCTP connection is not established")
	}
	_, err := m.conn.Write(data)
	return err
}

// Close closes the SCTP connection
func (m *SCTPManager) Close() error {
	if m.conn != nil {
		return m.conn.Close()
	}
	return nil
}
