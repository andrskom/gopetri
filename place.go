package gopetri

// Place of chip in net.
type Place struct {
	ID             string
	ToTransitions  []*Transition
	FromTransition *Transition
	IsFinished     bool
}

// NewPlace create place.
func NewPlace(id string, isFinished bool) *Place {
	return &Place{
		ID:            id,
		ToTransitions: make([]*Transition, 0),
		IsFinished:    isFinished,
	}
}

// AddToTransitions to place in net.
func (s *Place) AddToTransitions(transition *Transition) error {
	for _, tr := range s.ToTransitions {
		if tr.ID == transition.ID {
			return NewErrorf(
				ErrCodeToTransitionAlreadyRegistered,
				"To transition with ID '%s' already registered for place '%s'",
				transition.ID,
				s.ID,
			)
		}
	}
	s.ToTransitions = append(s.ToTransitions, transition)
	return nil
}

// SetFromTransition to place in net.
func (s *Place) SetFromTransition(transition *Transition) error {
	if s.FromTransition != nil {
		return NewErrorf(
			ErrCodeFromTransitionAlreadyRegistered,
			"From transition already registered for place '%s'",
			s.ID,
		)
	}
	s.FromTransition = transition
	return nil
}
