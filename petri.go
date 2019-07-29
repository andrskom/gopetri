package gopetri

import (
	"fmt"

	"github.com/awalterschulze/gographviz"
)

// Net of petri net.
type Net struct {
	start              *Place
	placeRegistry      map[string]*Place
	transitionRegistry map[string]*Transition
	state              State
	consumer           Consumer
}

// State is current state of net, contains places for chips in places and transitions.
type State struct {
	Finished        bool           `json:"finished"`
	Err             *Error         `json:"err"`
	PlaceChips      map[string]int `json:"place_chips"`
	TransitionChips map[string]int `json:"transition_chips"`
}

// IsFinished state.
func (s State) IsFinished() bool {
	return s.Finished
}

// IsError state.
func (s State) IsError() bool {
	return s.Err != nil
}

// New net instance with consumer.
func New() *Net {
	return &Net{
		placeRegistry:      make(map[string]*Place),
		transitionRegistry: make(map[string]*Transition),
	}
}

// GetState return current state of net.
func (c *Net) GetState() State {
	return c.state
}

// StartPlace start net.
//
// !!! Not supported V of nodes and more than one chip.
func (c *Net) StartPlace( /*chipNumber int*/ ) error {
	chipNumber := 1
	c.state = State{
		PlaceChips:      make(map[string]int),
		TransitionChips: make(map[string]int),
	}
	for chipNumber > 0 {
		chipNumber--
		if err := c.SetPlace(c.start.ID); err != nil {
			return err
		}
	}
	return nil
}

// UpFromState set current state for work.
func (c *Net) UpFromState(state State) {
	c.state = state
}

// IsErrState check state on error.
func (c *Net) IsErrState() bool {
	return c.GetState().IsError()
}

// IsFinished return info about state, finished or not.
func (c *Net) IsFinished() bool {
	return c.GetState().IsFinished()
}

// GetErrFromState return err state reason.
func (c *Net) GetErrFromState() *Error {
	return c.GetState().Err
}

// FullReset state of net for reusing.
func (c *Net) FullReset() {
	c.state = State{}
	c.consumer = nil
}

// SetPlace for chip. If chip in finish place, will return true as first result.
func (c *Net) SetPlace(placeID string) error {
	if c.consumer == nil {
		return NewError(ErrCodeConsumerNotSet, "Consumer is not set")
	}
	if c.IsErrState() {
		return NewError(ErrCodeNetInErrState, "Net in err state")
	}
	if c.IsFinished() {
		return NewError(ErrCodeFinished, "Net is finished")
	}

	newPlace, err := c.GetPlace(placeID)
	if err != nil {
		return err
	}
	// validation
	if newPlace != c.start {
		if c.state.TransitionChips[newPlace.FromTransition.ID] == 0 {
			return NewErrorf(
				ErrCodeHasNotChipForNewPlace,
				"has not chip in transition '%s' for place '%s'", newPlace.FromTransition.ID, newPlace.ID,
			)
		}
	}

	if err := c.setPlace(newPlace); err != nil {
		c.state.Err = err
		return NewError(ErrCodeNetInErrState, "Net in err state")
	}
	return nil
}

func (c *Net) setPlace(newPlace *Place) *Error {
	if err := c.consumer.BeforePlace(newPlace.ID); err != nil {
		return NewError(ErrCodeBeforePlaceReturnedErr, err.Error())
	}

	if newPlace != c.start {
		c.state.TransitionChips[newPlace.FromTransition.ID]--
	}

	// set position
	c.state.PlaceChips[newPlace.ID]++
	c.consumer.AfterPlace(newPlace.ID)

	if newPlace.IsFinished {
		c.state.Finished = true
		return nil
	}

	// start transition
	canTransit := 0
	var nextTr *Transition
	for _, tr := range newPlace.ToTransitions {
		if c.consumer.CanTransit(tr.ID) {
			canTransit++
			nextTr = tr
		}
	}
	if canTransit != 1 {
		return NewErrorf(
			ErrCodeUnexpectedAvailableTransitionsNumber,
			"unexpected available transitions number, expected 1, actual %d", canTransit,
		)
	}

	hasUnavailable := false
	for _, place := range nextTr.FromPlaces {
		if c.state.PlaceChips[place.ID] == 0 {
			hasUnavailable = true
		}
	}
	if hasUnavailable {
		return nil
	}

	if err := c.consumer.BeforeTransit(nextTr.ID); err != nil {
		return NewError(ErrCodeBeforeTransitReturnedErr, err.Error())
	}

	for _, place := range nextTr.FromPlaces {
		c.state.PlaceChips[place.ID]--
	}

	c.state.TransitionChips[nextTr.ID] += len(nextTr.ToPlaces)
	c.consumer.AfterTransit(nextTr.ID)

	return nil
}

// SetConsumer for net.
func (c *Net) SetConsumer(l Consumer) {
	c.consumer = l
}

// AddPlace to petri net.
func (c *Net) AddPlace(place *Place) error {
	if _, ok := c.placeRegistry[place.ID]; ok {
		return NewErrorf(ErrCodePlaceAlreadyRegistered, "place with id '%c' already registered", place.ID)
	}
	c.placeRegistry[place.ID] = place

	return nil
}

// GetPlace return place by id.
func (c *Net) GetPlace(placeID string) (*Place, error) {
	place, ok := c.placeRegistry[placeID]
	if !ok {
		return nil, NewErrorf(
			ErrCodePlaceIsNotRegistered,
			"can't get place '%c', place is not registered", placeID,
		)
	}
	return place, nil
}

// AddTransition tp petri net.
func (c *Net) AddTransition(transition *Transition) error {
	if _, ok := c.placeRegistry[transition.ID]; ok {
		return NewErrorf(
			ErrCodeTransitionAlreadyRegistered,
			"transition with id '%c' already registered", transition.ID,
		)
	}
	c.transitionRegistry[transition.ID] = transition

	return nil
}

// SetStartPlace by name of state.
func (c *Net) SetStartPlace(place string) error {
	start, ok := c.placeRegistry[place]
	if !ok {
		return NewErrorf(ErrCodePlaceIsNotRegistered, "place '%c' is not registered")
	}
	c.start = start

	return nil
}

// SetStartPlace by name of state.
func (c *Net) SetErrorState(err *Error) error {
	if c.IsErrState() {
		return NewErrorf(ErrCodeCantSetErrState, "Already set err state")
	}

	if c.IsFinished() {
		return NewErrorf(ErrCodeCantSetErrState, "Net is finished")
	}

	return nil
}

// BuildFromCfg return net by Cfg.
func BuildFromCfg(cfg Cfg) (*Net, error) {
	comp := New()

	finishRegistry := make(map[string]struct{})
	for _, fPlace := range cfg.Finish {
		finishRegistry[fPlace] = struct{}{}
	}

	for _, place := range cfg.Places {
		_, ok := finishRegistry[place]
		if err := comp.AddPlace(NewPlace(place, ok)); err != nil {
			return nil, err
		}
	}

	if err := comp.SetStartPlace(cfg.Start); err != nil {
		return nil, err
	}

	for trID, tr := range cfg.Transitions {
		trModel := &Transition{ID: trID, ToPlaces: make([]*Place, 0, len(tr.To))}
		for _, dst := range tr.To {
			dstPlace, err := comp.GetPlace(dst)
			if err != nil {
				return nil, err
			}
			if err := dstPlace.SetFromTransition(trModel); err != nil {
				return nil, err
			}
			trModel.ToPlaces = append(trModel.ToPlaces, dstPlace)
		}
		if err := comp.AddTransition(trModel); err != nil {
			return nil, err
		}

		for _, src := range tr.From {
			srcPlace, err := comp.GetPlace(src)
			if err != nil {
				return nil, err
			}

			if err := srcPlace.AddToTransitions(trModel); err != nil {
				return nil, err
			}
			trModel.FromPlaces = append(trModel.FromPlaces, srcPlace)
		}
	}

	return comp, nil
}

// AsGraphvizDotLang return string with graphviz dot lang view of graph.
// If u set 'withState', you will have chips on graph.
// 'turquoise' is color for start place, 'sienna' is coor for finish places. 'red' is for chips on map.
func (c *Net) AsGraphvizDotLang(name string, withState bool) (string, error) {
	graphAst, _ := gographviz.ParseString(fmt.Sprintf(`digraph %s {}`, name))
	graph := gographviz.NewGraph()
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		return "", err
	}

	var (
		startPlaceAttr  = map[string]string{string(gographviz.Shape): "oval", string(gographviz.Color): "turquoise"}
		finishPlaceAttr = map[string]string{string(gographviz.Shape): "oval", string(gographviz.Color): "sienna"}
		placeAttr       = map[string]string{string(gographviz.Shape): "oval"}
		transitionAttr  = map[string]string{string(gographviz.Shape): "box"}
	)

	for _, place := range c.placeRegistry {
		attr := mapCopy(placeAttr)
		if place == c.start {
			attr = mapCopy(startPlaceAttr)
		}
		if place.IsFinished {
			attr = mapCopy(finishPlaceAttr)
		}
		if chips, ok := c.state.PlaceChips[place.ID]; ok && chips > 0 {
			attr[string(gographviz.Color)] = "red"
		}
		if err := graph.AddNode(name, place.ID, attr); err != nil {
			return "", err
		}
		if place.IsFinished {
			continue
		}

		for _, tr := range place.ToTransitions {
			if err := graph.AddEdge(place.ID, tr.ID, true, nil); err != nil {
				return "", err
			}
		}
	}

	for _, transition := range c.transitionRegistry {
		attr := mapCopy(transitionAttr)
		if chips, ok := c.state.TransitionChips[transition.ID]; ok && chips > 0 {
			attr[string(gographviz.Color)] = "red"
		}
		if err := graph.AddNode(name, transition.ID, attr); err != nil {
			return "", err
		}
		for _, dst := range transition.ToPlaces {
			if err := graph.AddEdge(transition.ID, dst.ID, true, nil); err != nil {
				return "", err
			}
		}
	}

	return graph.String(), nil
}

func mapCopy(input map[string]string) map[string]string {
	output := make(map[string]string)
	for k, v := range input {
		output[k] = v
	}
	return output
}
