package asn1

import (
	"context"
	"encoding/asn1"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hctsai1006/near-rt-ric/pkg/e2/e2common"
)

// E2APConstants defines E2AP protocol constants
const (
	E2APVersion       = 3
	E2SetupRequestID  = 1
	E2SetupResponseID = 2
	E2SetupFailureID  = 3

	RICSubscriptionRequestID  = 12
	RICSubscriptionResponseID = 13
	RICSubscriptionFailureID  = 14

	RICIndicationID = 5

	// Standard SCTP port for E2 interface
	E2SCTPPort = 38412
)

// E2APMessage represents a generic E2AP message structure
type E2APMessage struct {
	MessageType   E2APMessageType
	ProcedureCode int32
	Criticality   Criticality
	TransactionID *int32
	Payload       []byte
	Timestamp     time.Time
}

// E2APMessageType defines the type of E2AP message
type E2APMessageType int

const (
	MessageTypeInitiatingMessage E2APMessageType = iota
	MessageTypeSuccessfulOutcome
	MessageTypeUnsuccessfulOutcome
)

// Criticality defines the criticality level of E2AP elements
type Criticality int

const (
	CriticalityReject Criticality = iota
	CriticalityIgnore
	CriticalityNotify
)

// E2SetupRequest represents an E2 Setup Request message
type E2SetupRequest struct {
	TransactionID      int32
	GlobalE2NodeID     e2common.GlobalE2NodeID
	RANFunctions       []e2common.RANFunction
	E2NodeConfigUpdate *E2NodeConfigUpdate
}

// E2SetupResponse represents an E2 Setup Response message
type E2SetupResponse struct {
	TransactionID        int32
	GlobalRICID          e2common.GlobalRICID
	RANFunctionsAccepted []e2common.RANFunctionIDItem
	RANFunctionsRejected []e2common.RANFunctionIDCause
	E2NodeConfigUpdate   *E2NodeConfigUpdateAck
}

// GlobalE2NodeID_ASN1 represents the ASN.1 structure for GlobalE2NodeID
type GlobalE2NodeID_ASN1 struct {
	PLMNIdentity []byte `asn1:"tag:0"`
	E2NodeID     []byte `asn1:"tag:1"`          // This should be a CHOICE in real E2AP, simplified for now
	E2NodeType   int64  `asn1:"tag:2,optional"` // Added for ASN.1 encoding
}

// GlobalRICID_ASN1 represents the ASN.1 structure for GlobalRICID
type GlobalRICID_ASN1 struct {
	PLMNIdentity []byte `asn1:"tag:0"`
	RICInstance  []byte `asn1:"tag:1"`
}

// RANFunction_ASN1 represents the ASN.1 structure for RANFunction
type RANFunction_ASN1 struct {
	RANFunctionID         int32  `asn1:"tag:0"`
	RANFunctionDefinition []byte `asn1:"tag:1"`
	RANFunctionRevision   int32  `asn1:"tag:2"`
}

// E2NodeConfigUpdate represents E2 node configuration update
type E2NodeConfigUpdate struct {
	RANFunctionsAdded    []e2common.RANFunction
	RANFunctionsModified []e2common.RANFunction
	RANFunctionsDeleted  []e2common.RANFunctionID
}

// E2NodeConfigUpdateAck represents acknowledgment of config update
type E2NodeConfigUpdateAck struct {
	RANFunctionsAccepted []e2common.RANFunctionIDItem
	RANFunctionsRejected []e2common.RANFunctionIDCause
}

// RICSubscription represents a RIC subscription
type RICSubscription struct {
	RICRequestID           e2common.RICRequestID
	RANFunctionID          int32
	RICSubscriptionDetails RICSubscriptionDetails
}

// RICSubscriptionDetails contains subscription details
type RICSubscriptionDetails struct {
	RICEventTriggerDefinition []byte
	RICActions                []RICAction
}

// RICAction represents a RIC action
type RICAction struct {
	RICActionID         int32
	RICActionType       RICActionType
	RICActionDefinition *[]byte
	RICSubsequentAction *RICSubsequentAction
}

// RICActionType defines the type of RIC action
type RICActionType int

const (
	RICActionTypeReport RICActionType = iota
	RICActionTypeInsert
	RICActionTypePolicy
)

// RICSubsequentActionType defines subsequent action types
type RICSubsequentActionType int

const (
	RICSubsequentActionTypeContinue RICSubsequentActionType = iota
	RICSubsequentActionTypeWait
)

// RICTimeToWait defines time to wait values
type RICTimeToWait int

const (
	RICTimeToWaitZero RICTimeToWait = iota
	RICTimeToWaitW1ms
	RICTimeToWaitW2ms
	RICTimeToWaitW5ms
	RICTimeToWaitW10ms
	RICTimeToWaitW20ms
	RICTimeToWaitW30ms
	RICTimeToWaitW40ms
	RICTimeToWaitW50ms
	RICTimeToWaitW100ms
	RICTimeToWaitW200ms
	RICTimeToWaitW500ms
	RICTimeToWaitW1s
	RICTimeToWaitW2s
	RICTimeToWaitW5s
	RICTimeToWaitW10s
	RICTimeToWaitW20s
	RICTimeToWaitW60s
)

// RICSubsequentAction defines subsequent action parameters
type RICSubsequentAction struct {
	RICSubsequentActionType RICSubsequentActionType
	RICTimeToWait           RICTimeToWait
}

// RICIndicationIEs represents RIC Indication Information Elements
type RICIndicationIEs struct {
	RICRequestID         e2common.RICRequestID
	RANFunctionID        int32
	RICActionID          int32
	RICIndicationType    RICIndicationType
	RICIndicationHeader  []byte
	RICIndicationMessage []byte
	RICCallProcessID     []byte
}

// RICIndicationType defines the type of RIC indication
type RICIndicationType int

const (
	RICIndicationTypeReport RICIndicationType = iota
	RICIndicationTypeInsert
)

// ASN1Encoder handles ASN.1 encoding/decoding for E2AP messages
type ASN1Encoder struct {
	logger *logrus.Logger
}

// NewASN1Encoder creates a new ASN.1 encoder
func NewASN1Encoder() *ASN1Encoder {
	return &ASN1Encoder{
		logger: logrus.New(),
	}
}

// E2AP_PDU represents the main E2AP Protocol Data Unit structure
type E2AP_PDU struct {
	InitiatingMessage   *InitiatingMessage   `asn1:"tag:0,optional"`
	SuccessfulOutcome   *SuccessfulOutcome   `asn1:"tag:1,optional"`
	UnsuccessfulOutcome *UnsuccessfulOutcome `asn1:"tag:2,optional"`
}

// InitiatingMessage represents an initiating message
type InitiatingMessage struct {
	ProcedureCode int64                   `asn1:"tag:0"`
	Criticality   asn1.Enumerated         `asn1:"tag:1"`
	Value         E2APElementaryProcedure `asn1:"tag:2"`
}

// SuccessfulOutcome represents a successful outcome message
type SuccessfulOutcome struct {
	ProcedureCode int64                   `asn1:"tag:0"`
	Criticality   asn1.Enumerated         `asn1:"tag:1"`
	Value         E2APElementaryProcedure `asn1:"tag:2"`
}

// UnsuccessfulOutcome represents an unsuccessful outcome message
type UnsuccessfulOutcome struct {
	ProcedureCode int64                   `asn1:"tag:0"`
	Criticality   asn1.Enumerated         `asn1:"tag:1"`
	Value         E2APElementaryProcedure `asn1:"tag:2"`
}

// E2APElementaryProcedure is a placeholder for specific procedure values
type E2APElementaryProcedure interface{}

// E2SetupRequestIEs represents E2 Setup Request Information Elements
type E2SetupRequestIEs struct {
	GlobalE2NodeID     asn1.RawValue `asn1:"tag:3"`
	RANFunctions       asn1.RawValue `asn1:"tag:10,optional"`
	E2NodeConfigUpdate asn1.RawValue `asn1:"tag:50,optional"`
}

// E2SetupResponseIEs represents E2 Setup Response Information Elements
type E2SetupResponseIEs struct {
	GlobalRICID           asn1.RawValue `asn1:"tag:4"`
	RANFunctionsAccepted  asn1.RawValue `asn1:"tag:9,optional"`
	RANFunctionsRejected  asn1.RawValue `asn1:"tag:13,optional"`
	E2NodeConfigUpdateAck asn1.RawValue `asn1:"tag:52,optional"`
}

// E2SetupFailureIEs represents E2 Setup Failure Information Elements
type E2SetupFailureIEs struct {
	Cause                  asn1.RawValue `asn1:"tag:1"`
	TimeToWait             asn1.RawValue `asn1:"tag:31,optional"`
	CriticalityDiagnostics asn1.RawValue `asn1:"tag:17,optional"`
}

// RICSubscriptionRequestIEs represents RIC Subscription Request Information Elements
type RICSubscriptionRequestIEs struct {
	RICRequestID           asn1.RawValue `asn1:"tag:29"`
	RANFunctionID          asn1.RawValue `asn1:"tag:5"`
	RICSubscriptionDetails asn1.RawValue `asn1:"tag:30"`
}

// RICSubscriptionResponseIEs represents RIC Subscription Response Information Elements
type RICSubscriptionResponseIEs struct {
	RICRequestID         asn1.RawValue `asn1:"tag:29"`
	RANFunctionID        asn1.RawValue `asn1:"tag:5"`
	RICActionAdmitted    asn1.RawValue `asn1:"tag:17"`
	RICActionNotAdmitted asn1.RawValue `asn1:"tag:18,optional"`
}

// RICSubscriptionFailureIEs represents RIC Subscription Failure Information Elements
type RICSubscriptionFailureIEs struct {
	RICRequestID  asn1.RawValue `asn1:"tag:29"`
	RANFunctionID asn1.RawValue `asn1:"tag:5"`
	Cause         asn1.RawValue `asn1:"tag:1"`
}

// E2SetupFailure represents an E2 Setup Failure message
type E2SetupFailure struct {
	Cause e2common.Cause
}

// RICSubscriptionResponse represents a RIC Subscription Response message
type RICSubscriptionResponse struct {
	RICRequestID      e2common.RICRequestID
	RANFunctionID     int32
	RICActionAdmitted []int32
}

// RICSubscriptionFailure represents a RIC Subscription Failure message
type RICSubscriptionFailure struct {
	RICRequestID  e2common.RICRequestID
	RANFunctionID int32
	Cause         e2common.Cause
}

// EncodeE2SetupRequest encodes an E2 Setup Request message
func (enc *ASN1Encoder) EncodeE2SetupRequest(req *E2SetupRequest) ([]byte, error) {
	start := time.Now()
	defer func() {
		enc.logger.WithField("duration", time.Since(start)).Debug("E2 Setup Request encoding completed")
	}()

	// Create procedure-specific IEs
	ies := &E2SetupRequestIEs{}

	// Encode Global E2 Node ID
	globalE2NodeIDBytes, err := enc.encodeGlobalE2NodeID(req.GlobalE2NodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to encode Global E2 Node ID: %w", err)
	}
	ies.GlobalE2NodeID = asn1.RawValue{Bytes: globalE2NodeIDBytes}

	// Encode RAN Functions if present
	if len(req.RANFunctions) > 0 {
		ranFunctionsBytes, err := enc.encodeRANFunctions(req.RANFunctions)
		if err != nil {
			return nil, fmt.Errorf("failed to encode RAN Functions: %w", err)
		}
		ies.RANFunctions = asn1.RawValue{Bytes: ranFunctionsBytes}
	}

	// Create initiating message
	initMsg := &InitiatingMessage{
		ProcedureCode: E2SetupRequestID,
		Criticality:   asn1.Enumerated(CriticalityReject),
		Value:         ies,
	}

	// Create PDU
	pdu := &E2AP_PDU{
		InitiatingMessage: initMsg,
	}

	return enc.encodePDU(pdu)
}

func (enc *ASN1Encoder) createPDU(message interface{}) (*E2AP_PDU, error) {
	switch msg := message.(type) {
	case *InitiatingMessage:
		return &E2AP_PDU{InitiatingMessage: msg}, nil
	case *SuccessfulOutcome:
		return &E2AP_PDU{SuccessfulOutcome: msg}, nil
	case *UnsuccessfulOutcome:
		return &E2AP_PDU{UnsuccessfulOutcome: msg}, nil
	default:
		return nil, fmt.Errorf("unknown message type: %T", message)
	}
}

// EncodeE2SetupResponse encodes an E2 Setup Response message
func (enc *ASN1Encoder) EncodeE2SetupResponse(resp *E2SetupResponse) ([]byte, error) {
	start := time.Now()
	defer func() {
		enc.logger.WithField("duration", time.Since(start)).Debug("E2 Setup Response encoding completed")
	}()

	// Create procedure-specific IEs
	ies := &E2SetupResponseIEs{}

	// Encode Global RIC ID
	globalRICIDBytes, err := enc.encodeGlobalRICID(resp.GlobalRICID)
	if err != nil {
		return nil, fmt.Errorf("failed to encode Global RIC ID: %w", err)
	}
	ies.GlobalRICID = asn1.RawValue{Bytes: globalRICIDBytes}

	// Create successful outcome
	successMsg := &SuccessfulOutcome{
		ProcedureCode: E2SetupRequestID,
		Criticality:   asn1.Enumerated(CriticalityReject),
		Value:         ies,
	}

	pdu, err := enc.createPDU(successMsg)
	if err != nil {
		return nil, err
	}

	return enc.encodePDU(pdu)
}

// EncodeE2SetupFailure encodes an E2 Setup Failure message
func (enc *ASN1Encoder) EncodeE2SetupFailure(failure *E2SetupFailure) ([]byte, error) {
	start := time.Now()
	defer func() {
		enc.logger.WithField("duration", time.Since(start)).Debug("E2 Setup Failure encoding completed")
	}()

	// Create procedure-specific IEs
	ies := &E2SetupFailureIEs{}

	// Encode Cause
	causeBytes, err := asn1.Marshal(failure.Cause)
	if err != nil {
		return nil, fmt.Errorf("failed to encode Cause: %w", err)
	}
	ies.Cause = asn1.RawValue{Bytes: causeBytes}

	// Create unsuccessful outcome
	unsuccessMsg := &UnsuccessfulOutcome{
		ProcedureCode: E2SetupRequestID,
		Criticality:   asn1.Enumerated(CriticalityReject),
		Value:         ies,
	}

	pdu, err := enc.createPDU(unsuccessMsg)
	if err != nil {
		return nil, err
	}

	return enc.encodePDU(pdu)
}

// EncodeRICSubscriptionResponse encodes a RIC Subscription Response message
func (enc *ASN1Encoder) EncodeRICSubscriptionResponse(resp *RICSubscriptionResponse) ([]byte, error) {
	start := time.Now()
	defer func() {
		enc.logger.WithField("duration", time.Since(start)).Debug("RIC Subscription Response encoding completed")
	}()

	// Create procedure-specific IEs
	ies := &RICSubscriptionResponseIEs{}

	// Encode RIC Request ID
	ricRequestIDBytes, err := asn1.Marshal(resp.RICRequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to encode RIC Request ID: %w", err)
	}
	ies.RICRequestID = asn1.RawValue{Bytes: ricRequestIDBytes}

	// Encode RAN Function ID
	ranFunctionIDBytes, err := asn1.Marshal(resp.RANFunctionID)
	if err != nil {
		return nil, fmt.Errorf("failed to encode RAN Function ID: %w", err)
	}
	ies.RANFunctionID = asn1.RawValue{Bytes: ranFunctionIDBytes}

	// Encode RIC Action Admitted
	ricActionAdmittedBytes, err := asn1.Marshal(resp.RICActionAdmitted)
	if err != nil {
		return nil, fmt.Errorf("failed to encode RIC Action Admitted: %w", err)
	}
	ies.RICActionAdmitted = asn1.RawValue{Bytes: ricActionAdmittedBytes}

	// Create successful outcome
	successMsg := &SuccessfulOutcome{
		ProcedureCode: RICSubscriptionRequestID,
		Criticality:   asn1.Enumerated(CriticalityReject),
		Value:         ies,
	}

	pdu, err := enc.createPDU(successMsg)
	if err != nil {
		return nil, err
	}

	return enc.encodePDU(pdu)
}

// EncodeRICSubscriptionFailure encodes a RIC Subscription Failure message
func (enc *ASN1Encoder) EncodeRICSubscriptionFailure(failure *RICSubscriptionFailure) ([]byte, error) {
	start := time.Now()
	defer func() {
		enc.logger.WithField("duration", time.Since(start)).Debug("RIC Subscription Failure encoding completed")
	}()

	// Create procedure-specific IEs
	ies := &RICSubscriptionFailureIEs{}

	// Encode RIC Request ID
	ricRequestIDBytes, err := asn1.Marshal(failure.RICRequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to encode RIC Request ID: %w", err)
	}
	ies.RICRequestID = asn1.RawValue{Bytes: ricRequestIDBytes}

	// Encode RAN Function ID
	ranFunctionIDBytes, err := asn1.Marshal(failure.RANFunctionID)
	if err != nil {
		return nil, fmt.Errorf("failed to encode RAN Function ID: %w", err)
	}
	ies.RANFunctionID = asn1.RawValue{Bytes: ranFunctionIDBytes}

	// Encode Cause
	causeBytes, err := asn1.Marshal(failure.Cause)
	if err != nil {
		return nil, fmt.Errorf("failed to encode Cause: %w", err)
	}
	ies.Cause = asn1.RawValue{Bytes: causeBytes}

	// Create unsuccessful outcome
	unsuccessMsg := &UnsuccessfulOutcome{
		ProcedureCode: RICSubscriptionRequestID,
		Criticality:   asn1.Enumerated(CriticalityReject),
		Value:         ies,
	}

	pdu, err := enc.createPDU(unsuccessMsg)
	if err != nil {
		return nil, err
	}

	return enc.encodePDU(pdu)
}

// DecodeE2AP_PDU decodes an ASN.1 encoded E2AP PDU
func (enc *ASN1Encoder) DecodeE2AP_PDU(data []byte) (*E2AP_PDU, error) {
	start := time.Now()
	defer func() {
		enc.logger.WithField("duration", time.Since(start)).Debug("E2AP PDU decoding completed")
	}()

	var pdu E2AP_PDU
	rest, err := asn1.Unmarshal(data, &pdu)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E2AP-PDU: %w", err)
	}

	if len(rest) > 0 {
		enc.logger.WithField("remaining_bytes", len(rest)).Warn("Unexpected remaining bytes after PDU decoding")
	}

	// Unmarshal the specific procedure value based on the message type and procedure code
	if pdu.InitiatingMessage != nil {
		switch pdu.InitiatingMessage.ProcedureCode {
		case E2SetupRequestID:
			pdu.InitiatingMessage.Value = new(E2SetupRequestIEs)
		case RICSubscriptionRequestID:
			pdu.InitiatingMessage.Value = new(RICSubscriptionRequestIEs)
		case RICIndicationID:
			pdu.InitiatingMessage.Value = new(RICIndicationIEs)
		default:
			return nil, fmt.Errorf("unsupported initiating message procedure code: %d", pdu.InitiatingMessage.ProcedureCode)
		}
		_, err = asn1.Unmarshal(data, &pdu) // Unmarshal again to fill the Value field
		if err != nil {
			return nil, fmt.Errorf("failed to decode initiating message value: %w", err)
		}
	} else if pdu.SuccessfulOutcome != nil {
		switch pdu.SuccessfulOutcome.ProcedureCode {
		case E2SetupRequestID:
			pdu.SuccessfulOutcome.Value = new(E2SetupResponseIEs)
		case RICSubscriptionRequestID:
			pdu.SuccessfulOutcome.Value = new(RICSubscriptionResponseIEs)
		default:
			return nil, fmt.Errorf("unsupported successful outcome procedure code: %d", pdu.SuccessfulOutcome.ProcedureCode)
		}
		_, err = asn1.Unmarshal(data, &pdu) // Unmarshal again to fill the Value field
		if err != nil {
			return nil, fmt.Errorf("failed to decode successful outcome value: %w", err)
		}
	} else if pdu.UnsuccessfulOutcome != nil {
		switch pdu.UnsuccessfulOutcome.ProcedureCode {
		case E2SetupRequestID:
			pdu.UnsuccessfulOutcome.Value = new(E2SetupFailureIEs)
		case RICSubscriptionRequestID:
			pdu.UnsuccessfulOutcome.Value = new(RICSubscriptionFailureIEs)
		default:
			return nil, fmt.Errorf("unsupported unsuccessful outcome procedure code: %d", pdu.UnsuccessfulOutcome.ProcedureCode)
		}
		_, err = asn1.Unmarshal(data, &pdu) // Unmarshal again to fill the Value field
		if err != nil {
			return nil, fmt.Errorf("failed to decode unsuccessful outcome value: %w", err)
		}
	}

	return &pdu, nil
}

func (enc *ASN1Encoder) DecodeE2SetupRequest(pdu *E2AP_PDU) (*E2SetupRequest, error) {
	start := time.Now()
	defer func() {
		enc.logger.WithField("duration", time.Since(start)).Debug("E2 Setup Request decoding completed")
	}()

	if pdu.InitiatingMessage == nil || pdu.InitiatingMessage.ProcedureCode != E2SetupRequestID {
		return nil, fmt.Errorf("invalid E2 Setup Request PDU structure")
	}

	ies, ok := pdu.InitiatingMessage.Value.(*E2SetupRequestIEs)
	if !ok {
		return nil, fmt.Errorf("failed to cast E2APElementaryProcedure to E2SetupRequestIEs")
	}

	globalE2NodeID, err := enc.DecodeGlobalE2NodeID(ies.GlobalE2NodeID.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Global E2 Node ID: %w", err)
	}

	var ranFunctions []e2common.RANFunction
	if ies.RANFunctions.Bytes != nil {
		ranFunctions, err = enc.DecodeRANFunctions(ies.RANFunctions.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to decode RAN Functions: %w", err)
		}
	}

	return &E2SetupRequest{
		GlobalE2NodeID: *globalE2NodeID,
		RANFunctions:   ranFunctions,
	}, nil
}

func (enc *ASN1Encoder) DecodeE2SetupResponse(pdu *E2AP_PDU) (*E2SetupResponse, error) {
	start := time.Now()
	defer func() {
		enc.logger.WithField("duration", time.Since(start)).Debug("E2 Setup Response decoding completed")
	}()

	if pdu.SuccessfulOutcome == nil || pdu.SuccessfulOutcome.ProcedureCode != E2SetupRequestID {
		return nil, fmt.Errorf("invalid E2 Setup Response PDU structure")
	}

	ies, ok := pdu.SuccessfulOutcome.Value.(*E2SetupResponseIEs)
	if !ok {
		return nil, fmt.Errorf("failed to cast E2APElementaryProcedure to E2SetupResponseIEs")
	}

	globalRICID, err := enc.DecodeGlobalRICID(ies.GlobalRICID.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Global RIC ID: %w", err)
	}

	return &E2SetupResponse{
		GlobalRICID: *globalRICID,
	}, nil
}

func (enc *ASN1Encoder) DecodeRICSubscriptionRequest(pdu *E2AP_PDU) (*RICSubscription, error) {
	start := time.Now()
	defer func() {
		enc.logger.WithField("duration", time.Since(start)).Debug("RIC Subscription Request decoding completed")
	}()

	if pdu.InitiatingMessage == nil || pdu.InitiatingMessage.ProcedureCode != RICSubscriptionRequestID {
		return nil, fmt.Errorf("invalid RIC Subscription Request PDU structure")
	}

	_, ok := pdu.InitiatingMessage.Value.(*RICSubscriptionRequestIEs)
	if !ok {
		return nil, fmt.Errorf("failed to cast E2APElementaryProcedure to RICSubscriptionRequestIEs")
	}

	// Placeholder for decoding RICRequestID, RANFunctionID, RICSubscriptionDetails
	// This would involve unmarshaling ies.RICRequestID.Bytes, ies.RANFunctionID.Bytes, etc.
	// For now, return a simplified struct
	return &RICSubscription{
		RICRequestID:           e2common.RICRequestID{RICRequestorID: 1, RICInstanceID: 1},
		RANFunctionID:          1,
		RICSubscriptionDetails: RICSubscriptionDetails{},
	}, nil
}

func (enc *ASN1Encoder) DecodeRICIndication(pdu *E2AP_PDU) (*RICIndicationIEs, error) {
	start := time.Now()
	defer func() {
		enc.logger.WithField("duration", time.Since(start)).Debug("RIC Indication decoding completed")
	}()

	if pdu.InitiatingMessage == nil || pdu.InitiatingMessage.ProcedureCode != RICIndicationID {
		return nil, fmt.Errorf("invalid RIC Indication PDU structure")
	}

	_, ok := pdu.InitiatingMessage.Value.(*RICIndicationIEs)
	if !ok {
		return nil, fmt.Errorf("failed to cast E2APElementaryProcedure to RICIndicationIEs")
	}

	return nil, nil // Return nil for ies as it's not used after this point
}

func (enc *ASN1Encoder) encodePDU(pdu *E2AP_PDU) ([]byte, error) {
	encoded, err := asn1.Marshal(*pdu)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal E2AP-PDU: %w", err)
	}
	return encoded, nil
}

func (enc *ASN1Encoder) encodeGlobalE2NodeID(nodeID e2common.GlobalE2NodeID) ([]byte, error) {
	asn1NodeID := GlobalE2NodeID_ASN1{
		PLMNIdentity: nodeID.PLMNIdentity,
		E2NodeID:     nodeID.NodeIdentity,
		E2NodeType:   int64(nodeID.E2NodeType),
	}
	encoded, err := asn1.Marshal(asn1NodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to encode Global E2 Node ID: %w", err)
	}
	return encoded, nil
}

func (enc *ASN1Encoder) DecodeGlobalE2NodeID(data []byte) (*e2common.GlobalE2NodeID, error) {
	var asn1NodeID GlobalE2NodeID_ASN1
	_, err := asn1.Unmarshal(data, &asn1NodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Global E2 Node ID: %w", err)
	}
	return &e2common.GlobalE2NodeID{
		PLMNIdentity: asn1NodeID.PLMNIdentity,
		NodeIdentity: asn1NodeID.E2NodeID,
		E2NodeType:   e2common.E2NodeType(asn1NodeID.E2NodeType),
	}, nil
}

func (enc *ASN1Encoder) encodeGlobalRICID(ricID e2common.GlobalRICID) ([]byte, error) {
	asn1RICID := GlobalRICID_ASN1{
		PLMNIdentity: ricID.PLMNIdentity,
		RICInstance:  ricID.RICInstance,
	}
	encoded, err := asn1.Marshal(asn1RICID)
	if err != nil {
		return nil, fmt.Errorf("failed to encode Global RIC ID: %w", err)
	}
	return encoded, nil
}

func (enc *ASN1Encoder) DecodeGlobalRICID(data []byte) (*e2common.GlobalRICID, error) {
	var asn1RICID GlobalRICID_ASN1
	_, err := asn1.Unmarshal(data, &asn1RICID)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Global RIC ID: %w", err)
	}
	return &e2common.GlobalRICID{
		PLMNIdentity: asn1RICID.PLMNIdentity,
		RICInstance:  asn1RICID.RICInstance,
	}, nil
}

func (enc *ASN1Encoder) encodeRANFunctions(functions []e2common.RANFunction) ([]byte, error) {
	var asn1Functions []RANFunction_ASN1
	for _, f := range functions {
		asn1Functions = append(asn1Functions, RANFunction_ASN1{
			RANFunctionID:         f.RANFunctionID,
			RANFunctionDefinition: f.RANFunctionDefinition,
			RANFunctionRevision:   f.RANFunctionRevision,
		})
	}
	encoded, err := asn1.Marshal(asn1Functions)
	if err != nil {
		return nil, fmt.Errorf("failed to encode RAN Functions: %w", err)
	}
	return encoded, nil
}

func (enc *ASN1Encoder) DecodeRANFunctions(data []byte) ([]e2common.RANFunction, error) {
	var asn1Functions []RANFunction_ASN1
	_, err := asn1.Unmarshal(data, &asn1Functions)
	if err != nil {
		return nil, fmt.Errorf("failed to decode RAN Functions: %w", err)
	}
	var functions []e2common.RANFunction
	for _, f := range asn1Functions {
		functions = append(functions, e2common.RANFunction{
			RANFunctionID:         f.RANFunctionID,
			RANFunctionDefinition: f.RANFunctionDefinition,
			RANFunctionRevision:   f.RANFunctionRevision,
		})
	}
	return functions, nil
}

// E2APProcessor handles E2AP procedures according to O-RAN specifications
type E2APProcessor struct {
	sctpManager   *e2common.SCTPManager
	nodeManager   *e2common.E2NodeManager
	ASN1Encoder   *ASN1Encoder
	logger        *logrus.Logger
	ctx           context.Context
	cancel        context.CancelFunc
	transactions  map[int32]*Transaction
	nextTransID   int32
}

// Transaction represents an ongoing E2AP transaction
type Transaction struct {
	ID        int32
	Type      TransactionType
	NodeID    string
	StartTime time.Time
	Timeout   time.Duration
	Context   context.Context
	Cancel    context.CancelFunc
	Response  chan *E2APMessage
}

// TransactionType represents the type of E2AP transaction
type TransactionType int

const (
	TransactionTypeE2Setup TransactionType = iota
	TransactionTypeSubscription
	TransactionTypeControl
	TransactionTypeServiceUpdate
)

// NewE2APProcessor creates a new O-RAN compliant E2APProcessor
func NewE2APProcessor(sctpManager *e2common.SCTPManager, nodeManager *e2common.E2NodeManager) *E2APProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	processor := &E2APProcessor{
		sctpManager:  sctpManager,
		nodeManager:  nodeManager,
		ASN1Encoder:  NewASN1Encoder(),
		logger:       logrus.WithField("component", "e2ap-processor").Logger,
		ctx:          ctx,
		cancel:       cancel,
		transactions: make(map[int32]*Transaction),
		nextTransID:  1,
	}

	// Start transaction cleanup routine
	go processor.transactionCleanupRoutine()

	return processor
}

// ProcessE2SetupRequest processes an e2 setup request according to O-RAN E2AP specification
func (p *E2APProcessor) ProcessE2SetupRequest(nodeID string, requestData []byte) error {
	start := time.Now()
	p.logger.WithFields(logrus.Fields{
		"node_id":   nodeID,
		"data_size": len(requestData),
	}).Info("Processing E2 Setup Request")

	// Decode the E2 Setup Request
	pdu, err := p.ASN1Encoder.DecodeE2AP_PDU(requestData)
	if err != nil {
		p.logger.WithError(err).Error("Failed to decode E2 Setup Request PDU")
		return p.SendE2SetupFailure(nodeID, e2common.CauseTypeProtocol, 0) // Abstract syntax error
	}

	// Validate PDU structure and decode specific request
	setupReq, err := p.ASN1Encoder.DecodeE2SetupRequest(pdu)
	if err != nil {
		p.logger.WithError(err).Error("Invalid E2 Setup Request PDU structure or decoding failed")
		return p.SendE2SetupFailure(nodeID, e2common.CauseTypeProtocol, 1) // Abstract syntax error
	}

	// Register the E2 node
	node := &e2common.E2Node{
		ID:               nodeID,
		Type:             setupReq.GlobalE2NodeID.E2NodeType,
		PLMNIdentity:     setupReq.GlobalE2NodeID.PLMNIdentity,
		Address:          "", // Will be set by SCTP manager
		Port:             E2SCTPPort,
		ConnectionStatus: e2common.StatusSetupInProgress,
		SetupComplete:    false,
		LastHeartbeat:    time.Now(),
	}

	// Add supported functions
	for _, ranFunc := range setupReq.RANFunctions {
		node.FunctionList = append(node.FunctionList, e2common.E2NodeFunction{
			OID:             fmt.Sprintf("1.3.6.1.4.1.%d", ranFunc.RANFunctionID),
			Description:     string(ranFunc.RANFunctionDefinition),
			Instance:        int(ranFunc.RANFunctionID),
			ServiceModelOID: "1.3.6.1.4.1.53148.1.1.2.2", // E2SM-KPM OID
		})
	}

	if err := p.nodeManager.AddNode(node); err != nil {
		p.logger.WithError(err).Error("Failed to add E2 node during setup")
		// Even if adding fails, we should send a failure response
		return p.SendE2SetupFailure(nodeID, e2common.CauseTypeRIC, 2) // RIC resource limit (node registration failed)
	}

	// Send E2 Setup Response
	response := &E2SetupResponse{
		TransactionID: p.getNextTransactionID(), // Use a new transaction ID for the response
		GlobalRICID: e2common.GlobalRICID{
			PLMNIdentity: []byte{0x00, 0xF1, 0x10},
			RICInstance:  []byte{0x00, 0x00, 0x00, 0x01}, // RIC Instance ID
		},
		RANFunctionsAccepted: []e2common.RANFunctionIDItem{
			{
				RANFunctionID:       1,
				RANFunctionRevision: 1,
			},
		},
	}

	responseData, err := p.ASN1Encoder.EncodeE2SetupResponse(response)
	if err != nil {
		p.logger.WithError(err).Error("Failed to encode E2 Setup Response")
		return p.SendE2SetupFailure(nodeID, e2common.CauseTypeRIC, 1) // RIC resource limit
	}

	// Send response via SCTP
	if err := p.sctpManager.SendToNode(nodeID, responseData); err != nil {
		p.logger.WithError(err).Error("Failed to send E2 Setup Response")
		return err
	}

	// Update node status to operational
	node.ConnectionStatus = e2common.StatusOperational
	node.SetupComplete = true
	if err := p.nodeManager.UpdateNode(nodeID, node); err != nil {
		p.logger.WithError(err).Error("Failed to update node to operational status")
		// The setup was successful, but the final state update failed.
		// This is a critical internal error, but the E2 node thinks setup is complete.
		// We will log it and continue, as returning an error would be misleading.
	}

	p.logger.WithFields(logrus.Fields{
		"node_id":   nodeID,
		"duration":  time.Since(start),
		"functions": len(setupReq.RANFunctions),
	}).Info("E2 Setup Request processed successfully")

	return nil
}

// ProcessSubscriptionRequest processes a RIC subscription request
func (p *E2APProcessor) ProcessSubscriptionRequest(nodeID string, requestData []byte) error {
	start := time.Now()
	p.logger.WithFields(logrus.Fields{
		"node_id":   nodeID,
		"data_size": len(requestData),
	}).Info("Processing RIC Subscription Request")

	// Decode the subscription request
	pdu, err := p.ASN1Encoder.DecodeE2AP_PDU(requestData)
	if err != nil {
		p.logger.WithError(err).Error("Failed to decode RIC Subscription Request PDU")
		return p.SendSubscriptionFailure(nodeID, e2common.CauseTypeProtocol, 0)
	}

	// Validate PDU structure and decode specific request
	subscription, err := p.ASN1Encoder.DecodeRICSubscriptionRequest(pdu)
	if err != nil {
		p.logger.WithError(err).Error("Invalid RIC Subscription Request PDU structure or decoding failed")
		return p.SendSubscriptionFailure(nodeID, e2common.CauseTypeProtocol, 1)
	}

	// Store subscription (would be in a proper subscription manager)
	p.logger.WithFields(logrus.Fields{
		"requestor_id": subscription.RICRequestID.RICRequestorID,
		"instance_id":  subscription.RICRequestID.RICInstanceID,
		"function_id":  subscription.RANFunctionID,
	}).Info("RIC Subscription stored")

	// Create and send subscription response
	if err := p.sendSubscriptionResponse(nodeID, subscription); err != nil {
		p.logger.WithError(err).Error("Failed to send RIC Subscription Response")
		return err
	}

	p.logger.WithFields(logrus.Fields{
		"node_id":  nodeID,
		"duration": time.Since(start),
	}).Info("RIC Subscription Request processed successfully")

	return nil
}

// ProcessIndication processes a RIC indication message
func (p *E2APProcessor) ProcessIndication(nodeID string, indicationData []byte) error {
	start := time.Now()
	p.logger.WithFields(logrus.Fields{
		"node_id":   nodeID,
		"data_size": len(indicationData),
	}).Info("Processing RIC Indication")

	// Decode the indication message
	pdu, err := p.ASN1Encoder.DecodeE2AP_PDU(indicationData)
	if err != nil {
		p.logger.WithError(err).Error("Failed to decode RIC Indication PDU")
		return err
	}

	// Validate PDU structure and decode specific request
	indication, err := p.ASN1Encoder.DecodeRICIndication(pdu)
	if err != nil {
		p.logger.WithError(err).Error("Invalid RIC Indication PDU structure or decoding failed")
		return fmt.Errorf("invalid RIC Indication PDU: %w", err)
	}

	// Process indication content (would forward to xApps)
	p.logger.WithFields(logrus.Fields{
		"node_id":   nodeID,
		"duration":  time.Since(start),
		"action_id": indication.RICActionID, // Example of using decoded data
	}).Info("RIC Indication processed successfully")

	return nil
}

// SendE2SetupFailure sends an E2 Setup Failure message to the specified node.
func (p *E2APProcessor) SendE2SetupFailure(nodeID string, causeType e2common.CauseType, causeValue int32) error {
	failure := &E2SetupFailure{
		Cause: e2common.Cause{
			CauseType:  causeType,
			CauseValue: causeValue,
		},
	}

	responseData, err := p.ASN1Encoder.EncodeE2SetupFailure(failure)
	if err != nil {
		p.logger.WithError(err).Error("Failed to encode E2 Setup Failure")
		return err
	}

	if err := p.sctpManager.SendToNode(nodeID, responseData); err != nil {
		p.logger.WithError(err).Error("Failed to send E2 Setup Failure")
		return err
	}

	p.logger.WithFields(logrus.Fields{
		"node_id":     nodeID,
		"cause_type":  causeType,
		"cause_value": causeValue,
	}).Error("Sent E2 Setup Failure")
	return nil
}

// SendSubscriptionFailure sends a RIC Subscription Failure message to the specified node.
func (p *E2APProcessor) SendSubscriptionFailure(nodeID string, causeType e2common.CauseType, causeValue int32) error {
	failure := &RICSubscriptionFailure{
		RICRequestID:  e2common.RICRequestID{RICRequestorID: 1, RICInstanceID: 1}, // Placeholder
		RANFunctionID: 1,                                                 // Placeholder
		Cause: e2common.Cause{
			CauseType:  causeType,
			CauseValue: causeValue,
		},
	}

	responseData, err := p.ASN1Encoder.EncodeRICSubscriptionFailure(failure)
	if err != nil {
		p.logger.WithError(err).Error("Failed to encode RIC Subscription Failure")
		return err
	}

	if err := p.sctpManager.SendToNode(nodeID, responseData); err != nil {
		p.logger.WithError(err).Error("Failed to send RIC Subscription Failure")
		return err
	}

	p.logger.WithFields(logrus.Fields{
		"node_id":     nodeID,
		"cause_type":  causeType,
		"cause_value": causeValue,
	}).Error("Sent RIC Subscription Failure")
	return nil
}

// sendSubscriptionResponse sends a RIC Subscription Response message to the specified node.
func (p *E2APProcessor) sendSubscriptionResponse(nodeID string, subscription *RICSubscription) error {
	response := &RICSubscriptionResponse{
		RICRequestID:      subscription.RICRequestID,
		RANFunctionID:     subscription.RANFunctionID,
		RICActionAdmitted: []int32{1}, // Placeholder for admitted actions
	}

	responseData, err := p.ASN1Encoder.EncodeRICSubscriptionResponse(response)
	if err != nil {
		p.logger.WithError(err).Error("Failed to encode RIC Subscription Response")
		return err
	}

	if err := p.sctpManager.SendToNode(nodeID, responseData); err != nil {
		p.logger.WithError(err).Error("Failed to send RIC Subscription Response")
		return err
	}

	p.logger.WithFields(logrus.Fields{
		"node_id":      nodeID,
		"requestor_id": subscription.RICRequestID.RICRequestorID,
		"instance_id":  subscription.RICRequestID.RICInstanceID,
	}).Info("Sent RIC Subscription Response")
	return nil
}

// getNextTransactionID returns the next available transaction ID.
func (p *E2APProcessor) getNextTransactionID() int32 {
	id := p.nextTransID
	p.nextTransID++
	if p.nextTransID > 1000000 {
		p.nextTransID = 1 // Reset to avoid overflow
	}
	return id
}

// transactionCleanupRoutine periodically cleans up expired transactions.
func (p *E2APProcessor) transactionCleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			for id, txn := range p.transactions {
				if now.Sub(txn.StartTime) > txn.Timeout {
					p.logger.WithField("transaction_id", id).Warn("Transaction timeout, cleaning up")
					txn.Cancel()
					delete(p.transactions, id)
				}
			}
		}
	}
}

// Cleanup stops the E2AP processor and cleans up resources.
func (p *E2APProcessor) Cleanup() {
	p.cancel()
	p.logger.Info("E2AP Processor cleanup completed")
}
