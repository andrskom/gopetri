package gopetri

type Transition struct {
	ID         string
	ToPlaces   []*Place
	FromPlaces []*Place
}
