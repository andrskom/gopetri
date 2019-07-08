package gopetri

import (
	"encoding/json"
	"fmt"

	"github.com/awalterschulze/gographviz"
)

// Component of petri net.
type Component struct {
	start              *Place
	placeRegistry      map[string]*Place
	transitionRegistry map[string]*Transition
	state              State
	consumer           Consumer
	errState           *Error
}

// State is current state of net, contains places for chips in places and transitions.
type State struct {
	PlaceChips      map[string]int
	TransitionChips map[string]int
}

// New net instance with consumer.
func New(consumer Consumer) *Component {
	return &Component{
		placeRegistry:      make(map[string]*Place),
		transitionRegistry: make(map[string]*Transition),
		consumer:           consumer,
	}
}

// GetState return current state of net.
func (c *Component) GetState() State {
	return c.state
}

// StartPlace start net.
//
// !!! Not supported V of nodes and more than one chip.
func (c *Component) StartPlace( /*chipNumber int*/ ) error {
	chipNumber := 1
	c.state = State{
		PlaceChips:      make(map[string]int),
		TransitionChips: make(map[string]int),
	}
	for chipNumber > 0 {
		chipNumber--
		if _, err := c.SetPlace(c.start.ID); err != nil {
			return err
		}
	}
	return nil
}

// UpFromState set current state for work.
func (c *Component) UpFromState(state State) {
	c.state = state
}

// IsErrState check state on error.
func (c *Component) IsErrState() bool {
	return c.errState != nil
}

// GetErrFromState return err state reason.
func (c *Component) GetErrFromState() *Error {
	return c.errState
}

// SetPlace for chip. If chip in finish place, will return true as first result.
func (c *Component) SetPlace(placeID string) (bool, error) {
	if c.IsErrState() {
		return true, NewError(ErrCodeNetInErrState, "Net in err state")
	}
	newPlace, err := c.GetPlace(placeID)
	if err != nil {
		return false, err
	}
	// validation
	if newPlace != c.start {
		if c.state.TransitionChips[newPlace.FromTransition.ID] == 0 {
			return false, NewErrorf(
				ErrCodeHasNotChipForNewPlace,
				"has not chip in transition '%s' for place '%s'", newPlace.FromTransition.ID, newPlace.ID,
			)
		}
	}

	if err := c.setPlace(newPlace); err != nil {
		c.errState = err
		return true, NewError(ErrCodeNetInErrState, "Net in err state")
	}
	return newPlace.IsFinished, nil
}

func (c *Component) setPlace(newPlace *Place) *Error {
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
func (c *Component) SetConsumer(l Consumer) {
	c.consumer = l
}

// AddPlace to petri net.
func (c *Component) AddPlace(place *Place) error {
	if _, ok := c.placeRegistry[place.ID]; ok {
		return NewErrorf(ErrCodePlaceAlreadyRegistered, "place with id '%c' already registered", place.ID)
	}
	c.placeRegistry[place.ID] = place

	return nil
}

// GetPlace return place by id.
func (c *Component) GetPlace(placeID string) (*Place, error) {
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
func (c *Component) AddTransition(transition *Transition) error {
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
func (c *Component) SetStartPlace(place string) error {
	start, ok := c.placeRegistry[place]
	if !ok {
		return NewErrorf(ErrCodePlaceIsNotRegistered, "place '%c' is not registered")
	}
	c.start = start

	return nil
}

func BuildFromJSONString(jsonBytes []byte, consumer Consumer) (*Component, error) {
	var cfg Cfg
	if err := json.Unmarshal(jsonBytes, &cfg); err != nil {
		return nil, err
	}

	return BuildFromCfg(cfg, consumer)
}

func BuildFromCfg(cfg Cfg, consumer Consumer) (*Component, error) {
	comp := New(consumer)

	finishRegistry := make(map[string]struct{})
	for _, fPlace := range cfg.Finish {
		finishRegistry[fPlace] = struct{}{}
	}

	for _, place := range cfg.Places {
		_, ok := finishRegistry[place]
		if err := comp.AddPlace(&Place{ID: place, IsFinished: ok, ToTransitions: make([]*Transition, 0)}); err != nil {
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
func (c *Component) AsGraphvizDotLang(name string, withState bool) (string, error) {
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
