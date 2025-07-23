package asn1

import (
	e2ap "github.com/onosproject/onos-e2t/api/e2ap/v2/e2ap-pdu-descriptions"
	"github.com/onosproject/onos-lib-go/pkg/asn1/aper"
	"github.com/sirupsen/logrus"
)

// Codec handles ASN.1 encoding and decoding.
type Codec struct {
	log *logrus.Logger
}

// NewCodec creates a new ASN.1 codec.
func NewCodec(log *logrus.Logger) *Codec {
	return &Codec{
		log: log,
	}
}

// Decode decodes an E2 message.
func (c *Codec) Decode(data []byte) (*e2ap.E2ApPdu, error) {
	c.log.Info("Decoding E2 message")

	pdu := &e2ap.E2ApPdu{}
	err := aper.Unmarshal(data, pdu)
	if err != nil {
		return nil, err
	}

	return pdu, nil
}

// Encode encodes an E2 message.
func (c *Codec) Encode(pdu *e2ap.E2ApPdu) ([]byte, error) {
	c.log.Info("Encoding E2 message")

	data, err := aper.Marshal(pdu)
	if err != nil {
		return nil, err
	}

	return data, nil
}
