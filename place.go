package gopetri

type Place struct {
	ID             string
	ToTransitions  []*Transition
	FromTransition *Transition
	IsFinished     bool
}

func (s *Place) AddToTransitions(transition *Transition) error {
	s.ToTransitions = append(s.ToTransitions, transition)
	return nil
}

func (s *Place) SetFromTransition(transition *Transition) error {
	s.FromTransition = transition
	return nil
}
