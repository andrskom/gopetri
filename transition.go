package gopetri

// Transition model of net.
type Transition struct {
	ID         string
	ToPlaces   []*Place
	FromPlaces []*Place
}
