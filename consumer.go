package gopetri

// Consumer is interface for external consumer.
type Consumer interface {
	BeforePlace(placeID string) error
	AfterPlace(placeID string)
	CanTransit(transitionID string) bool
	BeforeTransit(transitionID string) error
	AfterTransit(transitionID string)
}
