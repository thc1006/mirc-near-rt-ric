package e2

// E2Interface defines the public interface for the E2 service.
type E2Interface interface {
	Start() error
	Stop() error
	HandleE2Setup(*E2SetupRequest) (*E2SetupResponse, error)
	Subscribe(*SubscriptionRequest) (*SubscriptionResponse, error)
	ProcessIndication(*Indication) error
}
