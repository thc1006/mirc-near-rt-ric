// pkg/e2/asn1.go
package e2

import (
	"encoding/asn1"
	"fmt"
)

// E2AP-PDU is a placeholder for the actual E2AP PDU structure
type E2AP_PDU struct {
	InitiatingMessage InitiatingMessage `asn1:"choice:initiatingMessage"`
}

type InitiatingMessage struct {
	ProcedureCode int64       `asn1:"explicit,tag:0"`
	Criticality   int64       `asn1:"explicit,tag:1"`
	Value         interface{} `asn1:"explicit,tag:2"`
}

// EncodeE2AP_PDU encodes an E2AP_PDU structure into ASN.1 DER format
func EncodeE2AP_PDU(pdu E2AP_PDU) ([]byte, error) {
	encoded, err := asn1.Marshal(pdu)
	if err != nil {
		return nil, fmt.Errorf("failed to encode E2AP-PDU: %w", err)
	}
	return encoded, nil
}

// DecodeE2AP_PDU decodes an ASN.1 DER byte slice into an E2AP_PDU structure
func DecodeE2AP_PDU(data []byte) (*E2AP_PDU, error) {
	var pdu E2AP_PDU
	if _, err := asn1.Unmarshal(data, &pdu); err != nil {
		return nil, fmt.Errorf("failed to decode E2AP-PDU: %w", err)
	}
	return &pdu, nil
}
