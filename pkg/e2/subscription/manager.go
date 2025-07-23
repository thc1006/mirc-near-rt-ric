package subscription

import (
	"sync"

	e2appducontents "github.com/onosproject/onos-e2t/api/e2ap/v2/e2ap-pdu-contents"
	"github.com/sirupsen/logrus"
)

// Manager manages E2 subscriptions.
type Manager struct {
	log           *logrus.Logger
	subscriptions map[string]*e2appducontents.RicsubscriptionRequest
	mu            sync.RWMutex
}

// NewManager creates a new subscription manager.
func NewManager(log *logrus.Logger) *Manager {
	return &Manager{
		log:           log,
		subscriptions: make(map[string]*e2appducontents.RicsubscriptionRequest),
	}
}

// Add adds a new subscription.
func (m *Manager) Add(req *e2appducontents.RicsubscriptionRequest) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Keying the subscription on the requestor ID and RIC instance ID
	key := string(req.ProtocolIes.E2ApProtocolIes5.Value.RicRequestorId.Value) + "-" + string(req.ProtocolIes.E2ApProtocolIes5.Value.RicInstanceId.Value)
	m.subscriptions[key] = req
	m.log.Infof("Added new subscription with key %s", key)
}
