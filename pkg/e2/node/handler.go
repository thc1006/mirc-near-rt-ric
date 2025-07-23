package node

import (
	"io"
	"net"

	"github.com/hctsai1006/near-rt-ric/pkg/e2/asn1"
	"github.com/hctsai1006/near-rt-ric/pkg/e2/subscription"
	e2apv2 "github.com/onosproject/onos-e2t/api/e2ap/v2"
	e2ap_commondatatypes "github.com/onosproject/onos-e2t/api/e2ap/v2/e2ap-commondatatypes"
	e2appducontents "github.com/onosproject/onos-e2t/api/e2ap/v2/e2ap-pdu-contents"
	e2appdudescriptions "github.com/onosproject/onos-e2t/api/e2ap/v2/e2ap-pdu-descriptions"
	"github.com/sirupsen/logrus"
)

// Handler handles E2 messages.
type Handler struct {
	log                *logrus.Logger
	codec              *asn1.Codec
	subscriptionManager *subscription.Manager
}

// NewHandler creates a new E2 message handler.
func NewHandler(log *logrus.Logger, codec *asn1.Codec, sm *subscription.Manager) *Handler {
	return &Handler{
		log:                log,
		codec:              codec,
		subscriptionManager: sm,
	}
}

// Handle handles an E2 connection.
func (h *Handler) Handle(conn net.Conn) {
	h.log.Infof("Handling new E2 connection from %s", conn.RemoteAddr())
	defer conn.Close()

	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				h.log.WithError(err).Error("Failed to read from connection")
			}
			return
		}

		pdu, err := h.codec.Decode(buf[:n])
		if err != nil {
			h.log.WithError(err).Error("Failed to decode message")
			continue
		}

		switch pdu.GetInitiatingMessage().GetProcedureCode().GetE2ApProtocolIes().GetE2ApProtocolIe29().GetValue() {
		case e2apv2.ProcedureCodeIDE2setup:
			h.handleE2SetupRequest(conn, pdu.GetInitiatingMessage().GetValue().GetE2SetupRequest())
		case e2apv2.ProcedureCodeIDRicSubscription:
			h.handleRicSubscriptionRequest(conn, pdu.GetInitiatingMessage().GetValue().GetRicSubscriptionRequest())
		default:
			h.log.Warnf("Received unhandled message type: %v", pdu)
		}
	}
}

func (h *Handler) handleE2SetupRequest(conn net.Conn, req *e2appducontents.E2SetupRequest) {
	h.log.Infof("Handling E2 setup request: %v", req)

	// Create an E2 setup response
	resp := &e2appducontents.E2SetupResponse{
		ProtocolIes: &e2appducontents.E2SetupResponseIes{
			E2ApProtocolIes3: &e2appducontents.E2SetupResponseIes_E2ApProtocolIes3{
				Id:          e2apv2.ProtocolIeIDGlobalRicID,
				Criticality: e2ap_commondatatypes.Criticality_CRITICALITY_REJECT,
				Value: &e2ap_commondatatypes.GlobalRicId{
					PLmnIdentity: []byte{0x00, 0x01, 0x02},
					RicId: &e2ap_commondatatypes.RicId{
						Value: 0x01,
					},
				},
			},
		},
	}

	// Create the E2AP PDU
	pdu := &e2appdudescriptions.E2ApPdu{
		E2ApPdu: &e2appdudescriptions.E2ApPdu_SuccessfulOutcome{
			SuccessfulOutcome: &e2appdudescriptions.SuccessfulOutcome{
				ProcedureCode: &e2appdudescriptions.ProcedureCode{
					E2ApProtocolIes: &e2appdudescriptions.E2ApProtocolIes{
						E2ApProtocolIe29: &e2appdudescriptions.E2ApProtocolIes_E2ApProtocolIe29{
							Id:          e2apv2.ProtocolIeIDInitiatingMessage,
							Criticality: e2ap_commondatatypes.Criticality_CRITICALITY_IGNORE,
							Value:       e2apv2.ProcedureCodeIDE2setup,
						},
					},
				},
				Value: &e2appdudescriptions.SuccessfulOutcomeE2ApPdu{
					SuccessfulOutcomeE2ApPdu: &e2appdudescriptions.SuccessfulOutcomeE2ApPdu_E2Setup{
						E2Setup: resp,
					},
				},
			},
		},
	}

	// Encode the PDU
	bytes, err := h.codec.Encode(pdu)
	if err != nil {
		h.log.WithError(err).Error("Failed to encode E2 setup response")
		return
	}

	// Send the response
	_, err = conn.Write(bytes)
	if err != nil {
		h.log.WithError(err).Error("Failed to send E2 setup response")
	}
}

func (h *Handler) handleRicSubscriptionRequest(conn net.Conn, req *e2appducontents.RicsubscriptionRequest) {
	h.log.Infof("Handling RIC subscription request: %v", req)
	h.subscriptionManager.Add(req)

	// Create a RIC subscription response
	resp := &e2appducontents.RicsubscriptionResponse{
		ProtocolIes: &e2appducontents.RicsubscriptionResponseIes{
			E2ApProtocolIes5: &e2appducontents.RicsubscriptionResponseIes_E2ApProtocolIes5{
				Id:          e2apv2.ProtocolIeIDRicrequestID,
				Criticality: e2ap_commondatatypes.Criticality_CRITICALITY_REJECT,
				Value:       req.ProtocolIes.E2ApProtocolIes5.Value,
			},
		},
	}

	// Create the E2AP PDU
	pdu := &e2appdudescriptions.E2ApPdu{
		E2ApPdu: &e2appdudescriptions.E2ApPdu_SuccessfulOutcome{
			SuccessfulOutcome: &e2appdudescriptions.SuccessfulOutcome{
				ProcedureCode: &e2appdudescriptions.ProcedureCode{
					E2ApProtocolIes: &e2appdudescriptions.E2ApProtocolIes{
						E2ApProtocolIe29: &e2appdudescriptions.E2ApProtocolIes_E2ApProtocolIe29{
							Id:          e2apv2.ProtocolIeIDInitiatingMessage,
							Criticality: e2ap_commondatatypes.Criticality_CRITICALITY_IGNORE,
							Value:       e2apv2.ProcedureCodeIDRicSubscription,
						},
					},
				},
				Value: &e2appdudescriptions.SuccessfulOutcomeE2ApPdu{
					SuccessfulOutcomeE2ApPdu: &e2appdudescriptions.SuccessfulOutcomeE2ApPdu_RicSubscription{
						RicSubscription: resp,
					},
				},
			},
		},
	}

	// Encode the PDU
	bytes, err := h.codec.Encode(pdu)
	if err != nil {
		h.log.WithError(err).Error("Failed to encode RIC subscription response")
		return
	}

	// Send the response
	_, err = conn.Write(bytes)
	if err != nil {
		h.log.WithError(err).Error("Failed to send RIC subscription response")
	}
}
